package main

import "regexp"

// MatcherLookup returns the map value of the named captures.
func MatcherLookup(match []string, re *regexp.Regexp) map[string]string {
	result := make(map[string]string)
	for i, name := range re.SubexpNames() {
		if i != 0 {
			result[name] = match[i]
		}
	}
	return result
}
