package main

import (
	"fmt"
	"unicode"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

func isMn(r rune) bool { return unicode.Is(unicode.Mn, r) }

func main() {
	str1 := "cafe"
	str2 := "café"
	str3 := "cafe\u0301"
	fmt.Println(str1 == str2)
	fmt.Println(str2 == str3)

	t := transform.Chain(norm.NFD)
	str1a, _, _ := transform.String(t, str1)
	str2a, _, _ := transform.String(t, str2)
	str3a, _, _ := transform.String(t, str3)

	fmt.Println(str1a == str2a)
	fmt.Println(str2a == str3a)
}
