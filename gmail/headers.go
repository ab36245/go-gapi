package gmail

type Headers []Header

func (hs Headers) Content() Headers {
	return hs.filter(func(h Header) bool {
		return h.MatchesNames(contentNames)
	})
}

func (hs Headers) Primary() Headers {
	return hs.filter(func(h Header) bool {
		return h.MatchesNames(primaryNames)
	})
}

func (hs Headers) filter(predicate func(Header) bool) Headers {
	var filtered Headers
	for _, h := range hs {
		if predicate(h) {
			filtered = append(filtered, h)
		}
	}
	return filtered
}

var contentNames = []string{
	"Content-Disposition",
	"Content-Transfer-Encoding",
	"Content-Type",
	"Mime-Version",
}

var primaryNames = []string{
	"Cc",
	"Date",
	"Delivered-To",
	"From",
	"In-Reply-To",
	"List-Unsubscribe",
	"Message-ID",
	"Reply-To",
	"Subject",
	"To",
}
