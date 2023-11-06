README.md: go.* *.go
	goreadme -badge-godoc -badge-goreportcard -import-path github.com/jspc/pipelines-orchestrator -types > $@
