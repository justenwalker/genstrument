module genstrument/example

go 1.23.0

require (
	github.com/go-logr/stdr v1.2.2
	github.com/justenwalker/genstrument v0.0.0
	github.com/justenwalker/genstrument/genstrument v0.0.0
	go.opentelemetry.io/otel v1.30.0
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.30.0
	go.opentelemetry.io/otel/sdk v1.30.0
	go.opentelemetry.io/otel/trace v1.30.0
)

require (
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	go.opentelemetry.io/otel/metric v1.30.0 // indirect
	golang.org/x/mod v0.21.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/sys v0.25.0 // indirect
	golang.org/x/tools v0.25.0 // indirect
)

replace github.com/justenwalker/genstrument => ../

replace github.com/justenwalker/genstrument/genstrument => ../genstrument
