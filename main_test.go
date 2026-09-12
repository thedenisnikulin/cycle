package main

import (
	"strings"
	"testing"
)

// stepCase is an input a step function should recognize; prev picks the direction.
type stepCase struct {
	in   string
	prev bool
	want string
}

// checkStep checks that step turns each case's input into want, and that it rejects every
// input in rejected in both directions.
func checkStep(t *testing.T, name string, step Step, cases []stepCase, rejected []string) {
	t.Helper()

	for _, c := range cases {
		if got, ok := step(c.in, c.prev); !ok || got != c.want {
			t.Errorf("%s(%q, prev=%v) = %q, %v; want %q, true", name, c.in, c.prev, got, ok, c.want)
		}
	}

	for _, in := range rejected {
		for _, prev := range []bool{false, true} {
			if got, ok := step(in, prev); ok {
				t.Errorf("%s(%q, prev=%v) = %q, true; want no match", name, in, prev, got)
			}
		}
	}
}

func TestStepWord(t *testing.T) {
	checkStep(t, "stepWord", stepWord, []stepCase{
		// two-word lists toggle in both directions
		{"true", false, "false"},
		{"false", false, "true"},
		{"true", true, "false"},
		{"yes", false, "no"},
		{"on", false, "off"},
		{"enabled", false, "disabled"},
		{"disable", true, "enable"},
		{"and", false, "or"},
		{"&&", false, "||"},
		{"||", true, "&&"},
		{"==", false, "!="},
		{"!==", false, "==="},

		// case is copied from the input
		{"True", false, "False"},
		{"TRUE", false, "FALSE"},
		{"OFF", false, "ON"},
		{"tRuE", false, "false"}, // mixed case matches but isn't copied

		// longer lists step and wrap around
		{"monday", false, "tuesday"},
		{"Wednesday", true, "Tuesday"},
		{"sunday", false, "monday"},
		{"monday", true, "sunday"},
		{"FRI", false, "SAT"},
		{"January", true, "December"},
		{"dec", false, "jan"},

		// "may" is in both month lists; the full-name list comes first
		{"may", false, "june"},
		{"May", true, "April"},

		// operators
		{":=", false, "="},
		{"=", false, ":="},
		{"<", false, ">"},
		{">=", true, "<="},
		{"++", false, "--"},
		{"-=", false, "+="},

		// Go
		{"var", false, "const"},
		{"Const", false, "Var"},
		{"break", false, "continue"},
		{"int", false, "int8"},
		{"int64", false, "int"},
		{"int", true, "int64"},
		{"uint32", false, "uint64"},
		{"float64", false, "float32"},
		{"byte", false, "rune"},
		{"Println", false, "Print"},
		{"printf", true, "print"},
		{"Sprintf", false, "Sprintln"},
		{"Errorf", false, "Fatalf"},
		{"Lock", false, "Unlock"},
		{"min", false, "max"},

		// Markdown
		{"#", false, "##"},
		{"#", true, "######"},
		{"######", false, "#"},
		{"[ ]", false, "[x]"},
		{"[x]", false, "[ ]"},
		{"[X]", false, "[ ]"},
		{"-", false, "*"},
		{"+", false, "-"},
		{"NOTE", false, "TIP"},
		{"CAUTION", false, "NOTE"},
		{"Warning", true, "Important"},

		// Bash
		{"-eq", false, "-ne"},
		{"-lt", false, "-gt"},
		{"-ge", false, "-le"},
		{"-z", false, "-n"},
		{"$@", false, "$*"},
		{"head", false, "tail"},
		{"local", false, "export"},

		// opposites
		{"get", false, "set"},
		{"Open", false, "Close"},
		{"WIDTH", false, "HEIGHT"},
		{"src", false, "dst"},
		{"prev", false, "next"},
		{"async", false, "sync"},
		{"always", false, "never"},

		// levels and stages
		{"info", false, "warn"},
		{"trace", true, "error"},
		{"ERROR", false, "TRACE"},
		{"TODO", false, "FIXME"},
		{"hack", false, "todo"},
		{"http", false, "https"},
		{"wss", false, "ws"},
		{"dev", false, "staging"},
		{"prod", false, "dev"},
	}, []string{
		"", "untrue", "truee", "true false", "maybe", "tues",
		"int128", "[]", "[ x]", "- [ ]", "#######", "todos", "-eqq",
	})
}

