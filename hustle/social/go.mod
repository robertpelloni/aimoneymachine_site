module github.com/robertpelloni/hustle/hustle/social

go 1.24

replace github.com/robertpelloni/hustle/orchestrator => ../../orchestrator

require github.com/robertpelloni/hustle/orchestrator v0.0.0-00010101000000-000000000000

require (
	github.com/mattn/go-sqlite3 v1.14.44 // indirect
	golang.org/x/time v0.10.0 // indirect
)
