package usage

import (
	"strings"
	"unicode"
)

type Category struct {
	ID          string
	Distracting bool
}

var (
	CategoryWork          = Category{"work", false}
	CategoryProductivity  = Category{"productivity", false}
	CategoryCommunication = Category{"communication", false}
	CategorySocial        = Category{"social", true}
	CategoryVideo         = Category{"video", true}
	CategoryGames         = Category{"games", true}
	CategoryReading       = Category{"reading", false}
	CategoryOther         = Category{"other", false}
)

var knownPackages = map[string]Category{
	"com.google.android.youtube":            CategoryVideo,
	"com.google.android.apps.youtube.music": CategoryVideo,
	"com.instagram.android":                 CategorySocial,
	"com.zhiliaoapp.musically":              CategoryVideo,
	"com.ss.android.ugc.trill":              CategoryVideo,
	"com.ss.android.ugc.aweme":              CategoryVideo,
	"com.facebook.katana":                   CategorySocial,
	"com.twitter.android":                   CategorySocial,
	"com.x.android":                         CategorySocial,
	"com.vkontakte.android":                 CategorySocial,
	"com.reddit.frontpage":                  CategorySocial,
	"com.pinterest":                         CategorySocial,
	"com.snapchat.android":                  CategorySocial,
	"com.whatsapp":                          CategoryCommunication,
	"org.telegram.messenger":                CategoryCommunication,
	"org.thunderdog.challegram":             CategoryCommunication,
	"com.google.android.gm":                 CategoryCommunication,
	"com.discord":                           CategoryCommunication,
	"com.viber.voip":                        CategoryCommunication,
	"com.slack":                             CategoryCommunication,
	"com.microsoft.office.outlook":          CategoryCommunication,
	"com.netflix.mediaclient":               CategoryVideo,
	"com.google.android.videos":             CategoryVideo,
	"tv.twitch.android.app":                 CategoryVideo,
	"com.google.android.apps.docs":          CategoryProductivity,
	"com.android.chrome":                    CategoryOther,
	"com.heytap.browser":                    CategoryOther,
	"com.coloros.gallery3d":                 CategoryOther,
}

var categoryKeywords = []struct {
	fragment string
	category Category
}{
	{"game", CategoryGames},
	{"games", CategoryGames},
	{"video", CategoryVideo},
	{"player", CategoryVideo},
	{"messenger", CategoryCommunication},
	{"messaging", CategoryCommunication},
	{"mail", CategoryCommunication},
	{"reader", CategoryReading},
	{"book", CategoryReading},
	{"news", CategoryReading},
}

func distractingID(id string) bool {
	switch id {
	case CategorySocial.ID, CategoryVideo.ID, CategoryGames.ID:
		return true
	}
	return false
}

func CategoryOf(pkg string) Category {
	if c, ok := knownPackages[pkg]; ok {
		return c
	}
	lower := strings.ToLower(pkg)
	for _, k := range categoryKeywords {
		if strings.Contains(lower, k.fragment) {
			return k.category
		}
	}
	return CategoryOther
}

func LabelOf(pkg string) string {
	parts := strings.Split(pkg, ".")
	name := parts[len(parts)-1]
	if name == "" {
		return pkg
	}
	name = strings.ReplaceAll(name, "_", " ")
	runes := []rune(name)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}
