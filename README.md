# genai-cli-runner

YAMLで宣言した手順(steps)に従って LLM CLI(`claude` / `ollama`)を順番に実行するツール。
各ステップでは複数の CLI を並列に実行でき、その出力を次のステップのプロンプトへテンプレート経由で渡せる。

## 必要なもの

- Go 1.22 以上
- 利用する CLI が `PATH` 上にあること
  - `claude`(Claude Code)
  - `ollama`

## インストール

```sh
go install .
# または
go build -o genai-cli-runner .
```

## 使い方

```sh
genai-cli-runner init [file]         # テンプレートを書き出す(既存ファイルがあればエラー)
genai-cli-runner configcheck [file]  # YAMLとプロンプトテンプレートを静的検証する
genai-cli-runner dryrun [file]       # CLIを呼ばずにプロンプトを展開して標準出力に流す
genai-cli-runner run [file]          # 本番実行する
```

`file` を省略すると `runner.yaml` を使う。
`dryrun` と `run` は実行前に `configcheck` と同じ検証を行う。`run` は Ctrl+C で中断できる。

## 設定ファイル

```yaml
workdir: ./          # CLIの実行ディレクトリ。相対パスの基準
tempdir: ./tmp       # 中間結果の保存先(省略時は workdir)
output: ./output.md  # 最終出力の保存先(省略時は標準出力)
timeout: 10m         # 各CLI呼び出しのタイムアウト(Goのduration形式。省略時は無制限)
env:                 # プロンプトから .Env.<KEY> で参照できる値($VAR は環境変数で展開)
  ENV: development
steps:
  - input:           # プロンプトファイル(連結して1つのプロンプトにする)
      - ./foo.md
      - ./bar.md
    run:             # 並列実行するCLI
      - type: claude
        args: [--model, claude-opus-5-5]
      - type: ollama
        args: [llama3.2]
    expected:        # 出力JSONの想定(dryrun でのダミー値に使う)
      - type: string
        key: joke
        value: This is a test output.
  - input:
      - ./prompt.md
    run:
      - type: claude
        args: [--model, claude-sonnet-5-5]
```

未知のキーはエラーになる。

### CLI の呼び出し

プロンプトは標準入力で渡される。

| type     | 実行されるコマンド                                       |
| -------- | -------------------------------------------------------- |
| `claude` | `claude -p [--output-format json] <args...>`             |
| `ollama` | `ollama run <args...> [--format json]`                   |

最終ステップ以外は JSON 出力を指定して実行し、応答が JSON でなければエラーになる。

### プロンプトテンプレート

`input` のファイルは Go の [text/template](https://pkg.go.dev/text/template) として展開される(`missingkey=error`)。

| 参照          | 内容                                                       |
| ------------- | ---------------------------------------------------------- |
| `.Input`      | 前ステップの出力をJSONとしてパースした値(前ステップの run が1つの時のみ) |
| `.Inputs`     | 同上の配列(run の順)                                     |
| `.Response`   | 前ステップの出力文字列(前ステップの run が1つの時のみ)   |
| `.Responses`  | 同上の配列                                                 |
| `.Env.<KEY>`  | `env` で定義した値                                         |

最初のステップでは `input` が必須で、前ステップ出力は参照できない。
2つ目以降で `input` を省略すると、前ステップの結果JSONファイルをそのまま連結したものがプロンプトになる。

例:

```
以下の結果を比較してください。
{{range .Inputs}}- {{.joke}}
{{end}}
環境: {{.Env.ENV}}
```

## 出力

- 最終ステップの出力(複数 run の場合は連結)を `output` に書き出す。
- 中間ステップの各結果は `<tempdir>/<実行ID>/<ステップID>/<UUID>.json` に保存される(ID はすべて UUIDv7)。

```json
{
    "order": 0,
    "created_at": "2026-10-18T04:00:00Z",
    "completed_at": "2026-10-18T04:00:12Z",
    "run": { "type": "claude", "args": ["--model", "claude-opus-5-5"] },
    "response": "..."
}
```

- 進捗ログは標準エラーに出力される。

## テスト

```sh
go test ./...
```
