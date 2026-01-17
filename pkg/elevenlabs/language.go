package elevenlabs

import "strings"

//////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	codeLanguage = map[string]string{
		"afr": "afrikaans",
		"amh": "amharic",
		"ara": "arabic",
		"hye": "armenian",
		"asm": "assamese",
		"ast": "asturian",
		"aze": "azerbaijani",
		"bel": "belarusian",
		"ben": "bengali",
		"bos": "bosnian",
		"bul": "bulgarian",
		"mya": "burmese",
		"yue": "cantonese",
		"cat": "catalan",
		"ceb": "cebuano",
		"nya": "chichewa",
		"hrv": "croatian",
		"ces": "czech",
		"dan": "danish",
		"nld": "dutch",
		"eng": "english",
		"est": "estonian",
		"fil": "filipino",
		"fin": "finnish",
		"fra": "french",
		"ful": "fulah",
		"glg": "galician",
		"lug": "ganda",
		"kat": "georgian",
		"deu": "german",
		"ell": "greek",
		"guj": "gujarati",
		"hau": "hausa",
		"heb": "hebrew",
		"hin": "hindi",
		"hun": "hungarian",
		"isl": "icelandic",
		"ibo": "igbo",
		"ind": "indonesian",
		"gle": "irish",
		"ita": "italian",
		"jpn": "japanese",
		"jav": "javanese",
		"kea": "kabuverdianu",
		"kan": "kannada",
		"kaz": "kazakh",
		"khm": "khmer",
		"kor": "korean",
		"kur": "kurdish",
		"kir": "kyrgyz",
		"lao": "lao",
		"lav": "latvian",
		"lin": "lingala",
		"lit": "lithuanian",
		"luo": "luo",
		"ltz": "luxembourgish",
		"mkd": "macedonian",
		"msa": "malay",
		"mal": "malayalam",
		"mlt": "maltese",
		"cmn": "mandarin chinese",
		"mri": "maori",
		"mar": "marathi",
		"mon": "mongolian",
		"nep": "nepali",
		"nso": "northern sotho",
		"nor": "norwegian",
		"oci": "occitan",
		"ori": "odia",
		"pus": "pashto",
		"fas": "persian",
		"pol": "polish",
		"por": "portuguese",
		"pan": "punjabi",
		"ron": "romanian",
		"rus": "russian",
		"srp": "serbian",
		"sna": "shona",
		"snd": "sindhi",
		"slk": "slovak",
		"slv": "slovenian",
		"som": "somali",
		"spa": "spanish",
		"swa": "swahili",
		"swe": "swedish",
		"tam": "tamil",
		"tgk": "tajik",
		"tel": "telugu",
		"tha": "thai",
		"tur": "turkish",
		"ukr": "ukrainian",
		"umb": "umbundu",
		"urd": "urdu",
		"uzb": "uzbek",
		"vie": "vietnamese",
		"cym": "welsh",
		"wol": "wolof",
		"xho": "xhosa",
		"zul": "zulu",
	}
	languageCode = make(map[string]string, len(codeLanguage))
)

//////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func init() {
	// Initialize the languageCode map from codeLanguage
	for code, language := range codeLanguage {
		languageCode[language] = code
	}
}

//////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// LanguageCode returns the language and three-letter ElevenLabs language
// code for a given tuple, or an empty string if the language
// is not recognized.
func LanguageCode(language string) (string, string) {
	language = strings.ToLower(language)
	if language_, ok := codeLanguage[language]; ok {
		return language_, language
	}
	if code, ok := languageCode[language]; ok {
		return language, code
	}
	return "", ""
}
