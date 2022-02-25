package shiritori

import (
	"testing"
)

func Test_isLastLong(t *testing.T) {
	testcases := []struct {
		name   string
		input  string
		result bool
	}{
		{
			name:   "not long 1",
			input:  "あ",
			result: false,
		},
		{
			name:   "not long 2",
			input:  "あいう",
			result: false,
		},
		{
			name:   "long 1",
			input:  "あー",
			result: true,
		},
		{
			name:   "long 2",
			input:  "あーー",
			result: true,
		},
	}
	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			result := isLastLong([]rune(testcase.input))
			if result != testcase.result {
				t.Errorf("Unexpected result for %s: expected=%v, actual=%v\n", testcase.input, testcase.result, result)
			}
		})
	}
}

func Test_trimLong(t *testing.T) {
	testcases := []struct {
		name   string
		input  string
		result string
	}{
		{
			name:   "no long",
			input:  "あ",
			result: "あ",
		},
		{
			name:   "has long",
			input:  "あー",
			result: "あ",
		},
		{
			name:   "has multiple long",
			input:  "あーー",
			result: "あ",
		},
		{
			name:   "only long",
			input:  "ーー",
			result: "",
		},
	}
	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			result := trimLong([]rune(testcase.input))
			if string(result) != testcase.result {
				t.Errorf("Unexpected result for %s: expected=%s, actual=%s\n", testcase.input, testcase.result, string(result))
			}
		})
	}
}

func Test_endsWithBannedChar(t *testing.T) {
	testcases := []struct {
		name   string
		input  string
		result bool
	}{
		{
			name:   "case 1",
			input:  "あ",
			result: false,
		},
		{
			name:   "case 2",
			input:  "あん",
			result: true,
		},
	}
	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			result := endsWithBannedChar([]rune(testcase.input))
			if result != testcase.result {
				t.Errorf("Unexpected result for %s: expected=%v, actual=%v\n", testcase.input, testcase.result, result)
			}
		})
	}
}

func Test_getLastVowel(t *testing.T) {
	testcases := []struct {
		name   string
		input  string
		result string
	}{
		{
			name:   "normal",
			input:  "いか",
			result: "あ",
		},
		{
			name:   "dakuon",
			input:  "いが",
			result: "あ",
		},
		{
			name:   "small 1",
			input:  "さぁ",
			result: "あ",
		},
		{
			name:   "small 2",
			input:  "あゎ",
			result: "あ",
		},
		{
			name:   "small 3",
			input:  "いゕ",
			result: "あ",
		},
		{
			name:   "ancient 1",
			input:  "あゐ",
			result: "い",
		},
		{
			name:   "ancient 1",
			input:  "あゑ",
			result: "え",
		},
	}
	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			result := getLastVowel([]rune(testcase.input))
			if string(result) != testcase.result {
				t.Errorf("Unexpected result for %s: expected=%s, actual=%s\n", testcase.input, testcase.result, string(result))
			}
		})
	}
}

func Test_isHiraganaString(t *testing.T) {
	testcases := []struct {
		name   string
		input  string
		result bool
	}{
		{
			name:   "normal",
			input:  "あか",
			result: true,
		},
		{
			name:   "long",
			input:  "あーか",
			result: true,
		},
		{
			name:   "ancient",
			input:  "あゐか",
			result: true,
		},
		{
			name:   "katakana",
			input:  "あカ",
			result: false,
		},
		{
			name:   "symbol",
			input:  "あ・か",
			result: false,
		},
	}
	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			result := isHiraganaString([]rune(testcase.input))
			if result != testcase.result {
				t.Errorf("Unexpected result for %s: expected=%v, actual=%v\n", testcase.input, testcase.result, result)
			}
		})
	}
}

func Test_getLastChar(t *testing.T) {
	testcases := []struct {
		name   string
		input  string
		result string
	}{
		{
			name:   "normal",
			input:  "あか",
			result: "か",
		},
		{
			name:   "small",
			input:  "てぃっしゅ",
			result: "しゅ",
		},
		{
			name:   "dakuon",
			input:  "ぎゃんぐ",
			result: "ぐ",
		},
		{
			name:   "double small",
			input:  "ぐしゃっ",
			result: "しゃっ",
		},
	}
	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			result := getLastChar([]rune(testcase.input))
			if string(result) != testcase.result {
				t.Errorf("Unexpected result for %s: expected=%s, actual=%s\n", testcase.input, testcase.result, result)
			}
		})
	}
}

