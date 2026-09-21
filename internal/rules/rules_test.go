package rules

import (
	"testing"

	"github.com/su-fu/cchooklint/internal/model"
)

func TestRules(t *testing.T) {
	cases := []struct {
		name string // このケースの名前
		file string // 読み込むJSONファイル
		rule Rule   // どのルールでチェックするか
		want int    // 期待する警告件数
	}{
		{"typo", "../../testdata/typo.json", TypoRule{}, 1},
		{"coverage_gap", "../../testdata/coverage_gap.json", CoverageRule{}, 1},
		{"clean", "../../testdata/clean.json", CoverageRule{}, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			settings, err := model.Load(c.file)
			if err != nil {
				t.Fatalf("Load failed: %v", err)
			}
			entries := model.Flatten(c.file, settings)
			findings := c.rule.Check(entries)
			if len(findings) != c.want {
				t.Errorf("len(findings) = %d, want %d", len(findings), c.want)
			}
		})
	}
}
