package formatter

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"

	"github.com/aifoundry-org/turtlenekko/internal/benchmark"
)

// JsonResult represents a benchmark result in JSON format
type JsonResult struct {
	Params                             map[string]string `json:"params"`
	ShortContextPromptTokensPerSec     float64           `json:"short_context_prompt_tokens_per_sec"`
	ShortContextCompletionTokensPerSec float64           `json:"short_context_completion_tokens_per_sec"`
	ShortContextRSquared               float64           `json:"short_context_r_squared"`

	LongContextPromptTokensPerSec     float64 `json:"long_context_prompt_tokens_per_sec"`
	LongContextCompletionTokensPerSec float64 `json:"long_context_completion_tokens_per_sec"`
	LongContextRSquared               float64 `json:"long_context_r_squared"`

	LocalScore *float64 `json:"localscore_estimate,omitempty"`

	Error string `json:"error,omitempty"`
}

// FormatJSON formats benchmark results as JSON and prints to stdout
func FormatJSON(matrixResults []benchmark.MatrixResult, showLocalScore bool) error {
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
			// Short context metrics
			if matrixResult.ShortContextModelFit != nil {
				shortPromptRate := matrixResult.ShortContextModelFit.PromptRate
				if shortPromptRate > 0 {
					result.ShortContextPromptTokensPerSec = math.Round((1000.0/shortPromptRate)*100) / 100
				}

				shortCompletionRate := matrixResult.ShortContextModelFit.CompletionRate
				if shortCompletionRate > 0 {
					result.ShortContextCompletionTokensPerSec = math.Round((1000.0/shortCompletionRate)*100) / 100
				}

				result.ShortContextRSquared = math.Round(matrixResult.ShortContextModelFit.RSquared*100) / 100
			}

			// Long context metrics
			if matrixResult.LongContextModelFit != nil {
				longPromptRate := matrixResult.LongContextModelFit.PromptRate
				if longPromptRate > 0 {
					result.LongContextPromptTokensPerSec = math.Round((1000.0/longPromptRate)*100) / 100
				}

				longCompletionRate := matrixResult.LongContextModelFit.CompletionRate
				if longCompletionRate > 0 {
					result.LongContextCompletionTokensPerSec = math.Round((1000.0/longCompletionRate)*100) / 100
				}

				result.LongContextRSquared = math.Round(matrixResult.LongContextModelFit.RSquared*100) / 100
			}

			// Include LocalScore if enabled and available
			if showLocalScore && matrixResult.LocalScore != nil {
				result.LocalScore = matrixResult.LocalScore
			}
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
func FormatText(matrixResults []benchmark.MatrixResult, showLocalScore bool) {
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

		// Print short context results
		fmt.Printf("\nShort Context Results:\n")
		if matrixResult.ShortContextModelFit != nil {
			shortPromptRate := matrixResult.ShortContextModelFit.PromptRate
			shortCompletionRate := matrixResult.ShortContextModelFit.CompletionRate

			if shortPromptRate > 0 {
				fmt.Printf("  Prompt processing: %.2f tokens/sec\n",
					math.Round((1000.0/shortPromptRate)*100)/100)
			} else {
				fmt.Printf("  Prompt processing: No data\n")
			}

			if shortCompletionRate > 0 {
				fmt.Printf("  Completion generation: %.2f tokens/sec\n",
					math.Round((1000.0/shortCompletionRate)*100)/100)
			} else {
				fmt.Printf("  Completion generation: No data\n")
			}

			fmt.Printf("  Model fit quality (R²): %.2f\n", math.Round(matrixResult.ShortContextModelFit.RSquared*100)/100)

		} else {
			fmt.Printf("  No short context data available\n")
		}

		// Print long context results
		fmt.Printf("\nLong Context Results:\n")
		if matrixResult.LongContextModelFit != nil {
			longPromptRate := matrixResult.LongContextModelFit.PromptRate
			longCompletionRate := matrixResult.LongContextModelFit.CompletionRate

			if longPromptRate > 0 {
				fmt.Printf("  Prompt processing: %.2f tokens/sec\n",
					math.Round((1000.0/longPromptRate)*100)/100)
			} else {
				fmt.Printf("  Prompt processing: No data\n")
			}

			if longCompletionRate > 0 {
				fmt.Printf("  Completion generation: %.2f tokens/sec\n",
					math.Round((1000.0/longCompletionRate)*100)/100)
			} else {
				fmt.Printf("  Completion generation: No data\n")
			}

			fmt.Printf("  Model fit quality (R²): %.2f\n", math.Round(matrixResult.LongContextModelFit.RSquared*100)/100)

			if showLocalScore && matrixResult.LocalScore != nil {
				fmt.Printf("\nLocalscore Estimate: %.2f\n", *matrixResult.LocalScore)
			}

			fmt.Printf("\n")
		} else {
			fmt.Printf("  No long context data available\n\n")
		}
	}
}

