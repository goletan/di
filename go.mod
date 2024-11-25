module github.com/goletan/di

go 1.23

require (
	github.com/goletan/observability v0.0.0-00010101000000-000000000000
	go.uber.org/zap v1.27.0
)

require go.uber.org/multierr v1.10.0 // indirect

replace github.com/goletan/observability => github.com/goletan/observability v0.0.0-20241125134743-9c25602a6b25
