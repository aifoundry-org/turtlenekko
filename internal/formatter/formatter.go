package formatter

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/aifoundry-org/turtlenekko/internal/benchmark"
	// "github.com/aifoundry-org/turtlenekko/internal/types"
)

// JsonResult represents a benchmark result in JSON format
type JsonResult struct {
	Params                     map[string]string `json:"params"`
	PromptRateMs               float64           `json:"prompt_rate_ms"`
	PromptRateTokensPerSec     float64           `json:"prompt_rate_tokens_per_sec"`
	CompletionRateMs           float64           `json:"completion_rate_ms"`
	CompletionRateTokensPerSec float64           `json:"completion_rate_tokens_per_sec"`
	RSquared                   float64           `json:"r_squared"`
	Error                      string            `json:"error,omitempty"`
}

// FormatJSON formats benchmark results as JSON and prints to stdout
func FormatJSON(matrixResults []benchmark.MatrixResult) error {
	var jsonResults []JsonResult

	for _, matrixResult := range matrixResults {
		// Filter parameters based on output flags
		filteredParams := make(map[string]string)
		for k, v := range matrixResult.Params {
			if outputFlag, exists := matrixResult.OutputFlags[k]; exists && outputFlag {
				filteredParams[k] = v
			}
		}

		result := JsonResult{
			Params: filteredParams,
		}

		if matrixResult.Error != nil {
			result.Error = matrixResult.Error.Error()
		} else {
			result.PromptRateMs = matrixResult.ModelFit.PromptRate
			result.PromptRateTokensPerSec = 1000.0 / matrixResult.ModelFit.PromptRate
			result.CompletionRateMs = matrixResult.ModelFit.CompletionRate
			result.CompletionRateTokensPerSec = 1000.0 / matrixResult.ModelFit.CompletionRate
			result.RSquared = matrixResult.ModelFit.RSquared
		}

		jsonResults = append(jsonResults, result)
	}

	// Marshal to JSON
	jsonData, err := json.MarshalIndent(jsonResults, "", "  ")
	if err != nil {
		return fmt.Errorf("error creating JSON output: %v", err)
	}

	fmt.Println(string(jsonData))
	return nil
}

// FormatText formats benchmark results as human-readable text and prints to stdout
func FormatText(matrixResults []benchmark.MatrixResult) {
	for i, matrixResult := range matrixResults {
		// Output to console
		fmt.Printf("\n=== Matrix Combination %d ===\n", i+1)

		// Print parameters used
		fmt.Println("Parameters:")
		for k, v := range matrixResult.Params {
			if outputFlag, exists := matrixResult.OutputFlags[k]; exists && outputFlag {
				fmt.Printf("  %s: %s\n", k, v)
			}
		}

		if matrixResult.Error != nil {
			fmt.Printf("Error: %v\n", matrixResult.Error)
			continue
		}

		// Print model fit results
		fmt.Printf("\nModel fit results:\n")

		fmt.Printf("Prompt processing rate: %.2f ms/token (%.2f tokens/sec)\n",
			matrixResult.ModelFit.PromptRate, 1000.0/matrixResult.ModelFit.PromptRate)

		fmt.Printf("Completion generation rate: %.2f ms/token (%.2f tokens/sec)\n",
			matrixResult.ModelFit.CompletionRate, 1000.0/matrixResult.ModelFit.CompletionRate)

		fmt.Printf("R-squared (goodness of fit): %.4f\n\n", matrixResult.ModelFit.RSquared)
	}
}

// FormatCSV formats benchmark results as CSV and prints to stdout
func FormatCSV(matrixResults []benchmark.MatrixResult) {
	// Get all unique parameter keys with output:true
	paramKeys := make(map[string]bool)
	for _, result := range matrixResults {
		for k, outputFlag := range result.OutputFlags {
			if outputFlag {
				paramKeys[k] = true
			}
		}
	}

	// Convert map to sorted slice for consistent output
	var sortedParamKeys []string
	for k := range paramKeys {
		sortedParamKeys = append(sortedParamKeys, k)
	}
	sort.Strings(sortedParamKeys)

	// Print CSV header
	// First the parameter columns
	for i, key := range sortedParamKeys {
		if i > 0 {
			fmt.Print(",")
		}
		fmt.Print(key)
	}
	
	// Then the metrics columns
	if len(sortedParamKeys) > 0 {
		fmt.Print(",")
	}
	fmt.Println("prompt_rate_ms,prompt_rate_tokens_per_sec,completion_rate_ms,completion_rate_tokens_per_sec,r_squared")

	// Print each result row
	for _, result := range matrixResults {
		if result.Error != nil {
			continue // Skip rows with errors
		}

		// Print parameter values
		for i, key := range sortedParamKeys {
			if i > 0 {
				fmt.Print(",")
			}
			// Get parameter value, empty string if not found
			value := ""
			if v, ok := result.Params[key]; ok {
				value = v
			}
			fmt.Print(value)
		}

		// Print metrics
		if len(sortedParamKeys) > 0 {
			fmt.Print(",")
		}
		
		promptRate := result.ModelFit.PromptRate
		promptRateTokensPerSec := 1000.0 / promptRate
		completionRate := result.ModelFit.CompletionRate
		completionRateTokensPerSec := 1000.0 / completionRate
		rSquared := result.ModelFit.RSquared

		fmt.Printf("%.2f,%.2f,%.2f,%.2f,%.4f\n",
			promptRate,
			promptRateTokensPerSec,
			completionRate,
			completionRateTokensPerSec,
			rSquared)
	}
}

// WriteToFile writes detailed benchmark results to a log file
func WriteToFile(file *os.File, matrixResults []benchmark.MatrixResult) {
	for i, matrixResult := range matrixResults {
		// Output to log file
		fmt.Fprintf(file, "\n=== Matrix Combination %d ===\n", i+1)

		// Print parameters used
		fmt.Fprintf(file, "Parameters:\n")
		for k, v := range matrixResult.Params {
			if outputFlag, exists := matrixResult.OutputFlags[k]; exists && outputFlag {
				fmt.Fprintf(file, "  %s: %s\n", k, v)
			}
		}

		if matrixResult.Error != nil {
			fmt.Fprintf(file, "Error: %v\n", matrixResult.Error)
			continue
		}

		// Print model fit results
		fmt.Fprintf(file, "\nModel fit results:\n")

		fmt.Fprintf(file, "Prompt processing rate: %.2f ms/token (%.2f tokens/sec)\n",
			matrixResult.ModelFit.PromptRate, 1000.0/matrixResult.ModelFit.PromptRate)

		fmt.Fprintf(file, "Completion generation rate: %.2f ms/token (%.2f tokens/sec)\n",
			matrixResult.ModelFit.CompletionRate, 1000.0/matrixResult.ModelFit.CompletionRate)

		fmt.Fprintf(file, "R-squared (goodness of fit): %.4f\n\n", matrixResult.ModelFit.RSquared)

		// Print CSV header
		fmt.Fprintf(file, "prompt_tokens,completion_tokens,response_time_ms\n")

		// Print results as CSV
		for _, result := range matrixResult.Results {
			if result == nil {
				continue // Skip failed results
			}

			// Convert response time to milliseconds
			responseTimeMs := result.ResponseTime.Milliseconds()

			// Output as CSV
			fmt.Fprintf(file, "%d,%d,%d\n",
				result.PromptTokens,
				result.CompletionTokens,
				responseTimeMs)
		}
	}
}
