package shiritori

import (
	"strings"

	"golang.org/x/text/unicode/runenames"
)

func isLastLong(str []rune) bool {
	return str[len(str)-1] == 'ー'
}

func trimLong(str []rune) []rune {
	for i := len(str) - 1; i >= 0; i-- {
		if str[i] != 'ー' {
			return str[0 : i+1]
		}
	}
	return str[0:0]
}

func endsWithBannedChar(str []rune) bool {
	return str[len(str)-1] == 'ん'
}

func getLastVowel(str []rune) string {
	runeName := runenames.Name(str[len(str)-1])
	switch runeName[len(runeName)-1] {
	case 'A':
		return "あ"
	case 'I':
		return "い"
	case 'U':
		return "う"
	case 'E':
		return "え"
	case 'O':
		return "お"
	}
	return "お"
}

func isHiraganaString(str []rune) bool {
	for _, r := range str {
		if r == 'ー' {
			continue
		}
		runeName := runenames.Name(r)
		if strings.Index(runeName, "HIRAGANA LETTER ") == 0 {
			continue
		}
		return false
	}
	return true
}

func getLastChar(str []rune) string {
	for i := len(str) - 1; i >= 0; i-- {
		runeName := runenames.Name(str[i])
		if strings.Index(runeName, " SMALL ") == -1 {
			return string(str[i:])
		}
	}
	return string(str)
}

func GetPrefix(word string) string {
	str := []rune(word)
	if len(str) == 1 {
		return ""
	}

	if isLastLong(str) {
		return string(trimLong(str))
	}

	lastChar := getLastChar(str)
	return string(str[:len(str)-len([]rune(lastChar))])
}

func GetSuffix(word string) string {
	str := []rune(word)
	if isLastLong(str) {
		return string(getLastVowel(trimLong(str)))
	}
	return string(getLastChar(str))
}

func IsValidShiritori(prev, cur string) bool {
	p := []rune(prev)
	c := []rune(cur)
	pl := len(p)
	cl := len(c)

	if pl == 0 || cl == 0 {
		return false
	}

	if !isHiraganaString(p) || !isHiraganaString(c) {
		return false
	}

	// "ーー" みたいな入力を落とす
	trimedCur := trimLong(c)
	if len(trimedCur) == 0 {
		return false
	}
	if endsWithBannedChar(trimedCur) {
		return false
	}

	if isLastLong(p) {
		// 長音で終わる場合
		trimed := trimLong(p)
		if len(trimed) == 0 {
			return false
		}

		// 長音で終わる場合、最後に出てくる母音を最後の文字として一致するか確認
		vowel := getLastVowel(trimed)
		if strings.Index(string(c), vowel) != 0 {
			return false
		}
	} else {
		prevLast := getLastChar(p)
		if strings.Index(string(c), prevLast) != 0 {
			return false
		}
	}

	return true
}
