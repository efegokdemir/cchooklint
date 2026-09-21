# cchooklint 進捗ログ

このファイルには、マイルストーンの進捗と、その過程でできるようになったこと（学んだGoの概念など）を記録していく。新しいエントリは上に追記する。

## 2026-09-21: Milestone 6 進行中（testdata整備）

Milestone 6（testdata整備・README・GitHub公開・Issueへの紹介コメント）のうち、まず`testdata整備`から着手。GitHub公開・Issue投稿は外部に見える・元に戻しにくいアクションなので、進める前にユーザーに順番を確認し、testdata→README→公開/アウトリーチ相談の順で進める方針にした。

- `testdata/`（design.mdのアーキテクチャ通りリポジトリ直下）に3つのサンプル`settings.json`を作成（JSONデータなのでこちらが直接作成）
  - `typo.json`: `matcher: "Bahs"`（綴りミス検出用）
  - `coverage_gap.json`: `matcher: "Bash"`単体 + 危険なコマンド（カバレッジ漏れ検出用）
  - `clean.json`: `matcher: "Bash|PowerShell"`（両方カバー済み）+ 危険なコマンド（誤検知しないことの確認用）
- `internal/rules/rules_test.go` を新規実装。これまでは動作確認のたびに一時テストファイルを作って消していたが、今回は消さずに残る本物のテストにした
  - まず単発のテスト（`typo.json`を`Load`→`Flatten`→`TypoRule{}.Check`し、件数を確認）を書き、その後テーブル駆動テストの`TestRules`に発展させ、単発版は重複するため削除
  - `TestRules`は3ケース（typo/coverage_gap/clean）を無名構造体のスライスにまとめ、`t.Run`でサブテストとして実行。`rule Rule`フィールドに`TypoRule{}`/`CoverageRule{}`どちらも入れられる点で、Milestone 5で学んだインターフェースの利点を再び実感する形に
  - `go test ./...`で全パッケージ（テストがあるのは`rules`のみ）が通ることを確認
- 学んだGoの概念:
  - `_test.go`ファイル・`func TestXxx(t *testing.T)`という命名規則、`go test`がこれを自動検出する仕組み
  - `t.Fatalf`（即座に打ち切り）と`t.Errorf`（記録して続行）の使い分け
  - `go test`実行時の作業ディレクトリは「そのテストが属するパッケージのディレクトリ」になるため、`testdata`への相対パスの書き方に注意が必要
  - テーブル駆動テスト: `[]struct{...}{...}`という無名構造体のスライスで「入力と期待値の組み合わせ」を表にまとめ、`for`でループしながら同じ処理を繰り返すGoの定番パターン
  - `t.Run(name, func(t *testing.T) {...})`によるサブテスト。内側のクロージャで外側と同じ名前`t`を使っても、内側のスコープが優先される
  - テストデータ（`"rm -rf /"`のような文字列）は静的解析ツールの中では単なる比較対象の文字列でしかなく、実行されることはない、という安全性の確認（`os/exec`系のコードがコードベースに存在しないことをgrepで確認）
- 次のステップ: README作成、そのあとGitHub公開・Issueへの紹介コメントについて相談

## 2026-09-21: Milestone 5完了（Rule 2: カバレッジチェック）

design.mdのRule 2のうち、`ScriptPath`（外部スクリプトファイルの中身を読む部分）はスコープ外とし、`Command`文字列のみを対象にする方針で進めた（`HookEntry`にはまだ`ScriptPath`フィールドを追加していない）。

- `internal/rules/rules.go` に3つの部品を追加
  - `targetEvents`: 対象5イベント（`PreToolUse`, `PostToolUse`, `PostToolUseFailure`, `PermissionRequest`, `PermissionDenied`）を`map[string]bool`で保持し、O(1)で所属判定
  - `isSingleToolMatcher(matcher string) bool`: `model.SplitMatcher`でトークン化し、トークンが1個かつその中身が`"Bash"`または`"PowerShell"`と一致するかを判定
  - `containsDangerKeyword(command string) bool`: `deny`, `permissionDecision`, `block`, `exit 2`, `rm -rf`, `Remove-Item`, `Force`のいずれかを`strings.Contains`で含むか判定
  - `CoverageRule`（`Rule`インターフェースを満たす2つ目の型）: 上記3条件がすべて真のとき、修正案`"Bash|PowerShell"`付きの`Finding`を生成