// Every word must be lowercase (the input decides the case) and in only one list, since the
// first list containing a word wins. "may" is the known exception.
func TestWordLists(t *testing.T) {
	seen := map[string]int{}

	for i, words := range wordLists {
		if len(words) < 2 {
			t.Errorf("list %d has fewer than two words: %q", i, words)
		}

		for _, word := range words {
			if word != strings.ToLower(word) {
				t.Errorf("%q in list %d isn't lowercase", word, i)
			}

			if j, ok := seen[word]; ok && word != "may" {
				t.Errorf("%q is in lists %d and %d", word, j, i)
			}

			seen[word] = i
		}
	}
}

func TestStepNumber(t *testing.T) {
	checkStep(t, "stepNumber", stepNumber, []stepCase{
		// decimal
		{"0", false, "1"},
		{"41", false, "42"},
		{"42", true, "41"},
		{"9", false, "10"},
		{"10", true, "9"},
		{"100", true, "99"},
		{"0", true, "-1"},
		{"-1", false, "0"},
		{"-10", true, "-11"},
		{"+5", false, "6"}, // the plus sign isn't kept
		{"9223372036854775806", false, "9223372036854775807"},

		// leading zeros keep the width
		{"007", false, "008"},
		{"099", false, "100"},
		{"010", true, "009"},
		{"000", true, "-001"},

		// hex keeps its width, prefix case and digit case
		{"0x9", false, "0xa"},
		{"0xf", false, "0x10"},
		{"0x09", false, "0x0a"},
		{"0x0f", false, "0x10"},
		{"0x10", true, "0x0f"},
		{"0xff", false, "0x100"},
		{"0xFF", false, "0x100"},
		{"0xAE", false, "0xAF"},
		{"0xAF", false, "0xB0"},
		{"0xAbC", false, "0xABD"}, // any uppercase digit makes the result uppercase
		{"0X1f", false, "0X20"},
		{"0x00ff", false, "0x0100"},
		{"0x0100", true, "0x00ff"},
		{"0xffff", false, "0x10000"},
		{"0x10000", true, "0x0ffff"},
		{"0xdeadbeef", false, "0xdeadbef0"},
		{"0xDEADBEEF", true, "0xDEADBEEE"},
		{"0xcafebabe", false, "0xcafebabf"},
		{"0x7fffffff", false, "0x80000000"},
		{"0x80000000", true, "0x7fffffff"},
		{"0xffffffff", false, "0x100000000"},
		{"0x00000000ffffffff", false, "0x0000000100000000"},
		{"0x7" + strings.Repeat("f", 14) + "e", false, "0x7" + strings.Repeat("f", 15)},
		{"0x7" + strings.Repeat("F", 15), true, "0x7" + strings.Repeat("F", 14) + "E"},

		// hex can go below zero
		{"-0x5", false, "-0x4"},
		{"0x0", true, "-0x1"},
		{"0x00", true, "-0x01"},
		{"0x00000000", true, "-0x00000001"},
		{"-0xff", true, "-0x100"},

		// binary
		{"0b1", true, "0b0"},
		{"0b0", true, "-0b1"},
		{"0b0111", false, "0b1000"},
		{"0b1111", false, "0b10000"},
		{"0b10000", true, "0b01111"},
		{"0B101", false, "0B110"},
		{"0b00000001", false, "0b00000010"},
		{"0b10101010", false, "0b10101011"},
		{"0b11111111", false, "0b100000000"},
		{"0b100000000", true, "0b011111111"},
		{"0b1010101010101010", true, "0b1010101010101001"},
		{"0b" + strings.Repeat("1", 31), false, "0b1" + strings.Repeat("0", 31)},
		{"0b" + strings.Repeat("1", 63), true, "0b" + strings.Repeat("1", 62) + "0"},

		// octal
		{"0o7", false, "0o10"},
		{"0o17", false, "0o20"},
		{"0O17", false, "0O20"},
		{"0o0", true, "-0o1"},
		{"0o007", false, "0o010"},
		{"0o10", true, "0o07"},
		{"0o0644", true, "0o0643"},
		{"0o0755", false, "0o0756"},
		{"0o777", false, "0o1000"},
		{"0o1000", true, "0o0777"},
		{"0o17777777777", false, "0o20000000000"},
		{"0o" + strings.Repeat("7", 21), true, "0o" + strings.Repeat("7", 20) + "6"},
	}, []string{
		"", "-", "+", "--1", "1.5", "1e3", "12a", "ff", "beef", "99999999999999999999",
		"0x", "0xg", "0x12g", "0x1.5", "0x8000000000000000", "0xffffffffffffffff",
		"0b", "0b2", "0b102", "0b1.0", "0b" + strings.Repeat("1", 64),
		"0o", "0o8", "0o78", "0o1a", "0o" + strings.Repeat("7", 22),
	})
}

