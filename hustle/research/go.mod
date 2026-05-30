module github.com/robertpelloni/hustle/hustle/research

go 1.24

require (
	github.com/jung-kurt/gofpdf v1.16.2
	github.com/robertpelloni/hustle/orchestrator v0.0.0-00010101000000-000000000000
)

require (
	github.com/mattn/go-sqlite3 v1.14.44 // indirect
	golang.org/x/time v0.10.0 // indirect
)

replace github.com/robertpelloni/hustle/orchestrator => ../../orchestrator
