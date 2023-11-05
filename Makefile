README.md: go.* *.go
	goreadme -badge-godoc -badge-goreportcard -import-path github.com/jspc/pipelines-orchestrator -constants -factories -functions -methods -types -variabless > $@
