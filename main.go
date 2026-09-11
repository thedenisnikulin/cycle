package main

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

type Step func(string, bool) (string, bool)

var (
	wordLists = [][]string{
		// booleans and logic
		{"true", "false"},
		{"yes", "no"},
		{"on", "off"},
		{"enable", "disable"},
		{"enabled", "disabled"},
		{"and", "or"},
		{"&&", "||"},
		{"==", "!="},
		{"===", "!=="},

		// operators
		{"<", ">"},
		{"<=", ">="},
		{"+=", "-="},
		{"++", "--"},
		{":=", "="},

		// go
		{"var", "const"},
		{"break", "continue"},
		{"int", "int8", "int16", "int32", "int64"},
		{"uint", "uint8", "uint16", "uint32", "uint64"},
		{"float32", "float64"},
		{"byte", "rune"},
		{"print", "printf", "println"},
		{"sprint", "sprintf", "sprintln"},
		{"errorf", "fatalf"},
		{"lock", "unlock"},
		{"min", "max"},

		// markdown
		{"#", "##", "###", "####", "#####", "######"},
		{"[ ]", "[x]"},
		{"-", "*", "+"},
		{"note", "tip", "important", "warning", "caution"},

		// bash
		{"-eq", "-ne"},
		{"-lt", "-gt"},
		{"-le", "-ge"},
		{"-z", "-n"},
		{"$*", "$@"},
		{"head", "tail"},
		{"local", "export"},

		// opposites
		{"get", "set"},
		{"open", "close"},
		{"read", "write"},
		{"push", "pop"},
		{"start", "end"},
		{"first", "last"},
		{"up", "down"},
		{"left", "right"},
		{"top", "bottom"},
		{"width", "height"},
		{"src", "dst"},
		{"from", "to"},
		{"old", "new"},
		{"prev", "next"},
		{"before", "after"},
		{"req", "res"},
		{"input", "output"},
		{"allow", "deny"},
		{"include", "exclude"},
		{"show", "hide"},
		{"sync", "async"},
		{"pass", "fail"},
		{"always", "never"},

		// other dev stuff
		{"trace", "debug", "info", "warn", "error"},
		{"todo", "fixme", "hack"},
		{"http", "https"},
		{"ws", "wss"},
		{"dev", "staging", "prod"},

		// calendar; "may" is in both month lists and the full-name list wins
		{"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"},
		{"mon", "tue", "wed", "thu", "fri", "sat", "sun"},
		{"january", "february", "march", "april", "may", "june", "july", "august", "september",
			"october", "november", "december"},
		{"jan", "feb", "mar", "apr", "may", "jun", "jul", "aug", "sep", "oct", "nov", "dec"},
	}

	dateLayouts = map[string]time.Duration{
		"2006-01-02":          24 * time.Hour,
		"2006/01/02":          24 * time.Hour,
		"15:04":               time.Minute,
		"15:04:05":            time.Second,
		"2006-01-02 15:04":    time.Minute,
		"2006-01-02 15:04:05": time.Second,
		"2006-01-02T15:04":    time.Minute,
		"2006-01-02T15:04:05": time.Second,
	}

	semverRe = regexp.MustCompile(`^([vV]?\d+\.\d+\.)(\d+)$`)
	numberRe = regexp.MustCompile(`^([+-]?)(0[xXbBoO])?([0-9a-fA-F]+)$`)

	steps = []Step{
		stepDate,
		stepSemver,
		stepWord,
		stepNumber,
	}
)

func main() {
	input, _ := io.ReadAll(os.Stdin)
	fmt.Print(cycle(string(input), slices.Contains(os.Args[1:], "--prev")))
}

func cycle(text string, prev bool) string {
	value := strings.TrimSpace(text)
	start := strings.Index(text, value)

	for _, step := range steps {
		if next, ok := step(value, prev); ok {
			return text[:start] + next + text[start+len(value):]
		}
	}

	return text
}

func stepDate(s string, prev bool) (string, bool) {
	for layout, unit := range dateLayouts {
		if t, err := time.Parse(layout, s); err == nil {
			if prev {
				unit = -unit
			}

			return t.Add(unit).Format(layout), true
		}
	}

	return "", false
}

func stepSemver(s string, prev bool) (string, bool) {
	m := semverRe.FindStringSubmatch(s)

	if m == nil {
		return "", false
	}

	patch, err := strconv.Atoi(m[2])

	if prev {
		patch--
	} else {
		patch++
	}

	return m[1] + strconv.Itoa(max(patch, 0)), err == nil
}

func stepWord(s string, prev bool) (string, bool) {
	for _, words := range wordLists {
		for i, word := range words {
			if !strings.EqualFold(word, s) {
				continue
			}

			j := i + 1

			if prev {
				j = i - 1
			}

			next := words[(j+len(words))%len(words)]

			switch {
			case s != strings.ToLower(s) && s == strings.ToUpper(s):
				next = strings.ToUpper(next)
			case s[:1] == strings.ToUpper(s[:1]):
				next = strings.ToUpper(next[:1]) + next[1:]
			}

			return next, true
		}
	}

	return "", false
}

func stepNumber(s string, prev bool) (string, bool) {
	m := numberRe.FindStringSubmatch(s)

	if m == nil {
		return "", false
	}

	sign, prefix, digits := m[1], m[2], m[3]
	base, verb := 10, "d"

	switch strings.ToLower(prefix) {
	case "0x":
		base, verb = 16, "x"

		if strings.ContainsAny(digits, "ABCDEF") {
			verb = "X"
		}
	case "0b":
		base, verb = 2, "b"
	case "0o":
		base, verb = 8, "o"
	}

	n, err := strconv.ParseInt(sign+digits, base, 64)

	if err != nil {
		return "", false
	}

	if prev {
		n--
	} else {
		n++
	}

	width := 0

	if prefix != "" || digits[0] == '0' {
		width = len(digits) // keeps 007 -> 008 and 0x10 -> 0x0f
	}

	sign = ""

	if n < 0 {
		sign, n = "-", -n
	}

	return fmt.Sprintf("%s%s%0*"+verb, sign, prefix, width, n), true
}
