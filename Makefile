README.md: go.* *.go
	goreadme -badge-godoc -badge-goreportcard -import-path github.com/dapper-data/dapper-orchestrator > $@
