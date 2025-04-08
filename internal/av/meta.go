package av

import "time"

type Meta struct {
	Magnets []Magnet
	Actors  []Actor
}

type Magnet interface {
	GetMagnet() string
	GetName() string
	GetSize() uint64
	IsSubTitle() bool
	GetTags() []string
	GetDownloadCount() uint64
	GetFiles() []File
	GetReleaseDate() time.Time
}

type Actor struct {
	Id   string
	Name string
}

type File struct {
	Name string
	Size uint64
}