- `internal/i18n/message.go`の`MsgCoverageWarning`を、`%q`を2つ含むテンプレートに更新（en/ja/zh）。中国語訳は前回同様ユーザーが話せないためこちらが用意
- `cmd/cchooklint/main.go`を更新し、`[]rules.Rule{TypoRule{}, CoverageRule{}}`という形でRuleインターフェースを実際に活用。`for _, rule := range allRules { rule.Check(entries) ... }`という、Milestone 4で学んだインターフェースの「同じ形の型をまとめて扱える」という利点を実践する形になった
- 動作確認: `CoverageRule.Check`単体でのテスト（Bash単体+危険操作→警告、両方カバー→警告なし、安全なコマンド→警告なし、対象外イベント→警告なし、PowerShell単体+`Force`→警告）に加え、実際の`.claude/settings.json`（typoとカバレッジ漏れが両方存在するケース）で`go run`し、en/ja/zh全言語でTypoRule・CoverageRule両方の警告が正しく表示され、`"Bash|PowerShell"`と両方カバーしているエントリでは警告が出ないことを確認
- 学んだGoの概念:
  - `map[string]bool`を「集合（set）」として使うパターン（スライスで毎回ループ比較するよりO(1)で判定できる）
  - `== false`ではなく`!`（否定演算子）を使うのがGoの慣習
  - すでに確定した情報（トークンが1個であること）を再計算する無駄なループに気づいて自分で削る、というリファクタリング判断
  - `[]Rule{TypoRule{}, CoverageRule{}}`のように、インターフェース型のスライスに異なる具象型をまとめて入れられることを実際のコードで体感
- v1スコープの2大ルール（Typo Rule, Coverage Rule）が完成。次のステップ: Milestone 6（testdata整備・README・GitHub公開・Issueへの紹介コメント）

## 2026-09-21: Milestone 4完了（Rule 1: 綴りチェック + i18n経由での出力）

前回（下ごしらえ部分: `SplitMatcher`/`IsRegexMatcher`/`EditDistance`、DPアルゴリズムの学習）の続き。

- `internal/rules/rules.go` を新規実装
  - `Finding`構造体（`Severity`, `SourceFile`, `Event`, `MessageID`, `Args`）と`Rule`インターフェース（`Check(entries []model.HookEntry) []Finding`）を定義
  - `TypoRule`（空の構造体 + `Check`メソッド）を実装。`model.IsRegexMatcher`で正規表現っぽいmatcherを除外 → `model.SplitMatcher`でトークン分割 → `strings.EqualFold`で完全一致・大文字小文字違いを除外 → 残ったトークンについて`model.EditDistance`で`"Bash"`/`"PowerShell"`それぞれとの距離を計算し、近い方を修正候補として`Finding`を生成（距離2以下のみ警告）
  - `internal/model/settings.go`の`splitMatcher`/`isRegexMatcher`/`editDistance`を、`rules`パッケージから呼べるようexported化（`SplitMatcher`/`IsRegexMatcher`/`EditDistance`）
- `internal/i18n/message.go` を拡張
  - `MsgTypoWarning`のテンプレートに`%q`を2つ含める形に変更（en/ja/zh 3言語とも）。中国語訳はユーザーが中国語を話せないため、その場でこちらが用意
  - `T`関数を`T(lang string, messageID MessageID, args ...any) string`に変更し、`fmt.Sprintf(template, args...)`でプレースホルダーに値を埋め込むように拡張
- `cmd/cchooklint/main.go` を更新し、全部品を結線
  - `discover.Find()` → 各パスを`model.Load`/`model.Flatten`で`[]model.HookEntry`化して`append`で合体 → `rules.TypoRule{}.Check(entries)` → `i18n.T(resolvedLang, finding.MessageID, finding.Args...)`で表示、という一気通貫のパイプラインが完成
