package main

import (
	"fmt"
	"testing"
)

func TestImageRegex(t *testing.T) {
	cases := map[string]struct {
		name  string
		match bool
	}{
		"jpg":          {"image.jpg", true},
		"JPG":          {"image.JPG", true},
		"jpeg":         {"image.jpeg", true},
		"JPeG":         {"image.JPeG", true},
		"JPG.zip":      {"image.JPG.zip", false},
		"no extention": {"imageJPG", false},
		"movie":        {"movie.MOV", false},
	}
	var caseNameLength, nameLength int
	for name, tc := range cases {
		if len(name) > caseNameLength {
			caseNameLength = len(name)
		}
		if len(tc.name) > nameLength {
			nameLength = len(tc.name)
		}
	}
	format := fmt.Sprintf("%%%ds: %%%ds %%s", caseNameLength, nameLength)
	for name, tc := range cases {
		if match := imageRegex.MatchString(tc.name); match != tc.match {
			t.Errorf(format, name, tc.name,
				func() string {
					if tc.match {
						return "didn't match but shoud"
					}
					return "matched but shoudn't"
				}(),
			)
		}
	}

}

func TestParceLocationString(t *testing.T) {
	cases := map[string]struct {
		location string
		lat      float64
		lon      float64
	}{
		"default": {"0.0,0.0", 0.0, 0.0},
		"KFC":     {"53.0326000,158.6307500", 53.0326, 158.63075},
	}
	var caseNameLength int
	for name := range cases {
		if len(name) > caseNameLength {
			caseNameLength = len(name)
		}
	}
	format := fmt.Sprintf("%%%ds: %%22s should be %%f,%%f, not %%f,%%f", caseNameLength)
	for name, tc := range cases {
		if lat, lon := parceLocationString(tc.location); tc.lat != lat || tc.lon != lon {
			t.Errorf(format, name, tc.location, tc.lat, tc.lon, lat, lon)
		}
	}
}
