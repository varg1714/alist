package av

type Meta struct {
	Magnets []Magnet
	Actors  []Actor
}

type Magnet struct {
	Magnet        string
	Name          string
	Size          uint64
	Subtitle      bool
	Tags          []string
	DownloadCount uint64
	Files         []File
	Source        string
}

type Actor struct {
	Id   string
	Name string
}

type File struct {
	Name string
	Size uint64
}
