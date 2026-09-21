package rules

import (
	"strings"

	"github.com/su-fu/cchooklint/internal/i18n"
	"github.com/su-fu/cchooklint/internal/model"
)

type Rule interface {
	Check(entries []model.HookEntry) []Finding
}

type Finding struct {
	Severity   string // "warn" | "info"
	Code       string // stable machine-readable rule identifier
	SourceFile string
	Event      string
	MessageID  i18n.MessageID
	Args       []any // i18nメッセージのフォーマット引数
}

const (
	CodeMatcherTypo = "matcher_typo"
	CodeCoverageGap = "coverage_gap"
)

type TypoRule struct{}

func (r TypoRule) Check(entries []model.HookEntry) []Finding {
	var result []Finding

	for _, entry := range entries {
		if model.IsRegexMatcher(entry.Matcher) {
			continue
		}
		tokens := model.SplitMatcher(entry.Matcher)
		for _, token := range tokens {
			if strings.EqualFold(token, "Bash") || strings.EqualFold(token, "PowerShell") {
				continue
			}
			distBash := model.EditDistance(token, "Bash")
			distPS := model.EditDistance(token, "PowerShell")
			if distBash <= distPS {
				if distBash <= 2 {
					result = append(result, Finding{Severity: "WARN", Code: CodeMatcherTypo, SourceFile: entry.SourceFile, Event: entry.Event, MessageID: i18n.MsgTypoWarning, Args: []any{token, "Bash"}})
				}

			} else {
				if distPS <= 2 {
					result = append(result, Finding{Severity: "WARN", Code: CodeMatcherTypo, SourceFile: entry.SourceFile, Event: entry.Event, MessageID: i18n.MsgTypoWarning, Args: []any{token, "PowerShell"}})
				}
			}
		}
	}
	return result
}

var targetEvents = map[string]bool{
	"PreToolUse":         true,
	"PostToolUse":        true,
	"PostToolUseFailure": true,
	"PermissionRequest":  true,
	"PermissionDenied":   true,
}

func isSingleToolMatcher(matcher string) bool {
	tokens := model.SplitMatcher(matcher)
	if len(tokens) != 1 {
		return false
	}
	if tokens[0] == "Bash" || tokens[0] == "PowerShell" {
		return true
	}
	return false
}

func containsDangerKeyword(command string) bool {
	var dangerKeywords = []string{"deny", "permissionDecision", "block", "exit 2", "rm -rf", "Remove-Item", "Force"}
	for _, keyword := range dangerKeywords {
		if strings.Contains(command, keyword) {
			return true
		}
	}
	return false
}

type CoverageRule struct{}

func (r CoverageRule) Check(entries []model.HookEntry) []Finding {
	var result []Finding

	for _, entry := range entries {
		if !targetEvents[entry.Event] {
			continue
		}
		if !isSingleToolMatcher(entry.Matcher) {
			continue
		}
		if !containsDangerKeyword(entry.Command) {
			continue
		}
		result = append(result, Finding{Severity: "WARN", Code: CodeCoverageGap, SourceFile: entry.SourceFile, Event: entry.Event, MessageID: i18n.MsgCoverageWarning, Args: []any{entry.Matcher, "Bash|PowerShell"}})
	}
	return result
}
