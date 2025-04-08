package av

import (
	"cmp"
	"slices"
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
