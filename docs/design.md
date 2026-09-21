# cchooklint 設計資料

- 作成日: 2026-09-19
- 位置づけ: 実装を進めながら随時更新する生きたドキュメント（決定事項が変わったらこのファイルを直接書き換える）

## 背景・課題

Claude CodeはWindows上で「Bashツール（Git Bash）」と「PowerShellツール」の2種類でコマンドを実行できる。hooksの`matcher`はツール名でフィルタするため、`matcher: "Bash"`だけを書いたhookは、Claudeが**PowerShellツールを使った場合は発火しない**（エラーも出ず無音でスキップされる）。

これを個人の検証プロジェクトで実機再現した（参照: Knowledgesリポジトリの`claude_code_windows_hooks_investigation.md`）。特に危険なのは、「危険コマンドをブロックする」といった安全装置系のhookがこの穴を持っていると、**設定した本人が気づかないまま安全装置が半分無効化される**（fail open）という点。

さらに検証中、`matcher`の綴りミス（例: `"Bahs"`）も同じ症状（無音で無効化）を起こすことを確認済み。静的な注意喚起記事だけでは防げず、実際に設定ファイルを読んで検証するツールが必要、というのがこのプロジェクトの動機。

## v1のスコープ

**やること:**
- `.claude/settings.json` 系ファイル（プロジェクト直下・`settings.local.json`・`~/.claude/settings.json`）を静的解析
- hooksの`matcher`について、以下2種類の問題を検出する読み取り専用の診断CLI
  1. `Bash`/`PowerShell`の綴りミス（Typo Rule）
  2. 安全装置らしいhookが`Bash`/`PowerShell`の片方しかカバーしていない（Coverage Rule）

**やらないこと（v1では）:**
- 自動修正（設定ファイルを書き換える機能）
- hookを実際に発火させての動作テスト
- `Bash`/`PowerShell`以外の全ツール名の網羅チェック

理由: 網羅的なツール名チェックはClaude Code側のツール名一覧を継続的にメンテするコストが発生し続ける。今回実証したのは「Bash/PowerShell限定の穴」なので、そこに絞れば個人開発の時間内で作り切れる規模に収まる。

## アーキテクチャ

```
cchooklint/
├── cmd/cchooklint/main.go     # エントリポイント、フラグ解析
├── internal/
│   ├── i18n/                   # メッセージID・言語ごとの文字列テーブル
│   ├── discover/                # 設定ファイルの探索
│   ├── model/                    # hooks部分のJSON型定義
│   ├── rules/                     # ルールエンジン（Typo Rule, Coverage Rule）
│   └── report/                     # 結果の出力フォーマット
├── testdata/                          # 既知の壊れたsettings.jsonサンプル
└── docs/
    ├── design.md                        # このファイル
    └── progress.md                       # 進捗・学習ログ
```

## データモデル（案）

```go
type HookEntry struct {
    SourceFile     string   // どのsettings.jsonから来たか
    Event          string   // "PreToolUse" など
    Matcher        string   // 生の文字列
    Tokens         []string // "|"/","で分割した場合のトークン（exact-listモードのときのみ）
    IsRegexMatcher bool     // matcherが正規表現扱いになる文字を含むか
    Command        string
    Args           []string
    ScriptPath     string   // argsやcommandから解決したローカルスクリプトパス（あれば）
}
```

## ルールエンジン v1

```go
type Rule interface {
    Check(entries []HookEntry) []Finding
}

type Finding struct {
    Severity   string // "warn" | "info"
    SourceFile string
    Event      string
    MessageID  i18n.MessageID
    Args       []any // i18nメッセージのフォーマット引数
}
```

### Rule 1: 綴りチェック（Typo Rule）
- `matcher`がexact-listモード（`|`/`,`区切りの完全一致リスト、正規表現特殊文字を含まない）のとき、各トークンについて`"Bash"``"PowerShell"`との編集距離を計算
- 完全一致・大文字小文字違いのどちらでもなく、編集距離が近い（目安: 2以下）場合に警告

### Rule 2: カバレッジチェック（Coverage Rule）
- 対象イベント: `PreToolUse`, `PostToolUse`, `PostToolUseFailure`, `PermissionRequest`, `PermissionDenied`
- `matcher`のトークン集合が`{"Bash"}`または`{"PowerShell"}`のどちらか片方のみ
- かつ`ScriptPath`先のファイル内容（読める場合）または`Command`文字列に危険操作らしきキーワード（`deny`, `permissionDecision`, `block`, `exit 2`, `rm -rf`, `Remove-Item`, `Force`）が含まれる
- → 警告し、`"Bash|PowerShell"`への修正案を提示

## i18n設計

外部ライブラリは使わず自前実装（学習目的・依存最小化のため）。

- メッセージはIDで管理し、`internal/i18n`パッケージ内に言語ごとの文字列テーブルを持つ
- 対応言語: `en`（デフォルト）, `ja`, `zh`
- 言語選択の優先順位: `--lang`フラグ > `CCHOOKLINT_LANG`環境変数 > デフォルト`en`
- 未対応言語が指定された場合は`en`にフォールバック

## CLI出力イメージ

```
> cchooklint scan
[WARN] .claude/settings.local.json > hooks.PreToolUse[1] (matcher: "Bash")
  This hook looks like a safety guard, but it won't fire when the PowerShell tool is used.
  Suggestion: "matcher": "Bash|PowerShell"

[WARN] .claude/settings.local.json > hooks.PreToolUse[0] (matcher: "Bahs")
  "Bahs" does not match any known tool name. Did you mean "Bash"?
```

## 配布・収益化の順序

1. Go製の単一バイナリでOSS公開（無料、まずはprivateで開発しpublicに切り替え）
2. 検証済みの需要がある場所（GitHub Issue #46601, #90077, #59225、dev.to記事のスレッド）に、この検証結果とツールを紹介するコメントを投稿
3. 反応を見てから、有料hooksテンプレート集や発展ルール（実動作テスト）の追加を検討

## 命名の経緯

- `cchdoctor`案 → v1のスコープ（2ルールのみ）に対して大げさすぎると判断し却下
- `hooklint` → スコープに見合った謙虚な名前として候補に
- Claude Code関連であることを名前に示す必要があると判断（配布戦略がGitHub Issue/記事経由の検索流入前提のため）
- 既存コミュニティツール（`ccusage`等）の命名慣習（`cc` + 単語、ハイフンなし）に合わせ、**`cchooklint`**に決定
- README等の対外文書には「Anthropic非公式の個人開発ツールです」と明記する

## マイルストーン

0. 環境準備・プロジェクト雛形（`go mod init`、`--lang`フラグを読むだけの最小main.go）
1. i18nの土台（メッセージID・言語ごとの文字列テーブル・選択ロジック）
2. discover（設定ファイル探索）
3. model（hooks部分のJSONパース）
4. Rule 1（綴りチェック）+ i18n経由での出力
5. Rule 2（カバレッジチェック）
6. testdata整備・README・GitHub公開・Issueへの紹介コメント
