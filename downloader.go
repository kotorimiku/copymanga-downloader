package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/tidwall/gjson"
)

type BookInfo struct {
	Series      string
	Author      string
	Description string
	Genre       string
	Title       string
	Cover       string
}

type Downloader struct {
	urlBase     string
	pathWord    string
	ChapterList []ChapterInfo
	bookInfo    BookInfo

	config *Config
}

type HeaderRoundTripper struct {
	Headers      map[string]string
	RoundTripper http.RoundTripper
}

func (h *HeaderRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// 为每个请求设置请求头
	for key, value := range h.Headers {
		req.Header.Set(key, value)
	}
	// 执行请求并返回响应
	return h.RoundTripper.RoundTrip(req)
}

var client *http.Client = &http.Client{
	Transport: &HeaderRoundTripper{
		Headers: map[string]string{
			"User-Agent": "COPY/2.2.5",
			"Accept":     "application/json",
			// "Accept-Encoding": "gzip",
			"source":     "copyApp",
			"deviceinfo": "DCO-AL00-DCO-AL00",
			"webp":       "0",
			// "authorization":   "Token",
			"platform": "4",
			"referer":  "com.copymanga.app-2.2.5",
			"version":  "2.2.5",
			"region":   "0",
		},
		RoundTripper: http.DefaultTransport,
	},
}

func NewDownloader(urlBase string, pathWord string, config *Config) *Downloader {
	return &Downloader{
		urlBase:  urlBase,
		pathWord: pathWord,
		config:   config,
	}
}

func Search(urlBase string, keyword string, page int) ([]Comic, error) {
	limit := 12
	url := fmt.Sprintf("https://%s/api/v3/search/comic?offset=%d&platform=4&limit=%d&q=%s&q_type=",
		urlBase, (page-1)*limit, limit, keyword)
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 解析 JSON 响应
	result := gjson.GetBytes(body, "results.list")
	var comics []Comic
	err = json.Unmarshal([]byte(result.Raw), &comics)
	if err != nil {
		return nil, err
	}
	return comics, nil
}

