package main

import (
	"testing"
)

func TestParseMediaWikiTitle(t *testing.T) {
	p := &parser{wikiMode: true}

	testCases := []struct {
		input    string
		expected string
	}{
		{"123 ''foo'' bar", "foo"},

		{"123 '''foo''' bar", "foo"},

		{"123 '''foo'''' bar", "foo"},

		{"123 ''foo'' '''bar''' ''''baz''''", "foo"},

		{"123 '''foo''' ''bar'' ''''baz''''", "foo"},

		{"123 ''''foo'''' '''bar''' ''baz''", "foo"},
	}
	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			actual := p.parseTitle(tc.input)

			if actual != tc.expected {
				t.Errorf("expected: \"%s\", actual: \"%s\".", tc.expected, actual)
			}
		})
	}
}

func TestParseMediaWikiTitleNotFound(t *testing.T) {
	p := &parser{wikiMode: true}

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("parseTitle didn't panic as expected in MediaWiki mode when missing title.")
		}
	}()

	p.parseTitle("123")
}
