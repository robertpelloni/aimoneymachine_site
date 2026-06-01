module github.com/robertpelloni/hustle/hustle/social

go 1.24.3

require (
	github.com/robertpelloni/hustle/orchestrator v0.0.0
	golang.org/x/oauth2 v0.24.0
)

require (
	github.com/mattn/go-sqlite3 v1.14.44 // indirect
	golang.org/x/time v0.10.0 // indirect
)

replace github.com/robertpelloni/hustle/orchestrator => ../../orchestrator
