package main

type DownloaderManager struct {
	downloaders []*Downloader
	view        *Downloader
}

func (d *DownloaderManager) Search(keyword string, page int) ([]Comic, error) {
	return Search(ConfigInstance.UrlBase, keyword, page)
}

func (d *DownloaderManager) GetDownloader(pathWord string) {
	d.view = NewDownloader(ConfigInstance.UrlBase, pathWord, ConfigInstance)
}

func (d *DownloaderManager) GetBookInfo() (BookInfo, error) {
	d.view.GetBookInfo()
	return d.view.bookInfo, nil
}

func (d *DownloaderManager) GetComicChapter() ([]ChapterInfo, error) {
	d.view.GetComicChapter()
	return d.view.ChapterList, nil
}

func (d *DownloaderManager) DownloadList(chapters []int) {
	d.downloaders = append(d.downloaders, d.view)
	d.view.DownloadList(chapters)
}

func (d *DownloaderManager) GetDownloaders() []*Downloader {
	return d.downloaders
}
