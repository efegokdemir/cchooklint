package i18n

import "testing"

func TestTFormatsSupportedLanguages(t *testing.T) {
	tests := []struct {
		name string
		lang string
		id   MessageID
		args []any
		want string
	}{
		{
			name: "english typo warning",
			lang: "en",
			id:   MsgTypoWarning,
			args: []any{"Bahs", "Bash"},
			want: `[WARN] "Bahs" does not match any known tool name. Did you mean "Bash"?`,
		},
		{
			name: "japanese coverage warning",
			lang: "ja",
			id:   MsgCoverageWarning,
			args: []any{"Bash", "Bash|PowerShell"},
			want: `[WARN] このhookは"Bash"しかカバーしておらず、もう片方のツールが使われたときは発火しません。"Bash|PowerShell"への変更を検討してください`,
		},
		{
			name: "chinese typo warning",
			lang: "zh",
			id:   MsgTypoWarning,
			args: []any{"Powershel", "PowerShell"},
			want: `[WARN] "Powershel" 与已知工具名不匹配。你是否想输入 "PowerShell"？`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := T(tt.lang, tt.id, tt.args...); got != tt.want {
				t.Errorf("T(%q, %d, %v) = %q, want %q", tt.lang, tt.id, tt.args, got, tt.want)
			}
		})
	}
}

func TestTFallsBackToEnglish(t *testing.T) {
	want := `[WARN] "Bahs" does not match any known tool name. Did you mean "Bash"?`
	if got := T("fr", MsgTypoWarning, "Bahs", "Bash"); got != want {
		t.Errorf("T(unsupported language) = %q, want %q", got, want)
	}
}