func TestStepDate(t *testing.T) {
	checkStep(t, "stepDate", stepDate, []stepCase{
		// dates step by a day
		{"2026-09-11", false, "2026-09-12"},
		{"2026-09-11", true, "2026-09-10"},
		{"2026-09-30", false, "2026-10-01"},
		{"2026-12-31", false, "2027-01-01"},
		{"2026-01-01", true, "2025-12-31"},
		{"2026-02-28", false, "2026-03-01"},
		{"2024-02-28", false, "2024-02-29"},
		{"2024-02-29", false, "2024-03-01"},
		{"2024-03-01", true, "2024-02-29"},
		{"2026/09/30", false, "2026/10/01"},
		{"9999-12-31", false, "10000-01-01"}, // no year cap

		// times step by a minute, or a second when seconds are present, and wrap at midnight
		{"12:30", false, "12:31"},
		{"12:59", false, "13:00"},
		{"23:59", false, "00:00"},
		{"00:00", true, "23:59"},
		{"9:05", false, "09:06"}, // one-digit hours are accepted and padded
		{"12:30:59", false, "12:31:00"},
		{"00:00:00", true, "23:59:59"},

		// date-times roll over into the date
		{"2026-12-31 23:59", false, "2027-01-01 00:00"},
		{"2026-03-01 00:00", true, "2026-02-28 23:59"},
		{"2026-09-11 12:00:00", false, "2026-09-11 12:00:01"},
		{"2026-09-11T08:00", true, "2026-09-11T07:59"},
		{"2026-09-11T08:00:00", true, "2026-09-11T07:59:59"},
		{"2026-09-11T23:59:59", false, "2026-09-12T00:00:00"},
		{"2026-09-11  23:59", false, "2026-09-12 00:00"}, // extra spaces are accepted and collapsed
	}, []string{
		"", "today", "2026", "2026-09", "2026-9-11", "2026-13-01", "2026-02-29", "2026-02-30",
		"2026-09/11", "2026.09.11", "24:00", "12:60", "25:00", "12:30:60", "2026-09-11X23:59",
	})
}

func TestCycle(t *testing.T) {
	tests := []struct {
		in   string
		prev bool
		want string
	}{
		// surrounding whitespace is kept
		{"true", false, "false"},
		{"true ", false, "false "},
		{"  true\n", false, "  false\n"},
		{"\ttrue\t", false, "\tfalse\t"},
		{"\n\n41\n", false, "\n\n42\n"},
		{"  2026-09-11  ", true, "  2026-09-10  "},
		{" [ ] ", false, " [x] "},

		// unrecognized input comes back unchanged
		{"", false, ""},
		{"   ", false, "   "},
		{"\n", true, "\n"},
		{"hello", false, "hello"},
		{" hello ", false, " hello "},
		{"true false", false, "true false"},
		{"enabled = true", false, "enabled = true"},

		// each value goes to the right step function
		{"1.2.3", false, "1.2.4"},
		{"2026-09-11", false, "2026-09-12"},
		{"12:30", false, "12:31"},
		{"0x10", false, "0x11"},
		{"41", true, "40"},
		{"Monday", true, "Sunday"},
	}

	for _, tt := range tests {
		if got := cycle(tt.in, tt.prev); got != tt.want {
			t.Errorf("cycle(%q, prev=%v) = %q, want %q", tt.in, tt.prev, got, tt.want)
		}
	}
}
