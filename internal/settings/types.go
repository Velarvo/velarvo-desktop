package settings

type LanguageCode string

const (
	LanguageEnglish LanguageCode = "en"
	LanguageCzech   LanguageCode = "cs"

	DefaultLanguage = LanguageEnglish
)

var SupportedLanguages = []LanguageCode{
	LanguageEnglish,
	LanguageCzech,
}

func (c LanguageCode) Valid() bool {
	for _, supported := range SupportedLanguages {
		if c == supported {
			return true
		}
	}
	return false
}

type Settings struct {
	Language LanguageCode `json:"language"`
}
