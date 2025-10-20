package dict

import (
	"strings"
)

var harakaats = []rune{'a', 'u', 'i', 'F', 'N', 'K', '~', 'o'}

var buck2uni = map[rune]rune{
	'\'': 0x0621, // hamza-on-the-line
	'|':  0x0622, // madda
	'>':  0x0623, // hamza-on-'alif
	'&':  0x0624, // hamza-on-waaw
	'<':  0x0625, // hamza-under-'alif
	'}':  0x0626, // hamza-on-yaa'
	'A':  0x0627, // bare 'alif
	'b':  0x0628, // baa'
	'p':  0x0629, // taa' marbuuTa
	't':  0x062A, // taa'
	'v':  0x062B, // thaa'
	'j':  0x062C, // jiim
	'H':  0x062D, // Haa'
	'x':  0x062E, // khaa'
	'd':  0x062F, // daal
	'*':  0x0630, // dhaal
	'r':  0x0631, // raa'
	'z':  0x0632, // zaay
	's':  0x0633, // siin
	'$':  0x0634, // shiin
	'S':  0x0635, // Saad
	'D':  0x0636, // Daad
	'T':  0x0637, // Taa'
	'Z':  0x0638, // Zaa' (DHaa')
	'E':  0x0639, // cayn
	'g':  0x063A, // ghayn
	// '_':  0x0640, // taTwiil
	'f': 0x0641, // faa'
	'q': 0x0642, // qaaf
	'k': 0x0643, // kaaf
	'l': 0x0644, // laam
	'm': 0x0645, // miim
	'n': 0x0646, // nuun
	'h': 0x0647, // haa'
	'w': 0x0648, // waaw
	'Y': 0x0649, // 'alif maqSuura
	'y': 0x064A, // yaa'
	'F': 0x064B, // fatHatayn
	'N': 0x064C, // Dammatayn
	'K': 0x064D, // kasratayn
	'a': 0x064E, // fatHa
	'u': 0x064F, // Damma
	'i': 0x0650, // kasra
	'~': 0x0651, // shaddah
	'o': 0x0652, // sukuun
	'`': 0x0670, // dagger 'alif
	'{': 0x0671, // waSla
}

var uni2buck = func() map[rune]rune {
	r := map[rune]rune{}
	for k, v := range buck2uni {
		r[v] = k
	}
	return r
}()

// not used
func transliterate(s string) string {
	r := []rune{}
	for _, c := range s {
		v, ok := uni2buck[c]
		if ok {
			r = append(r, v)
		} else {
			r = append(r, c)
		}
	}
	return string(r)
}

func transliterateRmHarakats(s string) []string {
	r := []rune{}
	look := 0
loop:
	for _, c := range s {
		v, ok := uni2buck[c]
		if ok {
			for _, h := range harakaats {
				if h == v {
					continue loop
				}
			}
			r = append(r, v)
		} else if c == 0x6CC {
			// wired: ی this is either
			// a ya or alif maksura
			look++
			r = append(r, c)
		}
		// this just essensially removes all the char other than arabic char
		// else {
		// 	r = append(r, c)
		// }
	}
	str := string(r)
	if look == 0 {
		return []string{str}
	}

	wc := string(rune(0x6CC))
	res := []string{}
	for i := 0; i <= look; i++ {
		t := strings.Replace(str, wc, "Y", i)
		t = strings.ReplaceAll(t, wc, "y")
		res = append(res, t)
	}
	return res
}

func runeReplace(s []rune, r, v rune, t int) {
	rt := 0
	for i := 0; i < len(s); i++ {
		if s[i] == r {
			s[i] = v
			rt++
		}
		if rt == t {
			break
		}
	}
}

// buck to ar
func deTransliterate(s string) string {
	r := []rune{}
	for _, c := range s {
		v, ok := buck2uni[c]
		if ok {
			r = append(r, v)
		} else {
			r = append(r, c)
		}
	}
	return string(r)
}

func ContainsArabic(s string) bool {
	for _, r := range s {
		_, ok := uni2buck[r]
		if ok {
			return true
		}
	}
	return false
}
