package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/tidwall/gjson"
)

type DownloaderSingle struct {
	urlBase  string       `json:"-"`
	PathWord string       `json:"pathWord"`
	Chapter  *ChapterInfo `json:"chapter"`
	BookInfo *BookInfo    `json:"bookInfo"`
	Progress float64      `json:"progress"`
	config   *Config      `json:"-"`
}

func (d *DownloaderSingle) Download(processSend func()) error {
	chapter := d.Chapter
	index := chapter.Index
	imageUrls, err := d.GetImageUrlList(chapter.UUID)
	if err != nil {
		return err
	}

	var folderPath string
	if d.config.NamingStyle == "03d-index-title" {
		folderPath = filepath.Join(d.config.OutputPath, d.BookInfo.Series, fmt.Sprintf("%03d-%s", index, sanitizeFilename(chapter.Name)))
	} else if d.config.NamingStyle == "02d-index-title" {
		folderPath = filepath.Join(d.config.OutputPath, d.BookInfo.Series, fmt.Sprintf("%02d-%s", index, sanitizeFilename(chapter.Name)))
	} else if d.config.NamingStyle == "index-title" {
		folderPath = filepath.Join(d.config.OutputPath, d.BookInfo.Series, fmt.Sprintf("%d-%s", index, sanitizeFilename(chapter.Name)))
	} else {
		folderPath = filepath.Join(d.config.OutputPath, d.BookInfo.Series, sanitizeFilename(chapter.Name))
	}
	os.MkdirAll(folderPath, os.ModePerm)

	var wg sync.WaitGroup
	maxConcurrency := 16
	sem := make(chan struct{}, maxConcurrency)

	total := len(imageUrls)
	downloadedImages := 0

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
			downloadedImages++
			process := float64(downloadedImages) / float64(total) * 100
			d.Progress = process
			processSend()
			<-sem
		}(i, url)
	}

	wg.Wait()

	if d.config.PackageType == "cbz" {
		comicInfo := ComicInfo{
			Series:    d.BookInfo.Series,
			Writer:    d.BookInfo.Author,
			Summary:   d.BookInfo.Description,
			Genre:     d.BookInfo.Genre,
			Title:     chapter.Name,
			Number:    fmt.Sprintf("%d", index+1),
			PageCount: fmt.Sprintf("%d", len(imageUrls)),
		}
		comicInfo.Build(folderPath)
		zipPath := folderPath + ".cbz"
		err = CreateZipFromDirectory(folderPath, zipPath)
		if err != nil {
			return err
		}
		os.RemoveAll(folderPath)
	} else if d.config.PackageType == "zip" {
		zipPath := folderPath + ".zip"
		err = CreateZipFromDirectory(folderPath, zipPath)
		if err != nil {
			return err
		}
		os.RemoveAll(folderPath)
	} else if d.config.PackageType == "epub" {
		zipPath := folderPath + ".epub"
		index = index + 1
		MetaData := MetaData{
			Title:       d.Chapter.Name,
			Creator:     &d.BookInfo.Author,
			Description: &d.BookInfo.Description,
			Subject:     strings.Split(d.BookInfo.Genre, ", "),
			Index:       &index,
			Series:      &d.BookInfo.Series,
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

func (d *DownloaderSingle) DownloadImage(url, filePath string) error {
	maxRetries := 50
	url = strings.Replace(url, ".c800x.", ".c1500x.", 1)

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

var mu sync.Mutex

func (d *DownloaderSingle) GetImageUrlList(chapterUUID string) ([]string, error) {
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

func (d *DownloaderSingle) GetImageUrlListUseToken(chapterUUID string, token string) ([]string, error) {
	url := fmt.Sprintf("https://%s/api/v3/comic/%s/chapter2/%s", d.urlBase, d.PathWord, chapterUUID)
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

func CreateZipFromDirectory(sourceDir, zipPath string) error {
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