- 動作確認: 複数パターン（typo検出・完全一致・大文字小文字違い・正規表現除外・遠すぎる文字列・複数トークン）を`TypoRule.Check`単体でテストし全て期待通り。さらに実際に`.claude/settings.json`にtypo入りhooksを置いて`go run`し、en/ja/zh全言語で正しい警告文が表示されることを確認
- 学んだGoの概念:
  - `interface`の基本（構造的部分型）: 特定のシグネチャのメソッドを持ちさえすれば、明示的な宣言なしにそのインターフェースを満たす。`[]Rule{TypoRule{}, ...}`のように異なる型を同じインターフェースとして扱える
  - メソッド（レシーバ付き関数）の書き方: `func (r TypoRule) Check(...) ...`
  - exported/unexportedの制約は、パッケージ内で完結する部品を後から他パッケージに公開したくなったときに関数名の先頭を大文字に変える形で対応する、という実例
  - `strings.EqualFold`による大文字小文字を無視した比較
  - 可変長引数（`args ...any`）と、スライスを可変長引数として展開して渡す`args...`の書き方
  - `fmt.Sprintf`とその書式指定子（`%q`など）
  - `range`をスライスに対して変数1つだけで使うと、値ではなくインデックスが入ってしまうという典型的なハマりどころ（`paths`・`findings`両方で同じミスをして気づいた）
- 次のステップ: Milestone 5（Rule 2: カバレッジチェック）

## 2026-09-20: Milestone 3完了（model: hooks部分のJSONパース）

- `internal/model/settings.go` を実装
  - JSON構造をそのまま写した型 `Settings`（`hooks`）/ `HookMatcherGroup`（`matcher`, `hooks`）/ `HookCommand`（`type`, `command`）を`json`タグ付きで定義
  - `Load(path string) (Settings, error)`: `os.ReadFile` + `json.Unmarshal`でファイルをパース
  - `HookEntry`（`SourceFile`, `Event`, `Matcher`, `Command`の最小4フィールド）を定義。design.mdのフル版にある`Tokens`/`IsRegexMatcher`/`Args`/`ScriptPath`は、実際に使うMilestone 4以降で追加する方針に変更（design.mdは更新せず、進め方の判断としてここに記録）
  - `Flatten(sourceFile string, s Settings) []HookEntry`: ネストした`Settings`を3重の`for range`ループでたどり、フラットな`[]HookEntry`に変換
- 動作確認: 複数matcherグループ・複数コマンドを含むサンプルJSONで`Load`→`Flatten`を一気通貫実行し、期待通り3件の`HookEntry`に分解されることを確認。実機の`~/.claude/settings.json`（hooks未設定）でも`Load`がエラーなく`nil`のHooksを返すことを確認
- 学んだGoの概念:
  - `encoding/json`の`json:"..."`タグで、JSON側のキャメルケースのキーとGoの大文字始まりフィールド名を対応付ける
  - JSON側の値の"形"（文字列・オブジェクト・配列）とGo側の型（`string`・構造体・`[]構造体`）を一致させる考え方（`map`は「キー名が可変なとき」だけ使うものという区別）
  - `os.ReadFile`でファイルを丸ごと読む
  - `for key, value := range map`、`for _, v := range slice`という`range`の基本形（`range`は必ず`for`とセットで使う）
  - 3重にネストした`range`ループで、入れ子のデータ構造をフラットなスライスに変換するパターン
  - `type`はGoの予約語なので変数名に使えない、という制約
  - `:=`と`var`の使い分け（少なくとも1つ新しい変数があれば`:=`で既存変数を再利用できる）
- 次のステップ: Milestone 4（Rule 1: 綴りチェック + i18n経由での出力）

## 2026-09-20: Milestone 2完了（discover: 設定ファイルの探索）

- `internal/discover/filescan.go` を実装
  - `fileExists(path string) bool`: `os.Stat`と`errors.Is(err, os.ErrNotExist)`でファイルの存在確認
  - `homeSettingsPath() (string, error)`: `os.UserHomeDir()`と`filepath.Join`で`~/.claude/settings.json`の絶対パスを組み立て
  - `Find() ([]string, error)`: プロジェクト直下の`.claude/settings.json`・`.claude/settings.local.json`・ホームの`settings.json`のうち、実在するものだけを`[]string`で返す
- 動作確認: 実機で`Find()`を実行し、ホームの`~/.claude/settings.json`（実在）が検出され、プロジェクト直下の2ファイル（このリポジトリには未作成）が含まれないことを確認
- 学んだGoの概念:
  - `os.Stat` + `errors.Is(err, os.ErrNotExist)`によるファイル存在確認（`os.IsNotExist`より推奨される新しい書き方）
  - `os.UserHomeDir()`でOS差異を吸収してホームディレクトリを取得
  - `filepath.Join`でOSごとのパス区切り文字を気にせずパスを組み立てる
  - エラーの種類を区別する必要がない場面では`err != nil`だけで十分（`errors.Is`で特定のエラー種別と誤って比較しない）
  - 戻り値にエラーがある場合は他の戻り値をゼロ値にする、というGoの慣習（エラーと有効な値を同時に返さない）
  - `var result []string` + `append`によるスライスの構築（`append`は新しいスライスを返すので`result = append(...)`と代入し直す必要がある）
  - 関数名とローカル変数名が衝突すると紛らわしい（シャドーイング）という注意点
  - `if a && b`で条件のネストをフラットにするリファクタリング
  - Go Proverbs的な変数名の長さの考え方（生存期間が短い変数は短い名前でもよい）
