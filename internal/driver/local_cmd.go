package driver

import (
	"bytes"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
	"text/template"
)

// LocalCmdDriver implements the Driver interface for running local shell commands
type LocalCmdDriver struct {
	url         string
	model       Model
	setupCmd    string
	teardownCmd string
	params      map[string]interface{}
}

// NewLocalCmdDriver creates a new LocalCmdDriver instance
func NewLocalCmdDriver() *LocalCmdDriver {
	return &LocalCmdDriver{
		model:  Model{Name: ""},
		url:    "",
		params: make(map[string]interface{}),
	}
}

// interpolateCommand replaces template variables in the command string with parameter values
func (d *LocalCmdDriver) interpolateCommand(cmdTemplate string) (string, error) {
	tmpl, err := template.New("command").Parse(cmdTemplate)
	if err != nil {
		return "", fmt.Errorf("invalid command template: %v", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, d.params); err != nil {
		return "", fmt.Errorf("error interpolating command: %v", err)
	}

	return buf.String(), nil
}

// Setup prepares the environment by running the setup command
func (d *LocalCmdDriver) Setup(params map[string]interface{}) error {
	// Store all parameters for interpolation
	d.params = params

	// Extract URL if provided
	if url, ok := params["url"].(string); ok && url != "" {
		d.url = url
	}

	// Extract model if provided
	if modelName, ok := params["model"].(string); ok && modelName != "" {
		d.model.Name = modelName
	}

	// Extract setup command
	setupCmd, ok := params["setup_cmd"].(string)
	if !ok || setupCmd == "" {
		return nil // No setup command, nothing to do
	}
	d.setupCmd = setupCmd

	// Extract teardown command
	if teardownCmd, ok := params["teardown_cmd"].(string); ok {
		d.teardownCmd = teardownCmd
	}

	// Interpolate and run the setup command
	cmd, err := d.interpolateCommand(setupCmd)
	if err != nil {
		return fmt.Errorf("failed to prepare setup command: %v", err)
	}

	// Log the command
	slog.Info("Running setup command", "component", "local_cmd", "command", cmd)

	// Run the command
	shellCmd := exec.Command("sh", "-c", cmd)
	output, err := shellCmd.CombinedOutput()
	if err != nil {
		slog.Error("Setup command failed", "component", "local_cmd", "error", err, "output", string(output))
		return fmt.Errorf("setup command failed: %v, output: %s", err, output)
	}

	slog.Info("Setup command completed successfully", "component", "local_cmd")

	// If the command output contains a URL, use it
	outputStr := string(output)
	outputStr = strings.TrimSpace(outputStr)
	if strings.HasPrefix(outputStr, "http://") || strings.HasPrefix(outputStr, "https://") {
		d.url = outputStr
	}

	return nil
}

// GetURL returns the URL to connect to the service
func (d *LocalCmdDriver) GetURL() string {
	return d.url
}

// GetModel returns the model information
func (d *LocalCmdDriver) GetModel() Model {
	return d.model
}

// Teardown cleans up by running the teardown command
func (d *LocalCmdDriver) Teardown() error {
	if d.teardownCmd == "" {
		return nil // No teardown command, nothing to do
	}

	// Interpolate and run the teardown command
	cmd, err := d.interpolateCommand(d.teardownCmd)
	if err != nil {
		return fmt.Errorf("failed to prepare teardown command: %v", err)
	}

	// Log the command
	slog.Info("Running teardown command", "component", "local_cmd", "command", cmd)

	// Run the command
	shellCmd := exec.Command("sh", "-c", cmd)
	output, err := shellCmd.CombinedOutput()
	if err != nil {
		slog.Error("Teardown command failed", "component", "local_cmd", "error", err, "output", string(output))
		return fmt.Errorf("teardown command failed: %v, output: %s", err, output)
	}

	slog.Info("Teardown command completed successfully", "component", "local_cmd")

	return nil
}
