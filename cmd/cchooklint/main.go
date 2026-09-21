package main

import (
	"flag"
	"fmt"
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
	lang := flag.String("lang", "", "Language")
	flag.Parse()

	var resolvedLang string
	if *lang != "" {
		resolvedLang = *lang
	} else {
		envLang := os.Getenv("CCHOOKLINT_LANG")
		if envLang != "" {
			resolvedLang = envLang
		} else {
			resolvedLang = "en"
		}
	}
	paths, err := discover.Find()
	if err != nil {
		fmt.Fprintln(os.Stderr, i18n.T(resolvedLang, i18n.MsgDiscoverError, err))
		return 1
	}

	exitCode := 0
	for _, path := range paths {
		settings, err := model.Load(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, i18n.T(resolvedLang, i18n.MsgLoadError, path, err))
			exitCode = 1
			continue
		}
		entries := model.Flatten(path, settings)
		allRules := []rules.Rule{rules.TypoRule{}, rules.CoverageRule{}}
		for _, rule := range allRules {
			findings := rule.Check(entries)
			for _, finding := range findings {
				fmt.Println(i18n.T(resolvedLang, finding.MessageID, finding.Args...))
			}
		}
	}
	return exitCode
}