func Test_IsValidShiritori(t *testing.T) {
	testcases := []struct {
		name   string
		input1 string
		input2 string
		result bool
	}{
		{
			name:   "normal",
			input1: "あか",
			input2: "かき",
			result: true,
		},
		{
			name:   "normal",
			input1: "あか",
			input2: "きく",
			result: false,
		},
		{
			name:   "long",
			input1: "らんかー",
			input2: "あか",
			result: true,
		},
		{
			name:   "long_fail",
			input1: "らんかー",
			input2: "かき",
			result: false,
		},
		{
			name:   "small",
			input1: "てぃっしゅ",
			input2: "しゅっぱつ",
			result: true,
		},
		{
			name:   "small_fail",
			input1: "てぃっしゅ",
			input2: "うし",
			result: false,
		},
		{
			name:   "ancient",
			input1: "あゐ",
			input2: "ゐる",
			result: true,
		},
		{
			name:   "ancient_fail",
			input1: "あゐ",
			input2: "いる",
			result: false,
		},
		{
			name:   "symbol",
			input1: "かなでぃあん・ろっきー",
			input2: "いし",
			result: false,
		},
		{
			name:   "empty",
			input1: "いし",
			input2: "",
			result: false,
		},
		{
			name:   "long only",
			input1: "いし",
			input2: "ーー",
			result: false,
		},
		{
			name:   "banned",
			input1: "かもめ",
			input2: "めん",
			result: false,
		},
		{
			name:   "banned_long",
			input1: "かもめ",
			input2: "めんー",
			result: false,
		},
	}
	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			result := IsValidShiritori(testcase.input1, testcase.input2)
			if result != testcase.result {
				t.Errorf("Unexpected result for %s and %s: expected=%v, actual=%v\n", testcase.input1, testcase.input2, testcase.result, result)
			}
		})
	}
}

func Test_GetSuffix(t *testing.T) {
	testcases := []struct {
		name   string
		input  string
		result string
	}{
		{
			name:   "normal 1",
			input:  "あい",
			result: "い",
		},
		{
			name:   "normal 2",
			input:  "ほげ",
			result: "げ",
		},
		{
			name:   "long",
			input:  "さー",
			result: "あ",
		},
		{
			name:   "long 2",
			input:  "しゃー",
			result: "あ",
		},
		{
			name:   "small 1",
			input:  "しゃ",
			result: "しゃ",
		},
		{
			name:   "small 2",
			input:  "ばぁ",
			result: "ばぁ",
		},
		{
			name:   "one char",
			input:  "あ",
			result: "あ",
		},
		{
			name:   "multi small",
			input:  "しゃっ",
			result: "しゃっ",
		},
	}
	for _, testcase := range testcases {
		result := GetSuffix(testcase.input)
		if result != testcase.result {
			t.Errorf("Unexpected result for %s: expected=%v, actual=%v\n", testcase.input, testcase.result, result)
		}
	}
}

func Test_GetPrefix(t *testing.T) {
	testcases := []struct {
		name   string
		input  string
		result string
	}{
		{
			name:   "normal 1",
			input:  "あい",
			result: "あ",
		},
		{
			name:   "normal 2",
			input:  "ほげ",
			result: "ほ",
		},
		{
			name:   "long",
			input:  "さー",
			result: "さ",
		},
		{
			name:   "long 2",
			input:  "しゃー",
			result: "しゃ",
		},
		{
			name:   "long 2",
			input:  "しゃーー",
			result: "しゃ",
		},
		{
			name:   "small 1",
			input:  "しゃ",
			result: "",
		},
		{
			name:   "small 2",
			input:  "ばぁ",
			result: "",
		},
		{
			name:   "one char",
			input:  "あ",
			result: "",
		},
		{
			name:   "multi small",
			input:  "しゃっ",
			result: "",
		},
	}
	for _, testcase := range testcases {
		result := GetPrefix(testcase.input)
		if result != testcase.result {
			t.Errorf("Unexpected result for %s: expected=%v, actual=%v\n", testcase.input, testcase.result, result)
		}
	}
}