func (d *Downloader) GetImageUrlListUseToken(chapterUUID string, token string) ([]string, error) {
	url := fmt.Sprintf("https://%s/api/v3/comic/%s/chapter2/%s", d.urlBase, d.pathWord, chapterUUID)
	if token != "" {
		client.Transport.(*HeaderRoundTripper).Headers["Authorization"] = fmt.Sprintf("Token %s", token)
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	msg := gjson.GetBytes(body, "message").String()
	if strings.Contains(msg, "Expected available in") {
		return nil, fmt.Errorf("请求超过了限速")
	}

	// 获取图片 URL 列表
	result := gjson.GetBytes(body, "results.chapter.contents")
	var imageUrls []string
	result.ForEach(func(_, value gjson.Result) bool {
		imageUrls = append(imageUrls, value.Get("url").String())
		return true
	})

	if len(imageUrls) == 0 {
		return nil, fmt.Errorf("没有找到图片")
	}

	return imageUrls, nil
}

var mu sync.Mutex

func (d *Downloader) GetImageUrlList(chapterUUID string) ([]string, error) {
	mu.Lock()
	defer mu.Unlock()
	imageUrls, err := d.GetImageUrlListUseToken(chapterUUID, "")
	if err != nil {
		for _, user := range ConfigInstance.UserList {
			imageUrls, err = d.GetImageUrlListUseToken(chapterUUID, user.Token)
			if err == nil {
				break
			}
		}
		if err != nil {
			var user *User = &User{}
			err = Register(user)
			if err != nil {
				return nil, err
			}
			err = Login(user)
			if err != nil {
				return nil, err
			}
			ConfigInstance.UserList = append(ConfigInstance.UserList, user)
			ConfigInstance.SaveConfig(ConfigInstance)
			imageUrls, err = d.GetImageUrlListUseToken(chapterUUID, user.Token)
			if err != nil {
				return nil, err
			}
		}
	}
	return imageUrls, nil
}

func (d *Downloader) GetComicChapter() error {
	url := fmt.Sprintf("https://%s/api/v3/comic/%s/group/default/chapters?limit=500&offset=0", d.urlBase, d.pathWord)
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// 获取章节列表
	result := gjson.GetBytes(body, "results.list")
	var chapters []ChapterInfo
	result.ForEach(func(_, value gjson.Result) bool {
		chapters = append(chapters, ChapterInfo{
			UUID: value.Get("uuid").String(),
			Name: value.Get("name").String(),
		})
		return true
	})
	d.ChapterList = chapters
	return nil
}

func (d *Downloader) GetBookInfo() error {
	url := fmt.Sprintf("https://%s/api/v3/comic2/%s", d.urlBase, d.pathWord)
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// 获取书籍信息
	result := gjson.GetBytes(body, "results.comic")
	var comic Comic
	json.Unmarshal([]byte(result.Raw), &comic)
	d.bookInfo.Series = comic.Name
	var author []string
	for _, res := range comic.Author {
		author = append(author, res.Name)
	}
	var theme []string
	for _, res := range comic.Theme {
		theme = append(theme, res.Name)
	}
	d.bookInfo.Author = strings.Join(author, ", ")
	d.bookInfo.Description = comic.Brief
	d.bookInfo.Genre = strings.Join(theme, ", ")
	d.bookInfo.Cover = comic.Cover
	return nil
}

func (d *Downloader) GetComicInfo() error {
	var wg sync.WaitGroup
	wg.Add(2)

	var bookInfoErr error
	var chapterInfoErr error

	// 获取书籍信息
	go func() {
		defer wg.Done()
		bookInfoErr = d.GetBookInfo()
	}()

	// 获取章节信息
	go func() {
		defer wg.Done()
		chapterInfoErr = d.GetComicChapter()
	}()

	wg.Wait()

	if bookInfoErr != nil {
		return bookInfoErr
	}
	if chapterInfoErr != nil {
		return chapterInfoErr
	}
	return nil
}

func (d *Downloader) DownloadList(chapters []int) error {
	var wg sync.WaitGroup

	maxConcurrency := 2
	sem := make(chan struct{}, maxConcurrency)

	for _, index := range chapters {
		wg.Add(1)
		sem <- struct{}{}
		go func(index int) {
			defer wg.Done()
			err := d.Download(index)
			if err != nil {
				fmt.Println("Error downloading chapter:", err)
			}
			<-sem
		}(index)
	}
	wg.Wait()
	return nil
}

func (d *Downloader) Download(index int) error {
	chapter := d.ChapterList[index]
	imageUrls, err := d.GetImageUrlList(chapter.UUID)
	if err != nil {
		return err
	}

	// 创建文件夹
	folderPath := filepath.Join(d.config.OutputPath, d.bookInfo.Series, sanitizeFilename(chapter.Name))
	os.MkdirAll(folderPath, os.ModePerm)

	var wg sync.WaitGroup
	maxConcurrency := 16
	sem := make(chan struct{}, maxConcurrency)

	for i, url := range imageUrls {
		wg.Add(1)
		sem <- struct{}{}
		go func(i int, url string) {
			defer wg.Done()
			filePath := filepath.Join(folderPath, fmt.Sprintf("%03d.%s", i+1, strings.Split(url, ".")[len(strings.Split(url, "."))-1]))
			err := d.DownloadImage(url, filePath)
			if err != nil {
				fmt.Println("Error downloading image:", err)
			}
			<-sem
		}(i, url)
	}

	wg.Wait()

	if d.config.PackageType == "cbz" {
		comicInfo := ComicInfo{
			Series:    d.bookInfo.Series,
			Writer:    d.bookInfo.Author,
			Summary:   d.bookInfo.Description,
			Genre:     d.bookInfo.Genre,
			Title:     chapter.Name,
			Number:    fmt.Sprintf("%d", index+1),
			PageCount: fmt.Sprintf("%d", len(imageUrls)),
		}
		comicInfo.Build(folderPath)
		zipPath := folderPath + ".cbz"
		err = d.CreateZipFromDirectory(folderPath, zipPath)
		if err != nil {
			return err
		}
		os.RemoveAll(folderPath)
	} else if d.config.PackageType == "zip" {
		zipPath := folderPath + ".zip"
		err = d.CreateZipFromDirectory(folderPath, zipPath)
		if err != nil {
			return err
		}
		os.RemoveAll(folderPath)
	} else if d.config.PackageType == "epub" {
		zipPath := folderPath + ".epub"
		index = index + 1
		MetaData := MetaData{
			Title:       d.bookInfo.Series,
			Creator:     &d.bookInfo.Author,
			Description: &d.bookInfo.Description,
			Subject:     strings.Split(d.bookInfo.Genre, ", "),
			Index:       &index,
			Series:      &d.bookInfo.Series,
		}

		epubBuilder := EpubBuilder{
			metadata: MetaData,
		}
		epubBuilder.BuildComic(zipPath, folderPath)
		os.RemoveAll(folderPath)
	}
	println(folderPath)

	return nil
}

func (d *Downloader) DownloadImage(url, filePath string) error {
	maxRetries := 50

	for i := 0; i < maxRetries; i++ {
		resp, err := client.Get(url)
		if err != nil {
			fmt.Println("Error downloading image:", err)
			time.Sleep(3 * time.Second)
			continue
		}
		defer resp.Body.Close()

		imgData, err := io.ReadAll(resp.Body)
		if err != nil {
			time.Sleep(3 * time.Second)
			continue
		}

		if !isImage(imgData) {
			fmt.Println(string(imgData))
			time.Sleep(3 * time.Second)
			continue
		}

		err = os.WriteFile(filePath, imgData, 0644)
		if err != nil {
			fmt.Println("Error writing image file:", err)
			time.Sleep(3 * time.Second)
			continue
		}

		return nil
	}

	return fmt.Errorf("failed to download image: %s", url)

}

func (d *Downloader) CreateZipFromDirectory(sourceDir, zipPath string) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	err = filepath.Walk(sourceDir, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录本身
		if fi.IsDir() {
			return nil
		}

		// 创建一个文件在压缩包中
		relPath, err := filepath.Rel(sourceDir, file)
		if err != nil {
			return err
		}

		writer, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}

		fileReader, err := os.Open(file)
		if err != nil {
			return err
		}
		defer fileReader.Close()

		_, err = io.Copy(writer, fileReader)
		return err
	})

	return err
}
