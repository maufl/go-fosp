// Copyright (C) 2015 Felix Maurer
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>

package main

import (
	"net/url"
	"testing"
)

type urlFamilyTest struct {
	test   string
	result []string
}

var urlFamilyTests = []urlFamilyTest{
	{test: "fosp://alice@maufl.de/", result: []string{"fosp://alice@maufl.de/"}},
	{test: "fosp://a@b/1/2/3/4/", result: []string{"fosp://a@b/1/2/3/4", "fosp://a@b/1/2/3", "fosp://a@b/1/2", "fosp://a@b/1", "fosp://a@b/"}},
	{test: "fosp://a@b", result: []string{"fosp://a@b/"}},
	{test: "fosp://a@b/.", result: []string{"fosp://a@b/"}},
	{test: "fosp://alice@localhost.localdomain/me", result: []string{"fosp://alice@localhost.localdomain/me", "fosp://alice@localhost.localdomain/"}},
}

func TestURLFamily(t *testing.T) {
	for _, test := range urlFamilyTests {
		baseUrl, err := url.Parse(test.test)
		if err != nil {
			t.Errorf("Could not execute testcase, URL parsing of %s failed: %s", baseUrl, err)
			continue
		}
		urls := urlFamily(baseUrl)
		if len(urls) != len(test.result) {
			t.Errorf("Expected to get %d URLs but got %d: %v <=> %v", len(test.result), len(urls), test.result, urls)
		}
		urlStrings := make([]string, len(urls))
		for i, u := range urls {
			urlStrings[i] = u.String()
		}
		for _, u := range urls {
			if !contains(test.result, u.String()) {
				t.Errorf("Test result is expected to contain %s but does not", u)
			}
		}
		for _, us := range test.result {
			if !contains(urlStrings, us) {
				t.Errorf("Expected to get back %s but is not in %v", us, urlStrings)
			}
		}
	}
}
