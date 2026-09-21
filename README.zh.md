# cchooklint

[English](README.md) | [日本語](README.ja.md) | [中文](README.zh.md)

> **非官方工具。** 与 Anthropic 无关，未获得其认可或合作关系。

`cchooklint` 是一个诊断用 CLI 工具，用于在 Windows 上静态分析 Claude Code 的 `.claude/settings.json` hooks 配置，检测两个已在实机验证过的具体问题：

1. **`Bash`/`PowerShell` matcher 工具名拼写错误** —— 会导致该 hook 被静默禁用，Claude Code 本身不会报错或提示。
2. **安全防护类 hook 的 `matcher` 仅覆盖 `Bash` 或 `PowerShell` 中的一个** —— 当使用另一个工具时，该 hook 会静默地不触发（"fail open" 漏洞）。

## 为什么要做这个工具

在 Windows 上，Claude Code 可以通过两种不同的工具执行命令：Bash 工具（Git Bash）和 PowerShell 工具。hook 的 `matcher` 按工具名过滤，因此配置为 `matcher: "Bash"` 的 hook 只会在使用 Bash 工具时触发 —— 如果 Claude 改用 PowerShell，该 hook 会被静默跳过。

这对于"拦截危险命令"这类安全防护 hook 尤其危险：配置者往往不会意识到自己的安全防护已经有一半被静默禁用了。上述两个问题，在本项目开始之前已经在真实的 Windows 环境中复现并验证过。

`cchooklint` 是只读的：它不会修改你的配置文件，也不会在运行时实际触发 hook 进行测试。它只检查上述特定的 `Bash`/`PowerShell` 漏洞，并非通用的工具名检查工具。

## 安装

```sh
go install github.com/su-fu/cchooklint/cmd/cchooklint@latest
```

（目前尚未发布正式版本，请暂时从源码构建。）

## 用法

```sh
cchooklint
```

会按以下顺序查找配置文件：

- `.claude/settings.json`（项目级）
- `.claude/settings.local.json`（项目级，本地覆盖）
- `~/.claude/settings.json`（用户级）

对于存在上述两类问题的 hook，会分别输出警告。

### 示例

假设 `.claude/settings.json` 中存在一个拼写错误的 matcher，以及一个只覆盖 `Bash` 的 hook：

```sh
$ cchooklint --lang=zh
[WARN] "Bahs" 与已知工具名不匹配。你是否想输入 "Bash"？
[WARN] 此 hook 仅覆盖 "Bash"，当使用另一个工具时不会触发。建议改为 "Bash|PowerShell"。
```

如需机器可读的输出，请使用`--format=json`。JSON包含带版本的包装对象、
稳定的规则代码以及消息的原始参数。JSON输出不会受`--lang`影响。

### 显示语言

输出语言按以下优先级决定：

1. `--lang` 参数（例如 `cchooklint --lang=zh`）
2. `CCHOOKLINT_LANG` 环境变量
3. `en`（默认）

支持的语言：`en`、`ja`、`zh`。不支持的值会回退到 `en`。

### 退出代码

- `0` — 没有发现问题，也没有工具错误
- `1` — 发现了一个或多个钩子问题
- `2` — 发现或读取设置文件时发生错误

## 许可证

[MIT](LICENSE)
