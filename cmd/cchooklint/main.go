package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/su-fu/cchooklint/internal/discover"
	"github.com/su-fu/cchooklint/internal/i18n"
	"github.com/su-fu/cchooklint/internal/model"
	"github.com/su-fu/cchooklint/internal/rules"
)

func main() {
	os.Exit(run())
}

func run() int {
	format := flag.String("format", "text", "Output format: text or json")
	lang := flag.String("lang", "", "Language")
	flag.Parse()
	if *format != "text" && *format != "json" {
		fmt.Fprintf(os.Stderr, "unsupported output format %q (want text or json)\n", *format)
		return 2
	}

	resolvedLang := *lang
	if resolvedLang == "" {
		resolvedLang = os.Getenv("CCHOOKLINT_LANG")
	}
	if resolvedLang == "" {
		resolvedLang = "en"
	}

	jsonFindings := make([]jsonFinding, 0)
	paths, err := discover.Find()
	if err != nil {
		fmt.Fprintln(os.Stderr, i18n.T(resolvedLang, i18n.MsgDiscoverError, err))
		return 2
	}

	exitCode := 0
	findingsFound := false
	for _, path := range paths {
		settings, err := model.Load(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, i18n.T(resolvedLang, i18n.MsgLoadError, path, err))
			exitCode = 2
			continue
		}
		entries := model.Flatten(path, settings)
		allRules := []rules.Rule{rules.TypoRule{}, rules.CoverageRule{}}
		for _, rule := range allRules {
			findings := rule.Check(entries)
			findingsFound = findingsFound || len(findings) > 0
			for _, finding := range findings {
				if *format == "json" {
					jsonFindings = append(jsonFindings, jsonFinding{
						SourceFile: finding.SourceFile,
						Event:      finding.Event,
						Severity:   finding.Severity,
						Code:       finding.Code,
						Args:       finding.Args,
					})
				} else {
					fmt.Println(i18n.T(resolvedLang, finding.MessageID, finding.Args...))
				}
			}
		}
	}
	if *format == "json" {
		if err := writeJSONFindings(os.Stdout, jsonFindings); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 2
		}
	}
	if findingsFound && exitCode == 0 {
		return 1
	}
	return exitCode
}

type jsonFinding struct {
	SourceFile string `json:"source_file"`
	Event      string `json:"event"`
	Severity   string `json:"severity"`
	Code       string `json:"code"`
	Args       []any  `json:"args"`
}

func writeJSONFindings(writer io.Writer, findings []jsonFinding) error {
	return json.NewEncoder(writer).Encode(struct {
		Version  int           `json:"version"`
		Findings []jsonFinding `json:"findings"`
	}{Version: 1, Findings: findings})
}
