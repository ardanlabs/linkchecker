package main

import (
	"testing"
)

const (
	succeed = "\u2713"
	failed  = "\u2717"
)

func TestParseLinks(t *testing.T) {
	t.Log("Given the need to parse links.")
	{
		want := []string{
			"http://www.google.com",
			"http://www.ardanlabs.com",
			"http://www.ardanlabs.com/consulting",
			"http://www.ardanlabs.com/ultimate-go",
			"https://www.ardanlabs.com/careers",
		}

		got := parseLinks("https://www.ardanlabs.com/", testHtml)

		if len(want) != len(got) {
			t.Errorf("\tShould find %d links got %d. %s", len(want), len(got), failed)
		} else {
			t.Logf("\tShould find %d links got %d. %s", len(want), len(got), succeed)
		}

		notFound := 0
		for _, u1 := range want {
			found := false
			for _, u2 := range got {
				if u1 == u2 {
					found = true
					notFound++
				}
			}

			if !found {
				t.Errorf("\tShould return link %s. %s", u1, failed)
			} else {
				t.Errorf("\tShould return link %s. %s", u1, succeed)
			}
		}

	}
}

const testHtml = `<html><body><a href="http://www.google.com">link</a><a href="http://www.ardanlabs.com">link</a><a href="http://www.ardanlabs.com/consulting">link</a><a href=http://www.ardanlabs.com/ultimate-go>link</a><a href=/careers>link</a></body></html>`
