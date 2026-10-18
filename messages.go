package main

const (
	errUnsupportedCLI = "%q は未対応"
	errCLITimeout     = "%s: タイムアウト"
	errCLIFailed      = "%s: %w: %s"
	errCLINotJSON     = "%s: 応答がJSONではない"

	errInvalidTimeout = "timeout: %w"

	errStepsRequired       = "steps: 1つ以上必要"
	errRunRequired         = "%s.run: 1つ以上必要"
	errTypeRequired        = "%s.run[%d].type: 必須"
	errTypeUnsupported     = "%s.run[%d].type: %q は未対応(対応: %s)"
	errFirstInputRequired  = "%s.input: 最初のstepでは必須"
	errRefInFirstStep      = ".%s: 最初のstepでは前ステップ出力を参照できない"
	errRefSingleOfMultiple = ".%s: 前ステップの出力が%d個あるため単一の値として参照できない(.%ss を使う)"
	errEnvUndefined        = ".Env.%s: env で定義されていない"
	errUnknownRef          = ".%s: 未知の参照(使えるのは .Input .Inputs .Response .Responses .Env)"

	errStepFailed      = "step %d: %w"
	errRunFailed       = "step %d run %d: %w"
	errResponseNotJSON = "前ステップの出力[%d]をJSONとして解釈できない: %w"
	errNoExpected      = "%w(前ステップに expected がない)"

	errFileExists = "%s は既に存在する"
)
