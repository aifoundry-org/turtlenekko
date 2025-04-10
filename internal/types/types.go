package types

// ParameterConfig represents a parameter configuration with attributes
type ParameterConfig struct {
	Values []string `json:"values" yaml:"values"`
	Output bool     `json:"output,omitempty" yaml:"output,omitempty"`
}
