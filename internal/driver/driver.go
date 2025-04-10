package driver

import (
	"fmt"
)

// Model represents an LLM model
type Model struct {
	Name string
}

// Driver represents an interface for managing LLM runtime environments
// before and after benchmarking
type Driver interface {
	// Setup prepares the LLM runtime environment with the given parameters
	// Returns an error if setup fails
	Setup(params map[string]interface{}) error

	// Teardown cleans up the LLM runtime environment
	// Returns an error if teardown fails
	Teardown() error

	// GetURL returns the URL to connect to the LLM service
	GetURL() string

	// GetModel returns the model information
	GetModel() Model

}

// NewDriver creates a new driver instance based on the driver type
func NewDriver(driverType string) (Driver, error) {
	switch driverType {
	case "dummy":
		return NewDummyDriver(), nil
	case "local_cmd":
		return NewLocalCmdDriver(), nil
	default:
		return nil, fmt.Errorf("unsupported driver type: %s", driverType)
	}
}
