package main

import (
	"bytes"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var translit = map[rune]rune{
	'А': 'A',
	'а': 'a',

	'Б': 'Б',
	'б': '6',

	'В': 'B',
	'в': 'в',

	'Г': 'Г',
	'г': 'r',

	'Д': 'Д',
	'д': 'g',

	'Е': 'E',
	'е': 'e',

	'Ё': 'E',
	'ё': 'e',

	'Ж': 'Ж',
	'ж': 'ж',

	'З': '3',
	'з': 'з',

	'И': 'U',
	'и': 'u',

	'Й': 'Й',
	'й': 'й',

	'К': 'K',
	'к': 'k',

	'Л': 'Л',
	'л': 'л',

	'М': 'M',
	'м': 'm',

	'Н': 'H',
	'н': 'н',

	'О': 'O',
	'о': 'o',

	'П': 'П',
	'п': 'n',

	'Р': 'P',
	'р': 'p',

	'С': 'C',
	'с': 'c',

	'Т': 'T',
	'т': 'т',

	'У': 'Y',
	'у': 'y',

	'Ф': 'Ф',
	'ф': 'ф',

	'Х': 'X',
	'х': 'x',

	'Ц': 'Ц',
	'ц': 'ц',

	'Ч': 'Ч',
	'ч': 'ч',

	'Ш': 'Ш',
	'ш': 'ш',

	'Щ': 'Щ',
	'щ': 'щ',

	'Ъ': 'Ъ',
	'ъ': 'ъ',

	'Ы': 'Ы',
	'ы': 'ы',

	'Ь': 'b',
	'ь': 'ь',

	'Э': 'Э',
	'э': 'э',

	'Ю': 'Ю',
	'ю': 'ю',

	'Я': 'Я',
	'я': 'я',
}

func toLatin(str string) string {
	str = norm.NFC.String(str)

	result, _, err := transform.String(transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn))), str)
	if err == nil {
		str = result
	}

	var buffer bytes.Buffer

	for _, r := range []rune(str) {
		if unicode.Is(unicode.Cyrillic, r) {
			buffer.WriteRune(translit[r])

			continue
		}

		buffer.WriteRune(r)
	}

	return buffer.String()
}
