package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

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
	ChapterList []*ChapterInfo
	bookInfo    *BookInfo

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
			"source":        "copyApp",
			"deviceinfo":    "DCO-AL00-DCO-AL00",
			"webp":          "0",
			"authorization": "Token 38273e31a67233e84436a9088895aed8442d2458",
			"platform":      "4",
			"referer":       "com.copymanga.app-2.2.5",
			"version":       "2.2.5",
			"region":        "0",
		},
		RoundTripper: http.DefaultTransport,
	},
}

func NewDownloader(urlBase string, pathWord string, config *Config) *Downloader {
	return &Downloader{
		urlBase:  urlBase,
		pathWord: pathWord,
		config:   config,
		bookInfo: &BookInfo{},
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
	var chapters []*ChapterInfo
	err = json.Unmarshal([]byte(result.Raw), &chapters)
	if err != nil {
		return err
	}
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

func (d *Downloader) GetDownloadList(chapters []int) []*DownloaderSingle {
	var downloaderSinglesList []*DownloaderSingle = make([]*DownloaderSingle, 0, len(chapters))
	for _, index := range chapters {
		downloaderSingle := DownloaderSingle{
			urlBase:  d.urlBase,
			PathWord: d.pathWord,
			Chapter:  d.ChapterList[index],
			BookInfo: d.bookInfo,
			config:   d.config,
		}
		downloaderSinglesList = append(downloaderSinglesList, &downloaderSingle)
	}
	return downloaderSinglesList
}

func (d *Downloader) DownloadList(chapters []int, processSend func()) error {
	var wg sync.WaitGroup

	maxConcurrency := 2
	sem := make(chan struct{}, maxConcurrency)

	for _, index := range chapters {
		wg.Add(1)
		sem <- struct{}{}
		go func(index int) {
			defer wg.Done()
			downloaderSingle := DownloaderSingle{
				urlBase:  d.urlBase,
				PathWord: d.pathWord,
				Chapter:  d.ChapterList[index],
				BookInfo: d.bookInfo,
				config:   d.config,
			}
			err := downloaderSingle.Download(processSend)
			if err != nil {
				fmt.Println("Error downloading chapter:", err)
			}
			<-sem
		}(index)
	}
	wg.Wait()
	return nil
}

var maxConcurrency = 2
var sem = make(chan struct{}, maxConcurrency)
var isDownloading = false

func DownloadList(downloaderSingleList chan *DownloaderSingle, processSend func(), clearDownloaders func()) {
	// if isDownloading {
	// 	return
	// }
	isDownloading = true
	var wg sync.WaitGroup

	for downloaderSingle := range downloaderSingleList {
		wg.Add(1)
		sem <- struct{}{}
		go func(downloaderSingle *DownloaderSingle) {
			defer wg.Done()
			err := downloaderSingle.Download(processSend)
			if err != nil {
				fmt.Println("Error downloading chapter:", err)
			}
			clearDownloaders()
			<-sem
		}(downloaderSingle)
	}
	wg.Wait()
	isDownloading = false
}
