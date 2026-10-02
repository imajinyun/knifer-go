package main

import "strings"

type coverageTotals struct {
	Statements int     `json:"statements"`
	Covered    int     `json:"covered"`
	Percent    float64 `json:"percent"`
}

func coverageBreakdown(lines []profileLine, module string) map[string]coverageTotals {
	out := map[string]coverageTotals{}
	for _, line := range lines {
		rel := strings.TrimPrefix(line.file, module+"/")
		group := "root"
		switch {
		case strings.HasPrefix(rel, "bin/"):
			group = "governance_go_test"
		case strings.HasPrefix(rel, "internal/"):
			group = "internal"
		case strings.HasPrefix(rel, "v"):
			group = "facade"
		}
		keys := []string{group}
		if group != "governance_go_test" {
			keys = append(keys, "library")
		}
		for _, key := range keys {
			item := out[key]
			item.Statements += line.statements
			if line.count > 0 {
				item.Covered += line.statements
			}
			if item.Statements > 0 {
				item.Percent = float64(item.Covered) * 100 / float64(item.Statements)
			}
			out[key] = item
		}
	}
	return out
}
