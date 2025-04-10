package config

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/aifoundry-org/turtlenekko/internal/types"
	"gopkg.in/yaml.v3"
)

// Config represents the benchmark configuration
type Config struct {
	Driver string                           `yaml:"driver"`
	Matrix map[string]types.ParameterConfig `yaml:"matrix"`
}

// Load loads the configuration from a YAML file
func Load(path string) (*Config, error) {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		slog.Warn("Configuration file not found, printing example config", "path", path)
		PrintDefaultConfig()
		return nil, fmt.Errorf("configuration file not found: %s", path)
	}

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading configuration file: %v", err)
	}

	// Parse YAML
	// First try to parse with a flexible format that can handle both simple arrays and objects
	var flexConfig struct {
		Driver string                 `yaml:"driver"`
		Matrix map[string]interface{} `yaml:"matrix"`
	}

	if err := yaml.Unmarshal(data, &flexConfig); err != nil {
		return nil, fmt.Errorf("error parsing configuration file: %v", err)
	}

	// Debug log the parsed config
	slog.Debug("Parsed configuration", "driver", flexConfig.Driver)
	for k, v := range flexConfig.Matrix {
		slog.Debug("Matrix parameter", "key", k, "type", fmt.Sprintf("%T", v), "value", fmt.Sprintf("%v", v))
	}

	// Create the final config
	config := &Config{
		Driver: flexConfig.Driver,
		Matrix: make(map[string]types.ParameterConfig),
	}

	// Process each parameter in the matrix
	for key, value := range flexConfig.Matrix {
		switch v := value.(type) {
		case map[string]interface{}: // Object with attributes
			paramConfig := types.ParameterConfig{
				Output: true, // Default to true
			}

			// Handle the case where we have a list of objects with values
			if valuesArray, ok := v["values"].([]interface{}); ok {
				// This is the format: key: { values: [...], output: bool }
				strValues := make([]string, len(valuesArray))
				for i, val := range valuesArray {
					if str, ok := val.(string); ok {
						strValues[i] = str
					} else {
						strValues[i] = fmt.Sprintf("%v", val)
					}
				}
				paramConfig.Values = strValues

				// Extract output flag if present
				if output, ok := v["output"].(bool); ok {
					paramConfig.Output = output
				}

				config.Matrix[key] = paramConfig
			}

		case []interface{}: // Simple array of values
			// Convert to []string
			strValues := make([]string, len(v))
			for i, val := range v {
				if str, ok := val.(string); ok {
					strValues[i] = str
				} else {
					strValues[i] = fmt.Sprintf("%v", val)
				}
			}

			// Create parameter config with default attributes
			config.Matrix[key] = types.ParameterConfig{
				Values: strValues,
				Output: true, // Default to true
			}

		default:
			return nil, fmt.Errorf("invalid parameter format for key %s: %v", key, value)
		}
	}

	// Debug log the processed config
	slog.Debug("Processed configuration", "driver", config.Driver)
	for k, v := range config.Matrix {
		slog.Debug("Processed parameter", "key", k, "values", v.Values, "output", v.Output)
	}

	return config, nil
}