// FormatCSV formats benchmark results as CSV and prints to stdout
func FormatCSV(matrixResults []benchmark.MatrixResult, showLocalScore bool) {
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
	header := "short_context_prompt_tokens_per_sec," +
		"short_context_completion_tokens_per_sec,short_context_r_squared," +
		"long_context_prompt_tokens_per_sec," +
		"long_context_completion_tokens_per_sec,long_context_r_squared"

	if showLocalScore {
		header += ",localscore_estimate"
	}

	fmt.Println(header)

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

		// Short context metrics
		shortPromptRateTokensPerSec := 0.0
		shortCompletionRateTokensPerSec := 0.0
		shortRSquared := 0.0

		if result.ShortContextModelFit != nil {
			if result.ShortContextModelFit.PromptRate > 0 {
				shortPromptRateTokensPerSec = math.Round((1000.0/result.ShortContextModelFit.PromptRate)*100) / 100
			}

			if result.ShortContextModelFit.CompletionRate > 0 {
				shortCompletionRateTokensPerSec = math.Round((1000.0/result.ShortContextModelFit.CompletionRate)*100) / 100
			}

			shortRSquared = math.Round(result.ShortContextModelFit.RSquared*100) / 100
		}

		// Long context metrics
		longPromptRateTokensPerSec := 0.0
		longCompletionRateTokensPerSec := 0.0
		longRSquared := 0.0

		if result.LongContextModelFit != nil {
			if result.LongContextModelFit.PromptRate > 0 {
				longPromptRateTokensPerSec = math.Round((1000.0/result.LongContextModelFit.PromptRate)*100) / 100
			}

			if result.LongContextModelFit.CompletionRate > 0 {
				longCompletionRateTokensPerSec = math.Round((1000.0/result.LongContextModelFit.CompletionRate)*100) / 100
			}

			longRSquared = math.Round(result.LongContextModelFit.RSquared*100) / 100
		}

		// Format the output
		output := fmt.Sprintf("%.2f,%.2f,%.2f,%.2f,%.2f,%.2f",
			shortPromptRateTokensPerSec,
			shortCompletionRateTokensPerSec,
			shortRSquared,
			longPromptRateTokensPerSec,
			longCompletionRateTokensPerSec,
			longRSquared)

		// Add LocalScore if enabled and available
		if showLocalScore {
			if result.LocalScore != nil {
				output += fmt.Sprintf(",%.2f", *result.LocalScore)
			} else {
				output += ","
			}
		}

		fmt.Println(output)
	}
}

// WriteToFile writes detailed benchmark results to a log file
func WriteToFile(file *os.File, matrixResults []benchmark.MatrixResult, showLocalScore bool) {
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

		// Print short context results
		fmt.Fprintf(file, "\nShort Context Results:\n")
		if matrixResult.ShortContextModelFit != nil {
			shortPromptRate := matrixResult.ShortContextModelFit.PromptRate
			shortCompletionRate := matrixResult.ShortContextModelFit.CompletionRate

			if shortPromptRate > 0 {
				fmt.Fprintf(file, "  Prompt processing: %.2f tokens/sec\n",
					math.Round((1000.0/shortPromptRate)*100)/100)
			} else {
				fmt.Fprintf(file, "  Prompt processing: No data\n")
			}

			if shortCompletionRate > 0 {
				fmt.Fprintf(file, "  Completion generation: %.2f tokens/sec\n",
					math.Round((1000.0/shortCompletionRate)*100)/100)
			} else {
				fmt.Fprintf(file, "  Completion generation: No data\n")
			}

			fmt.Fprintf(file, "  Model fit quality (R²): %.2f\n", math.Round(matrixResult.ShortContextModelFit.RSquared*100)/100)

		} else {
			fmt.Fprintf(file, "  No short context data available\n")
		}

		// Print long context results
		fmt.Fprintf(file, "\nLong Context Results:\n")
		if matrixResult.LongContextModelFit != nil {
			longPromptRate := matrixResult.LongContextModelFit.PromptRate
			longCompletionRate := matrixResult.LongContextModelFit.CompletionRate

			if longPromptRate > 0 {
				fmt.Fprintf(file, "  Prompt processing: %.2f tokens/sec\n",
					math.Round((1000.0/longPromptRate)*100)/100)
			} else {
				fmt.Fprintf(file, "  Prompt processing: No data\n")
			}

			if longCompletionRate > 0 {
				fmt.Fprintf(file, "  Completion generation: %.2f tokens/sec\n",
					math.Round((1000.0/longCompletionRate)*100)/100)
			} else {
				fmt.Fprintf(file, "  Completion generation: No data\n")
			}

			fmt.Fprintf(file, "  Model fit quality (R²): %.2f\n", math.Round(matrixResult.LongContextModelFit.RSquared*100)/100)

			fmt.Fprintf(file, "\n")
		} else {
			fmt.Fprintf(file, "  No long context data available\n\n")
		}

		// Print CSV header
		fmt.Fprintf(file, "context,prompt_tokens,completion_tokens,response_time_ms\n")

		// Print results as CSV
		for _, result := range matrixResult.Results {
			if result == nil {
				continue // Skip failed results
			}

			// Convert response time to milliseconds
			responseTimeMs := result.ResponseTime.Milliseconds()

			// Determine if this is a short or long context result
			contextType := "short"
			if result.PromptTokens > 1000 {
				contextType = "long"
			}

			// Output as CSV
			fmt.Fprintf(file, "%s,%d,%d,%d\n",
				contextType,
				result.PromptTokens,
				result.CompletionTokens,
				responseTimeMs)
		}
	}
}
