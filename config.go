package orchestrator

// InputConfig contains the necessary values for coniguring an Input,
// such as how to connect to the input source, and the operations the
// input supports
type InputConfig struct {
	Name             string      `toml:"name"`
	Type             string      `toml:"type"`
	ConnectionString string      `toml:"connection_string"`
	Operations       []Operation `toml:"operation"`
}

// ID returns a (hopefully) unique value for this InputConfig
func (ic InputConfig) ID() string {
	return ic.Name
}

// ProcessConfig contains configuration options for processes, including
// an unkeyed map[string]string for arbitrary values
type ProcessConfig struct {
	Name             string            `toml:"name"`
	Type             string            `toml:"type"`
	ExecutionContext map[string]string `toml:"execution_context"`
}

// ID returns a (hopefully) unique value for this ProcessConfig
func (pc ProcessConfig) ID() string {
	return pc.Name
}
