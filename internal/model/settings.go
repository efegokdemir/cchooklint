package model

import (
	"encoding/json"
	"os"
	"strings"
)

type Settings struct {
	Hooks map[string][]HookMatcherGroup `json:"hooks"`
}

type HookMatcherGroup struct {
	Matcher string        `json:"matcher"`
	Hooks   []HookCommand `json:"hooks"`
}

type HookCommand struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

type HookEntry struct {
	SourceFile string
	Event      string
	Matcher    string
	Command    string
}

func Load(path string) (Settings, error) {
	var result Settings
	loadByte, err := os.ReadFile(path)

	if err != nil {
		return result, err
	}

	err = json.Unmarshal(loadByte, &result)
	if err != nil {
		return result, err
	}

	return result, nil
}

func Flatten(sourceFile string, s Settings) []HookEntry {
	var result []HookEntry

	for event, groups := range s.Hooks {
		for _, group := range groups {
			for _, cmd := range group.Hooks {
				result = append(result, HookEntry{SourceFile: sourceFile, Event: event, Matcher: group.Matcher, Command: cmd.Command})
			}
		}
	}

	return result
}

func SplitMatcher(matcher string) []string {
	result := strings.FieldsFunc(matcher, func(r rune) bool { return r == '|' || r == ',' })
	return result
}

func IsRegexMatcher(matcher string) bool {
	result := strings.ContainsAny(matcher, ".*+?()[]{}^$\\")
	return result
}

func EditDistance(a, b string) int {
	m := len(a)
	n := len(b)
	table := make([][]int, m+1)
	for i := range table {
		table[i] = make([]int, n+1)
	}

	for i := 0; i <= m; i++ {
		table[i][0] = i
	}

	for j := 0; j <= n; j++ {
		table[0][j] = j
	}

	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if a[i-1] == b[j-1] {
				table[i][j] = table[i-1][j-1]
			} else {
				table[i][j] = 1 + min(table[i-1][j], table[i][j-1], table[i-1][j-1])
			}
		}
	}

	return table[m][n]
}