- 次のステップ: Milestone 3（model: hooks部分のJSONパース）

## 2026-09-20: Milestone 1完了（i18nの土台）

- `internal/i18n/message.go` を実装
  - `type MessageID int` と `const (... = iota)` でメッセージIDを定義（`MsgTypoWarning`, `MsgCoverageWarning`）
  - 言語ごとの文字列テーブル（`map[MessageID]string`）をen/ja/zhの3言語分用意し、`map[string]map[MessageID]string`でまとめた
  - `T(lang string, messageID MessageID) string` で、指定言語の文言を返す。存在しない言語キーの場合は`en`にフォールバック
- `cmd/cchooklint/main.go` を更新
  - `--lang`フラグ（デフォルト空文字列）→ `CCHOOKLINT_LANG`環境変数 → デフォルト`en`、の優先順位で言語を解決するロジックを実装
  - `i18n.T`を呼び出して実際に文言を表示するところまで結線
  - 動作確認: `--lang=ja`で日本語、`--lang=aa`（未対応言語）でenへのフォールバック、フラグと環境変数の両方指定時はフラグ優先、をすべて確認済み
- 学んだGoの概念:
  - `map[K]V`の基本と、mapを値にした2段構えのmap（`map[string]map[MessageID]string`）
  - コンマOKイディオム（`v, ok := m[k]`）でキーの存在確認をしながら値を取り出す
  - nilなmapから読み取ってもpanicせずゼロ値が返る仕様（書き込み時はpanicする点との違い）
  - `string`型は`nil`と比較できない（ゼロ値は`""`であって`nil`ではない）というハマりどころ
  - `if`ブロック内で`return`する場合は`else`を省略するのがGoのイディオム
  - `os.Getenv`で環境変数を読む
  - `goimports`によるimportの整形（標準ライブラリと外部パッケージを空行で分けるスタイル）
- 次のステップ: Milestone 2（discover: 設定ファイルの探索）

## 2026-09-19: Milestone 0完了（プロジェクト雛形）

- `go mod init github.com/su-fu/cchooklint` でモジュール初期化
- `cmd/cchooklint/main.go` に、`--lang`フラグを読んで表示するだけの最小main.goを実装
  - `go run ./cmd/cchooklint --lang=ja` → `ja`、`go run ./cmd/cchooklint` → デフォルト値`en`を確認
- 学んだGoの概念:
  - `flag`パッケージの基本形（`flag.String(name, value, usage) *string` → `flag.Parse()` → 参照外しして利用、の3ステップ）
  - `flag.String`の戻り値がなぜポインタなのか（`flag.Parse()`実行前は値が未確定なため、先に「箱」だけ確保しておく設計）
  - `flag.StringVar`という代替スタイル（既存変数のアドレスを渡す）があることも確認
  - `go run <パッケージのディレクトリ>`でサブディレクトリのmainパッケージを実行する方法
- 次のステップ: Milestone 1（i18nの土台: メッセージID・言語ごとの文字列テーブル・選択ロジック）

## 2026-09-19: プロジェクト開始

- アイデアの元になった「Bash/PowerShell matcherのカバレッジ漏れ」バグを、Knowledgesリポジトリの`claude_code_windows_hooks_investigation.md`で実機再現・検証済み
- ツール名を`cchooklint`に決定（経緯は`design.md`の「命名の経緯」参照）
- リポジトリ作成: `gh repo create su-fu/cchooklint --private` → `ghq get github.com/su-fu/cchooklint`
- `docs/design.md`（設計資料）を作成
- `CLAUDE.md`を作成し、別セッション（Knowledgesリポジトリで作業していたセッション）からこのリポジトリ直下で新規セッションを開始する形に引き継ぎ
- 次のステップ: Milestone 0（`go mod init`、`--lang`フラグを読むだけの最小main.go）。ヒント方式（コードは書かず、標準ライブラリのどこを見ればいいか等を示す）で進める
