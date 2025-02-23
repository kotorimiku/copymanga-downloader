package main

import (
	"context"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

var muD sync.Mutex

type DownloaderManager struct {
	ctx             context.Context
	downloaders     []*DownloaderSingle
	downloadersChan chan *DownloaderSingle
	view            *Downloader
}

func (d *DownloaderManager) startup(ctx context.Context) {
	d.ctx = ctx
	d.downloadersChan = make(chan *DownloaderSingle, 999)
	DownloadList(d.downloadersChan, d.ProcessSend, d.ClearDownloaders)
}

func (d *DownloaderManager) Search(keyword string, page int) ([]Comic, error) {
	return Search(ConfigInstance.UrlBase, keyword, page)
}

func (d *DownloaderManager) GetDownloader(pathWord string) {
	d.view = NewDownloader(ConfigInstance.UrlBase, pathWord, ConfigInstance)
}

func (d *DownloaderManager) GetBookInfo() (BookInfo, error) {
	d.view.GetBookInfo()
	return *d.view.bookInfo, nil
}

func (d *DownloaderManager) GetComicChapter() ([]*ChapterInfo, error) {
	d.view.GetComicChapter()
	return d.view.ChapterList, nil
}

func (d *DownloaderManager) DownloadList(chapters []int) {
	downloaderSingleList := d.view.GetDownloadList(chapters)
	for _, downloaderSingle := range downloaderSingleList {
		d.downloadersChan <- downloaderSingle
	}
	muD.Lock()
	d.downloaders = append(d.downloaders, downloaderSingleList...)
	muD.Unlock()
}

func (d *DownloaderManager) GetDownloaders() []*DownloaderSingle {
	return d.downloaders
}

func (d *DownloaderManager) ClearDownloaders() {
	index := 0
	muD.Lock()
	for i, downloader := range d.downloaders {
		if downloader.Progress < 100 {
			break
		}
		index = i + 1
	}
	d.downloaders = d.downloaders[index:]
	muD.Unlock()
	d.ProcessSend()
}

func (d *DownloaderManager) ProcessSend() {
	runtime.EventsEmit(d.ctx, "progress", d.downloaders)
}
