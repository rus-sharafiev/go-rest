package localization

import (
	"net/http"
	"strings"
)

type Langs struct {
	En string
	Ru string
}

func SelectString(r *http.Request, langs Langs) string {
	preferredLang := strings.Split(r.Header.Get("Accept-Language"), ",")[0]

	if contentLang := r.Header.Get("Content-Language"); len(contentLang) != 0 {
		preferredLang = strings.Split(contentLang, ",")[0]
	}

	isRu := strings.Contains(strings.ToLower(preferredLang), "ru")

	if isRu {
		return langs.Ru
	} else {
		return langs.En
	}
}
