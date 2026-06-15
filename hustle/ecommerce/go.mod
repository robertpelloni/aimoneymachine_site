module github.com/robertpelloni/hustle/hustle/ecommerce

go 1.24.0

require (
	github.com/robertpelloni/hustle/hustle/research v0.0.0-00010101000000-000000000000
	github.com/robertpelloni/hustle/orchestrator v0.0.0-00010101000000-000000000000
)

replace github.com/robertpelloni/hustle/hustle/research => ../research

replace github.com/robertpelloni/hustle/orchestrator => ../../orchestrator
