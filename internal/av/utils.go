package av

import (
	"cmp"
	"regexp"
	"slices"
	"strings"
)

func sortMagnet(meta *Meta) {

	slices.SortFunc(meta.Magnets, func(a, b Magnet) int {

		if a.IsSubTitle() && !b.IsSubTitle() {
			return -1
		} else if !a.IsSubTitle() && b.IsSubTitle() {
			return 1
		}

		tagCmp := cmp.Compare(len(b.GetTags()), len(a.GetTags()))

		if tagCmp != 0 {
			return tagCmp
		}

		countCmp := cmp.Compare(b.GetDownloadCount(), a.GetDownloadCount())
		if countCmp != 0 {
			return countCmp
		}

		return cmp.Compare(b.GetSize(), a.GetSize())

	})

}

func GetFilmCode(name string) string {
	code := name
	split := strings.Split(name, " ")
	if len(split) >= 2 {
		code = split[0]
	} else {
		nameRegexp, _ := regexp.Compile("(.*?)(-cd\\d+)?.mp4")
		if nameRegexp.MatchString(name) {
			code = nameRegexp.ReplaceAllString(name, "$1")
		}
	}
	return code
}
