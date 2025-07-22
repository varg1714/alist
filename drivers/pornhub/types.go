package pornhub

const (
	PlayList = iota
	ACTOR
)

type videoInfo struct {
	DefaultQuality bool   `json:"defaultQuality"`
	Format         string `json:"format"`
	VideoURL       string `json:"videoUrl"`
	Quality        string `json:"quality"`
}

type PornFilm struct {
	Image    string
	Title    string
	ViewKey  string
	Username string
}

type MakeDirParam struct {
	Name string `json:"name"`
	Type int    `json:"type"`
	Url  string `json:"url"`
}
