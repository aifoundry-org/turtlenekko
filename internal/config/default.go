package config

import (
	"fmt"
	"os"
)

const defaultConfigTemplate = `# Turtlenekko Configuration File
# This is an example configuration file with common settings

# Driver configuration
# Available drivers: "dummy", "local_cmd"
driver: "dummy"

# Matrix of parameters to test
# Each parameter can be specified as:
# 1. A simple array of values: param: [value1, value2]
# 2. An object with values and output flag: param: {values: [value1, value2], output: true}
matrix:
  # URL to the LLM API endpoint
  url: 
    values: ["http://localhost:8080/v1/chat/completions"]
    output: true
  
  # Model to use for benchmarking
  model: 
    values: ["llama3"]
    output: true
  
# Example configuration for local_cmd driver
# Uncomment and modify as needed
#
# driver: "local_cmd"
# matrix:
#   url: 
#     values: ["http://localhost:8080/v1/chat/completions"]
#     output: false
#   model: 
#     values: ["llama3"]
#     output: true
#   # Command to run before benchmarking (supports Go templates)
#   setup_cmd:
#     values: ["docker run -d --name llm-server -p 8080:8080 llm-image:latest"]
#     output: false
#   # Command to run after benchmarking (supports Go templates)
#   teardown_cmd:
#     values: ["docker stop llm-server && docker rm llm-server"]
#     output: false`

// WriteDefaultConfig writes the default configuration to the specified path
func WriteDefaultConfig(path string) error {
	return os.WriteFile(path, []byte(defaultConfigTemplate), 0644)
}

// PrintDefaultConfig prints the default configuration to stdout
func PrintDefaultConfig() {
	fmt.Println(defaultConfigTemplate)
}
