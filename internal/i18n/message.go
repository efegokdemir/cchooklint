package i18n

import (
	"fmt"
)

type MessageID int

const (
	MsgTypoWarning     MessageID = iota // 0
	MsgCoverageWarning                  // 1
	MsgDiscoverError                    // 2
	MsgLoadError                        // 3
)

var messageEn = map[MessageID]string{
	MsgTypoWarning:     "[WARN] %q does not match any known tool name. Did you mean %q?",
	MsgCoverageWarning: "[WARN] This hook only covers %q. It won't fire when the other tool is used. Consider changing it to %q.",
	MsgDiscoverError:   "[ERROR] Could not discover Claude Code settings: %v",
	MsgLoadError:       "[WARN] Could not read settings file %q: %v",
}

var messageJa = map[MessageID]string{
	MsgTypoWarning:     "[WARN] %q は既知のツール名と一致しません。%q の間違いではありませんか？",
	MsgCoverageWarning: "[WARN] このhookは%qしかカバーしておらず、もう片方のツールが使われたときは発火しません。%qへの変更を検討してください",
	MsgDiscoverError:   "[ERROR] Claude Codeの設定ファイルを検出できませんでした: %v",
	MsgLoadError:       "[WARN] 設定ファイル%qを読み込めませんでした: %v",
}

var messageZh = map[MessageID]string{
	MsgTypoWarning:     "[WARN] %q 与已知工具名不匹配。你是否想输入 %q？",
	MsgCoverageWarning: "[WARN] 此 hook 仅覆盖 %q，当使用另一个工具时不会触发。建议改为 %q。",
	MsgDiscoverError:   "[ERROR] 无法发现 Claude Code 设置文件：%v",
	MsgLoadError:       "[WARN] 无法读取设置文件 %q：%v",
}

var messages = map[string]map[MessageID]string{
	"en": messageEn,
	"ja": messageJa,
	"zh": messageZh,
}

func T(lang string, messageID MessageID, args ...any) string {
	v, ok := messages[lang][messageID]
	if ok {
		return fmt.Sprintf(v, args...)
	}
	return fmt.Sprintf(messages["en"][messageID], args...)
}
