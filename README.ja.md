# cchooklint

[English](README.md) | [日本語](README.ja.md) | [中文](README.zh.md)

> **非公式ツールです。** Anthropicとは無関係であり、公認・提携もしていません。

`cchooklint`は、Windows上でのClaude Codeの`.claude/settings.json`のhooks設定を静的解析し、実機で検証済みの2種類のバグを検出する診断用CLIです。

1. **`Bash`/`PowerShell`のmatcherツール名の綴りミス** — hookが無音で無効化されます。Claude Code側からエラーも警告も出ません。
2. **安全装置系のhookの`matcher`が`Bash`か`PowerShell`の片方しかカバーしていない** — もう片方のツールが使われたとき、そのhookは無音で発火しません（「fail open」の穴）。

## 作った理由

Windows上のClaude Codeは、シェルコマンドを「Bashツール（Git Bash）」と「PowerShellツール」という2種類のツールで実行できます。hookの`matcher`はツール名でフィルタするため、`matcher: "Bash"`とだけ書かれたhookは、Bashツールが使われたときしか発火しません。Claudeが代わりにPowerShellを使った場合、そのhookは無音でスキップされます。

これは特に「危険なコマンドをブロックする」ような安全装置系のhookで危険です。設定した本人が、安全装置の半分が無音で無効化されていることに気づかないままになりがちだからです。上記2つのバグは、このプロジェクトを始める前に実機のWindows環境で再現・検証済みです。

`cchooklint`は読み取り専用です。設定ファイルを書き換えることも、hookを実際に発火させて動作テストすることも一切しません。上記の`Bash`/`PowerShell`の穴だけに絞った専用チェックであり、ツール名全般をチェックする汎用ツールではありません。

## インストール

```sh
go install github.com/su-fu/cchooklint/cmd/cchooklint@latest
```

（まだバージョン管理されたリリースは公開していません。現時点ではソースからビルドしてください。）

## 使い方

```sh
cchooklint
```

以下の順で設定ファイルを探索します。

- `.claude/settings.json`（プロジェクト）
- `.claude/settings.local.json`（プロジェクト、ローカル上書き）
- `~/.claude/settings.json`（ユーザー）

上記2種類のバグに該当するhookがあれば、それぞれ警告を表示します。

### 実行例

`.claude/settings.json`に、綴りミスのあるmatcherと、`Bash`しかカバーしていないhookがある場合:

```sh
$ cchooklint --lang=ja
[WARN] "Bahs" は既知のツール名と一致しません。"Bash" の間違いではありませんか？
[WARN] このhookは"Bash"しかカバーしておらず、もう片方のツールが使われたときは発火しません。"Bash|PowerShell"への変更を検討してください
```

機械可読な出力には`--format=json`を使用します。JSONにはバージョン付きの
エンベロープ、安定したルールコード、メッセージの元引数が含まれます。
JSON出力では`--lang`は無視されます。

### 表示言語

出力言語は以下の優先順位で決まります。

1. `--lang`フラグ（例: `cchooklint --lang=ja`）
2. `CCHOOKLINT_LANG`環境変数
3. `en`（デフォルト）

対応言語: `en`, `ja`, `zh`。未対応の値は`en`にフォールバックします。

### 終了コード

- `0` — 指摘もツールエラーもありません
- `1` — 1つ以上のフックの指摘が見つかりました
- `2` — 設定ファイルの検出または読み込みに失敗しました

## ライセンス

[MIT](LICENSE)
