package main

// thanks to: https://github.com/01walid/goarabic

// Vowels (Tashkeel) characters.
const (
	FATHA    rune = '\u064e'
	FATHATAN rune = '\u064b'
	DAMMA    rune = '\u064f'
	DAMMATAN rune = '\u064c'
	KASRA    rune = '\u0650'
	KASRATAN rune = '\u064d'
	SHADDA   rune = '\u0651'
	SUKUN    rune = '\u0652'
)

// Arabic Alphabet using the new Harf type.
const (
	ALEF_HAMZA_ABOVE     rune = '\u0623' // أ
	ALEF                 rune = '\u0627' // ا
	ALEF_MADDA_ABOVE     rune = '\u0622' // آ
	HAMZA                rune = '\u0621' // ء
	WAW_HAMZA_ABOVE      rune = '\u0624' // ؤ
	ALEF_HAMZA_BELOW     rune = '\u0625' // أ
	YEH_HAMZA_ABOVE      rune = '\u0626' // ئ
	BEH                  rune = '\u0628' // ب
	TEH                  rune = '\u062A' // ت
	TEH_MARBUTA          rune = '\u0629' // ة
	THEH                 rune = '\u062b' // ث
	JEEM                 rune = '\u062c' // ج
	HAH                  rune = '\u062d' // ح
	KHAH                 rune = '\u062e' // خ
	DAL                  rune = '\u062f' // د
	THAL                 rune = '\u0630' // ذ
	REH                  rune = '\u0631' // ر
	ZAIN                 rune = '\u0632' // ز
	SEEN                 rune = '\u0633' // س
	SHEEN                rune = '\u0634' // ش
	SAD                  rune = '\u0635' // ص
	DAD                  rune = '\u0636' // ض
	TAH                  rune = '\u0637' // ط
	ZAH                  rune = '\u0638' // ظ
	AIN                  rune = '\u0639' // ع
	GHAIN                rune = '\u063a' // غ
	FEH                  rune = '\u0641' // ف
	QAF                  rune = '\u0642' // ق
	KAF                  rune = '\u0643' // ك
	LAM                  rune = '\u0644' // ل
	MEEM                 rune = '\u0645' // م
	NOON                 rune = '\u0646' // ن
	HEH                  rune = '\u0647' // ه
	WAW                  rune = '\u0648' // و
	YEH                  rune = '\u06cc' // ی
	ARABICYEH            rune = '\u064a' // ي
	ALEF_MAKSURA         rune = '\u0649' // ى
	LAM_ALEF             rune = '\ufefb' // لا
	LAM_ALEF_HAMZA_ABOVE rune = '\ufef7' // ﻷ
)

var arabicAphabets = map[rune]struct{}{
	ALEF_HAMZA_ABOVE:     {},
	ALEF:                 {},
	ALEF_MADDA_ABOVE:     {},
	HAMZA:                {},
	WAW_HAMZA_ABOVE:      {},
	ALEF_HAMZA_BELOW:     {},
	YEH_HAMZA_ABOVE:      {},
	BEH:                  {},
	TEH:                  {},
	TEH_MARBUTA:          {},
	THEH:                 {},
	JEEM:                 {},
	HAH:                  {},
	KHAH:                 {},
	DAL:                  {},
	THAL:                 {},
	REH:                  {},
	ZAIN:                 {},
	SEEN:                 {},
	SHEEN:                {},
	SAD:                  {},
	DAD:                  {},
	TAH:                  {},
	ZAH:                  {},
	AIN:                  {},
	GHAIN:                {},
	FEH:                  {},
	QAF:                  {},
	KAF:                  {},
	LAM:                  {},
	MEEM:                 {},
	NOON:                 {},
	HEH:                  {},
	WAW:                  {},
	YEH:                  {},
	ARABICYEH:            {},
	ALEF_MAKSURA:         {},
	LAM_ALEF:             {},
	LAM_ALEF_HAMZA_ABOVE: {},
}

// remove non-arabic everything, except Underscore "_"
func rmNonAr(s string) string {
	r := make([]rune, 0, len(s))
	foundAr := false
	for _, v := range s {
		if _, ok := arabicAphabets[v]; ok {
			r = append(r, v)
			foundAr = true
		} else if v == '_' {
			r = append(r, v)
		}
	}

	if foundAr {
		return string(r)
	}
	return ""
}

// remove harakat
func rmHarakats(s string) string {
	r := make([]rune, 0, len(s))
loop:
	for _, v := range s {
		switch v {
		case FATHA, FATHATAN, DAMMA, DAMMATAN, KASRA, KASRATAN, SHADDA, SUKUN:
			continue loop
		default:
			r = append(r, v)
		}
	}
	return string(r)
}
