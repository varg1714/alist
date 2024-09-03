package virtual_file

import "encoding/xml"

type Media struct {
	XMLName   xml.Name `xml:"movie"`
	Title     Inner    `xml:"title"`
	SortTitle string   `xml:"sorttitle"`
	Number    string   `xml:"num"`
	Studio    Inner    `xml:"studio"`
	Maker     Inner    `xml:"maker"`
	Director  Inner    `xml:"director"`
	Release   string   `xml:"release"`
	Premiered string   `xml:"premiered"`
	Year      string   `xml:"year"`
	Plot      Inner    `xml:"plot"`
	Outline   Inner    `xml:"outline"`
	RunTime   string   `xml:"runtime"`
	Mpaa      string   `xml:"mpaa"`
	Country   string   `xml:"country"`
	Poster    string   `xml:"poster"`
	Thumb     string   `xml:"thumb"`
	FanArt    string   `xml:"fanart"`
	Actor     []Actor  `xml:"actor"`
	Tag       []Inner  `xml:"tag"`
	Genre     []Inner  `xml:"genre"`
	Set       string   `xml:"set"`
	Label     string   `xml:"label"`
	Cover     string   `xml:"cover"`
	WebSite   string   `xml:"website"`
	Month     string   `xml:"-"`
	DirPath   string   `xml:"-"`
	Source    string   `xml:"-"`
}

// Inner 文字数据，为了避免某些内容被转义。
type Inner struct {
	Inner string `xml:",innerxml"`
}

// Actor 演员信息，保存演员姓名及头像地址。
type Actor struct {
	Name  string `xml:"name"`
	Thumb string `xml:"thumb"`
}

type Option struct {
	MaxPageNum int
	CacheFile  bool
}
