package driver

import (
	"log/slog"
)

// DummyDriver implements the Driver interface without actually setting up any environment
// It just returns the configured URL and model name
type DummyDriver struct {
	url   string
	model Model
}

// NewDummyDriver creates a new DummyDriver instance
func NewDummyDriver() *DummyDriver {
	return &DummyDriver{
		url:   "",
		model: Model{Name: ""},
	}
}

// Setup stores the URL and model name from parameters
func (d *DummyDriver) Setup(params map[string]interface{}) error {
	// Extract URL if provided
	if url, ok := params["url"].(string); ok {
		d.url = url
		slog.Info("Setting URL", "component", "dummy", "url", d.url)
	}
	
	// Extract model if provided
	if modelName, ok := params["model"].(string); ok {
		d.model.Name = modelName
		slog.Info("Setting model", "component", "dummy", "model", d.model.Name)
	}
	
	slog.Info("Dummy driver setup completed", "component", "dummy")
	return nil
}

// Teardown does nothing for the dummy driver
func (d *DummyDriver) Teardown() error {
	slog.Info("Dummy driver teardown completed (no-op)", "component", "dummy")
	return nil
}

// GetURL returns the configured URL
func (d *DummyDriver) GetURL() string {
	return d.url
}

// GetModel returns the configured model
func (d *DummyDriver) GetModel() Model {
	return d.model
}
