package main

type Comic struct {
	Name     string     `json:"name"`
	UUID     string     `json:"uuid"`
	Cover    string     `json:"cover"`
	PathWord string     `json:"path_word"`
	Author   []PathWord `json:"author"`
	Theme    []PathWord `json:"theme"`
	Brief    string     `json:"brief"`
	Region   Display    `json:"region"`
}

type Display struct {
	Value   int    `json:"value"`
	Display string `json:"display"`
}

type PathWord struct {
	Name     string `json:"name"`
	PathWord string `json:"path_word"`
}

type ChapterInfo struct {
	Index int    `json:"index"`
	UUID  string `json:"uuid"`
	Count int    `json:"count"`
	Size  int    `json:"size"`
	Name  string `json:"name"`
}

type Chapter struct {
	Contents []URL `json:"contents"`
}

type URL struct {
	URL string `json:"url"`
}
