package benchmark

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"math/rand"
	"net/http"
	"time"

	"github.com/aifoundry-org/turtlenekko/internal/driver"
	"github.com/aifoundry-org/turtlenekko/internal/types"
)

// ChatMessage represents a message in the chat completion API
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionParams contains parameters for a chat completion request
type ChatCompletionParams struct {
	Messages            []ChatMessage
	Temperature         float64
	TopP                float64
	MaxCompletionTokens int
	Seed                int
}

// ChatCompletionRequest represents the request body for chat completion
type ChatCompletionRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
	TopP        float64       `json:"top_p,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Seed        int           `json:"seed,omitempty"`
}

// ChatCompletionResponse represents the response from chat completion API
type ChatCompletionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	Choices []struct {
		Message ChatMessage `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// CompletionResult contains token usage information and timing from the LLM response
type CompletionResult struct {
	PromptTokens     int
	CompletionTokens int
	ResponseTime     time.Duration
}

// Result represents the benchmark results
type Result struct {
	URL          string
	ResponseTime time.Duration
	Success      bool
	Result       *CompletionResult
	PromptLength int
}

// Benchmark represents a benchmark runner
type Benchmark struct {
	URL     string
	Model   string
	Timeout time.Duration
	Client  *http.Client
	Driver  driver.Driver
}

// NewBenchmark creates a new benchmark runner
func NewBenchmark(url string, model string, driverType string) *Benchmark {
	if model == "" {
		model = "llama" // Default model if none provided
	}

	// Default timeout of 600 seconds
	timeout := 600 * time.Second

	// Create driver if driver type is specified
	var d driver.Driver
	if driverType != "" {
		var err error
		d, err = driver.NewDriver(driverType)
		if err != nil {
			slog.Warn("Failed to create driver", "error", err)
		}
	}

	return &Benchmark{
		URL:     url,
		Model:   model,
		Timeout: timeout,
		Client: &http.Client{
			Timeout: timeout,
		},
		Driver: d,
	}
}

// ChatCompletion sends a chat completion request to the LLM
func (b *Benchmark) ChatCompletion(params ChatCompletionParams) (*CompletionResult, error) {
	// Create request body
	requestBody := ChatCompletionRequest{
		Model:       b.Model,
		Messages:    params.Messages,
		Temperature: params.Temperature,
		TopP:        params.TopP,
		MaxTokens:   params.MaxCompletionTokens,
		Seed:        params.Seed,
	}

	// Marshal request to JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}

	slog.Info("Sending request", "component", "benchmark", "url", b.URL)

	// Create HTTP request
	req, err := http.NewRequest("POST", b.URL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Start timing right before the API call
	startTime := time.Now()

	// Send request
	resp, err := b.Client.Do(req)

	// Stop timing right after receiving the response
	responseTime := time.Since(startTime)

	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		slog.Error("Received error response", "component", "benchmark", "status_code", resp.StatusCode)
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	slog.Info("Received successful response", "component", "benchmark", "status_code", resp.StatusCode)

	// Decode the response
	var response ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		slog.Error("Failed to decode response", "component", "benchmark", "error", err)
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	// Log the completion response content
	if len(response.Choices) > 0 {
		slog.Debug("Response content", "component", "benchmark", "content", response.Choices[0].Message.Content)
	} else {
		slog.Warn("Response contains no choices", "component", "benchmark")
	}

	// Extract usage information and include timing
	result := &CompletionResult{
		PromptTokens:     response.Usage.PromptTokens,
		CompletionTokens: response.Usage.CompletionTokens,
		ResponseTime:     responseTime,
	}

	slog.Info("Completion successful",
		"component", "benchmark",
		"prompt_tokens", result.PromptTokens,
		"completion_tokens", result.CompletionTokens)

	return result, nil
}

// Lorem ipsum text for generating realistic-looking content
// const loremIpsumText = `Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.`
const loremIpsumText = `Lorem ipsum dolor sit amet, consectetur adipiscing elit. Donec risus erat, interdum id magna egestas, sodales malesuada lacus. Nullam at sagittis lacus. Aliquam erat volutpat. Suspendisse sed dolor diam. Nunc ac purus ultrices, aliquet velit et, iaculis mauris. Nullam vitae justo est. Nam id nisi nisl. Pellentesque euismod ut urna a fringilla. Donec dictum, dolor vitae sagittis sollicitudin, dui quam posuere massa, non aliquet mauris justo maximus sapien. Proin suscipit ut turpis quis blandit. Sed sit amet convallis libero. Curabitur sed scelerisque nisi. Pellentesque faucibus commodo convallis. Nulla pellentesque ut turpis eu rutrum. Fusce ligula mi, elementum et dolor sit amet, accumsan eleifend dui. Vivamus vel massa vel nibh interdum euismod et vel elit. Praesent rutrum mi eu eleifend fringilla. Cras venenatis libero ac felis faucibus, et tincidunt est dignissim. Donec condimentum libero ex, at dictum odio maximus eu. Donec at accumsan turpis, at lacinia risus. Orci varius natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Fusce maximus orci diam, eget consequat eros laoreet in. Morbi iaculis tincidunt erat, eget maximus risus mattis a. Donec ut nunc a augue placerat gravida. Fusce vitae eros eget eros maximus cursus at ut dolor. Sed eu finibus nulla. Pellentesque id placerat felis. Mauris at risus bibendum, ultrices felis ac, viverra urna. Orci varius natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Donec lobortis cursus feugiat. Sed fermentum est nec sapien maximus, non lobortis tortor feugiat. Phasellus in molestie risus. Etiam faucibus sapien ex, nec elementum purus faucibus nec. Ut sed massa ornare nunc condimentum tincidunt et et massa. Nam interdum mattis nulla, et interdum nisl sollicitudin vitae. Maecenas eget quam ut tellus rhoncus placerat. Praesent eu felis quis nisi faucibus porta. Maecenas eleifend ultricies faucibus. Sed tempor felis at nulla mollis dignissim. Praesent ac accumsan elit. Maecenas efficitur, nunc a feugiat tristique, urna diam facilisis odio, gravida consectetur risus ex ac dui. Sed laoreet elit et tellus efficitur, id rhoncus risus interdum. In tincidunt porta bibendum. In porta nisl porttitor nisl rutrum, at auctor arcu eleifend. Mauris ac volutpat turpis. Maecenas consequat lectus sit amet nibh posuere, vitae euismod felis tristique. Aliquam imperdiet varius sodales. Aliquam eget mauris in felis elementum facilisis. In efficitur euismod orci porttitor scelerisque. Curabitur imperdiet tellus eros, in varius tellus egestas et. Vivamus auctor ipsum in varius vulputate. In hac habitasse platea dictumst. Vivamus lacinia tellus vel mattis auctor. Vivamus quis condimentum lacus. Sed imperdiet libero ut ipsum tempor, ut consequat quam consectetur. Etiam leo ex, viverra porta diam vitae, molestie imperdiet diam. Fusce a nisl eu arcu rhoncus volutpat. Vestibulum ante ipsum primis in faucibus orci luctus et ultrices posuere cubilia curae; Aenean rutrum rhoncus sem, sed rhoncus leo imperdiet in. Proin a euismod enim. Vivamus elementum ligula quis lacus vehicula fermentum. Aenean venenatis, est ut interdum suscipit, risus nibh molestie purus, a posuere dui sem ac nibh. Donec aliquet diam nec nunc vehicula sollicitudin. Donec feugiat faucibus diam sit amet vulputate. Praesent rhoncus diam ac felis facilisis varius. Fusce vulputate nisl id suscipit venenatis. Mauris fermentum, nisl quis interdum interdum, risus purus posuere libero, quis accumsan turpis magna id tortor. In tempus malesuada est, nec aliquam urna. Suspendisse tempor et orci tempor rutrum. Curabitur sit amet mauris libero. Etiam convallis libero ipsum, eget imperdiet sapien sodales vitae. Praesent quis commodo nisl. Vestibulum accumsan eget metus ut venenatis. Sed pharetra enim gravida nunc condimentum ullamcorper. Aliquam egestas iaculis mi. Donec finibus dapibus ante, nec rutrum diam feugiat et. Etiam pellentesque, nulla et congue porttitor, magna mi efficitur elit, eget congue lorem metus ac ante. Nullam blandit ligula mi, posuere lobortis risus efficitur id. Duis pharetra convallis urna, at efficitur sem vestibulum eu. Cras aliquam, nunc non venenatis lacinia, lacus ipsum luctus mauris, et placerat nibh sapien tempor nibh. Integer aliquet mauris id scelerisque sollicitudin. Etiam ac magna ipsum. Phasellus mattis ipsum et felis maximus consectetur. Proin fringilla vel dui et tempor. Nam rhoncus eu mauris vitae feugiat. Phasellus feugiat laoreet erat sit amet imperdiet. Fusce sodales ex sapien, vitae ultrices purus pretium sed. Suspendisse nec felis consectetur urna fermentum mollis eget dapibus enim. Cras consequat mauris et cursus accumsan. Ut semper rutrum nisl sit amet congue. Mauris nisl magna, lacinia vitae faucibus in, congue et elit. Maecenas ullamcorper nisl id libero sollicitudin lacinia. Praesent ultrices, massa vitae faucibus porta, nunc nibh venenatis lorem, aliquam ultrices augue nibh vitae lorem. Vivamus faucibus augue in dapibus cursus. Sed facilisis lectus convallis mauris venenatis pulvinar. Vivamus nec nibh vitae nisi pretium tristique. Sed nec est non mauris scelerisque aliquet. Duis a est feugiat, efficitur ex rutrum, condimentum arcu. Mauris ullamcorper molestie odio a sagittis. Mauris aliquam arcu vel ipsum lobortis blandit. Integer quis semper justo. Morbi quis consectetur quam. Curabitur vehicula feugiat ligula at venenatis. In et est vitae odio euismod interdum. Cras metus nulla, volutpat a magna vitae, facilisis hendrerit libero. Donec dictum odio et tellus sagittis tristique. Mauris at arcu velit. Vestibulum eu dolor id nulla sodales finibus a et elit. Morbi ultricies et magna ut fringilla. Interdum et malesuada fames ac ante ipsum primis in faucibus. Ut maximus scelerisque nibh, at ultrices magna iaculis vel. Quisque eu est ac arcu malesuada tristique. Vestibulum vestibulum elementum tellus, nec laoreet turpis ornare quis. Donec imperdiet vulputate tincidunt. Curabitur nisl risus, faucibus ut venenatis id, porttitor sit amet augue. Integer molestie iaculis condimentum. Donec varius elit ipsum, sed vestibulum eros finibus lacinia. Vestibulum congue mollis nisi, quis pretium ligula maximus in. Ut tincidunt auctor tincidunt. Nam a convallis erat. Donec dignissim porta cursus. Nam malesuada tempor sem, et cursus tellus. Nulla commodo fringilla tellus dictum dapibus. Sed sed sapien ante. Nullam luctus, neque nec faucibus auctor, erat urna condimentum ipsum, id imperdiet metus nisl eu tellus. Praesent id ante semper, commodo mi nec, placerat ipsum. Quisque mollis porta scelerisque. Cras feugiat, est sed tristique fermentum, diam lorem porta purus, eu semper est sapien ut velit. Vivamus sapien turpis, tincidunt ac mauris vitae, dapibus aliquam urna. Nulla vestibulum egestas felis. Sed ultricies ullamcorper justo eget fermentum. Morbi. `

// generateRandomContent creates a string of random alphanumeric characters of the specified length
func generateRandomContent(length int) string {
	const charset = "0123456789abcdefghijklmnopqrstuvwxyz"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

// generateLoremIpsum creates a string of lorem ipsum text repeated to reach the specified length
func generateLoremIpsum(length int) string {
	if length <= 0 {
		return ""
	}

	// Calculate how many times we need to repeat the text
	repeats := (length + len(loremIpsumText) - 1) / len(loremIpsumText)

	// Build the repeated text
	var result string
	for i := 0; i < repeats; i++ {
		result += loremIpsumText
	}

	// Truncate to the exact requested length
	if len(result) > length {
		result = result[:length]
	}

	return result
}

// generateMessages creates an array of chat messages with random content of specified lengths
func generateMessages(systemContentLength int, postfix string) []ChatMessage {
	// Random prefix prevents kv cache reuse.
	content := "seed:" + generateRandomContent(10) + "\n" + generateLoremIpsum(systemContentLength)
	if postfix != "" {
		content += postfix
	}

	return []ChatMessage{
		{
			Role:    "user",
			Content: content,
		},
	}
}

// RunWithPromptLength executes a benchmark with a specific prompt length and max completion tokens
func (b *Benchmark) RunWithPromptLength(promptLength int, maxCompletionTokens int, postfix string) (*CompletionResult, error) {
	slog.Info("Running benchmark",
		"component", "benchmark",
		"prompt_length", promptLength,
		"max_tokens", maxCompletionTokens,
		"url", b.URL)

	// Parameters with specified prompt length and max tokens
	params := ChatCompletionParams{
		Messages:            generateMessages(promptLength, postfix),
		Temperature:         2.0, // just take random
		TopP:                1.0,
		MaxCompletionTokens: maxCompletionTokens,
		Seed:                -1,
	}

	// Make the actual request to the LLM
	completionResult, err := b.ChatCompletion(params)

	if err != nil {
		slog.Error("Benchmark failed",
			"component", "benchmark",
			"prompt_length", promptLength,
			"max_tokens", maxCompletionTokens,
			"error", err)
		return nil, fmt.Errorf("chat completion failed: %v", err)
	}

	// Log the detailed timing information
	slog.Info("Benchmark completed",
		"component", "benchmark",
		"prompt_length", promptLength,
		"max_tokens", maxCompletionTokens,
		"response_time_ms", completionResult.ResponseTime.Milliseconds())

	return completionResult, nil
}

// ModelFitResult contains the fitted parameters for the completion time model
type ModelFitResult struct {
	PromptRate     float64 // ms per prompt token
	CompletionRate float64 // ms per completion token
	RSquared       float64 // goodness of fit (0-1)
}

// fitCompletionTimeModel fits the model: completion_time = pr * prompt_tokens + cr * completion_tokens
// to the measured data, where pr is prompt rate and cr is completion rate
func fitCompletionTimeModel(results []*CompletionResult) *ModelFitResult {
	if len(results) < 2 {
		return &ModelFitResult{
			PromptRate:     0,
			CompletionRate: 0,
			RSquared:       0,
		}
	}

	// Prepare data for linear regression
	n := len(results)
	sumX1 := 0.0
	sumX2 := 0.0
	sumY := 0.0
	sumX1X1 := 0.0
	sumX2X2 := 0.0
	sumX1X2 := 0.0
	sumX1Y := 0.0
	sumX2Y := 0.0

	for _, r := range results {
		if r == nil {
			continue
		}

		x1 := float64(r.PromptTokens)
		x2 := float64(r.CompletionTokens)
		y := float64(r.ResponseTime.Milliseconds())

		sumX1 += x1
		sumX2 += x2
		sumY += y
		sumX1X1 += x1 * x1
		sumX2X2 += x2 * x2
		sumX1X2 += x1 * x2
		sumX1Y += x1 * y
		sumX2Y += x2 * y
	}

	// Solve the system of linear equations
	// [sumX1X1 sumX1X2] [pr] = [sumX1Y]
	// [sumX1X2 sumX2X2] [cr] = [sumX2Y]

	determinant := sumX1X1*sumX2X2 - sumX1X2*sumX1X2

	if math.Abs(determinant) < 1e-10 {
		// Singular matrix, can't solve
		return &ModelFitResult{
			PromptRate:     0,
			CompletionRate: 0,
			RSquared:       0,
		}
	}

	pr := (sumX2X2*sumX1Y - sumX1X2*sumX2Y) / determinant
	cr := (sumX1X1*sumX2Y - sumX1X2*sumX1Y) / determinant

	// Calculate R-squared
	meanY := sumY / float64(n)
	totalSumSquares := 0.0
	residualSumSquares := 0.0

	for _, r := range results {
		if r == nil {
			continue
		}

		y := float64(r.ResponseTime.Milliseconds())
		yPred := pr*float64(r.PromptTokens) + cr*float64(r.CompletionTokens)

		totalSumSquares += math.Pow(y-meanY, 2)
		residualSumSquares += math.Pow(y-yPred, 2)
	}

	rSquared := 0.0
	if totalSumSquares > 0 {
		rSquared = 1.0 - (residualSumSquares / totalSumSquares)
	}

	return &ModelFitResult{
		PromptRate:     pr,
		CompletionRate: cr,
		RSquared:       rSquared,
	}
}

// RunScalingBenchmark runs benchmarks with increasing prompt sizes and different max tokens
func (b *Benchmark) RunScalingBenchmark(postfix string) ([]*CompletionResult, *ModelFitResult, *ModelFitResult, error) {
	slog.Info("Starting scaling benchmark", "component", "benchmark", "url", b.URL)

	// Run a warmup request to initialize the model
	slog.Info("Running warmup request", "component", "benchmark")
	_, warmupErr := b.RunWithPromptLength(100, 100, "Just a warmup request.")
	if warmupErr != nil {
		slog.Warn("Warmup request failed (continuing with benchmark)", "component", "benchmark", "error", warmupErr)
	} else {
		slog.Info("Warmup request completed successfully", "component", "benchmark")
	}

	// Short context benchmarks
	shortContextPromptLengths := []int{100, 500}
	shortContextMaxTokens := []int{1, 1, 100, 100}

	// Long context benchmarks
	longContextPromptLengths := []int{9000, 10000}
	longContextMaxTokens := []int{1, 1, 100, 100}

	var allResults []*CompletionResult
	var shortContextResults []*CompletionResult
	var longContextResults []*CompletionResult

	// Run short context benchmarks
	slog.Info("Running short context benchmarks", "component", "benchmark")
	for _, promptLength := range shortContextPromptLengths {
		for _, maxTokens := range shortContextMaxTokens {
			result, err := b.RunWithPromptLength(promptLength, maxTokens, postfix)

			if err != nil {
				slog.Error("Short context benchmark failed",
					"component", "benchmark",
					"prompt_length", promptLength,
					"max_tokens", maxTokens,
					"error", err)
			} else {
				// Add successful result
				shortContextResults = append(shortContextResults, result)
				allResults = append(allResults, result)
			}

			// Small delay between requests to avoid overwhelming the server
			time.Sleep(1000 * time.Millisecond)
		}
	}

	// Run long context benchmarks
	slog.Info("Running long context benchmarks", "component", "benchmark")
	for _, promptLength := range longContextPromptLengths {
		for _, maxTokens := range longContextMaxTokens {
			result, err := b.RunWithPromptLength(promptLength, maxTokens, postfix)

			if err != nil {
				slog.Error("Long context benchmark failed",
					"component", "benchmark",
					"prompt_length", promptLength,
					"max_tokens", maxTokens,
					"error", err)
			} else {
				// Add successful result
				longContextResults = append(longContextResults, result)
				allResults = append(allResults, result)
			}

			// Small delay between requests to avoid overwhelming the server
			time.Sleep(1000 * time.Millisecond)
		}
	}

	// Check if all benchmarks failed
	if len(shortContextResults) == 0 && len(longContextResults) == 0 {
		return allResults, nil, nil, fmt.Errorf("all benchmark configurations failed")
	}

	// Create model fit results
	var shortContextModelFit *ModelFitResult
	var longContextModelFit *ModelFitResult

	// Fit the completion time model to the data for short context if we have results
	if len(shortContextResults) > 0 {
		shortContextModelFit = fitCompletionTimeModel(shortContextResults)
	}

	// Fit the completion time model to the data for long context if we have results
	if len(longContextResults) > 0 {
		longContextModelFit = fitCompletionTimeModel(longContextResults)
	}

	slog.Info("Scaling benchmark completed",
		"component", "benchmark",
		"short_context_configs", len(shortContextResults),
		"long_context_configs", len(longContextResults),
		"short_prompt_tokens_per_sec", math.Round((1000.0/shortContextModelFit.PromptRate)*100)/100,
		"short_completion_tokens_per_sec", math.Round((1000.0/shortContextModelFit.CompletionRate)*100)/100,
		"short_r_squared", math.Round(shortContextModelFit.RSquared*100)/100,
		"long_prompt_tokens_per_sec", math.Round((1000.0/longContextModelFit.PromptRate)*100)/100,
		"long_completion_tokens_per_sec", math.Round((1000.0/longContextModelFit.CompletionRate)*100)/100,
		"long_r_squared", math.Round(longContextModelFit.RSquared*100)/100)

	return allResults, shortContextModelFit, longContextModelFit, nil
}

// MatrixResult contains benchmark results along with the driver parameters used
type MatrixResult struct {
	Params               map[string]string
	OutputFlags          map[string]bool
	Results              []*CompletionResult
	ShortContextModelFit *ModelFitResult
	LongContextModelFit  *ModelFitResult
	LocalScore           *float64
	Error                error
}

// Run is a package-level function that runs a scaling benchmark with a provided driver
func Run(d driver.Driver, driverParams map[string]interface{}) ([]*CompletionResult, *ModelFitResult, *ModelFitResult, error) {
	// Setup driver if provided
	if d != nil {
		if err := d.Setup(driverParams); err != nil {
			return nil, nil, nil, fmt.Errorf("driver setup failed: %v", err)
		}
		defer d.Teardown()
	}

	// Get URL and model from driver
	url := ""
	model := ""

	if d != nil {
		url = d.GetURL()
		model = d.GetModel().Name
	}

	// Create benchmark with URL and model from driver
	benchmark := NewBenchmark(url, model, "")
	benchmark.Driver = d

	postfix := "\nI need some filler content. Please generate as much lorem ipsum as you can."
	return benchmark.RunScalingBenchmark(postfix)
}

// RunMatrix runs benchmarks with all combinations of parameters from the matrix
func RunMatrix(driverType string, baseParams map[string]interface{}, matrix map[string]types.ParameterConfig) ([]MatrixResult, error) {
	// Create driver first
	var d driver.Driver
	var err error

	if driverType != "" {
		d, err = driver.NewDriver(driverType)
		if err != nil {
			return nil, fmt.Errorf("failed to create driver: %v", err)
		}

		// No need to set logger for the driver anymore
	}

	// Generate all combinations of parameters
	paramCombinations := generateParamCombinations(matrix)

	// If no combinations were generated, return an error
	if len(paramCombinations) == 0 {
		return nil, fmt.Errorf("no parameter combinations generated from matrix")
	}

	// Extract output flags
	outputFlags := make(map[string]bool)
	for k, config := range matrix {
		outputFlags[k] = config.Output
	}

	// Run benchmark for each combination
	var matrixResults []MatrixResult

	for _, paramSet := range paramCombinations {
		// Create a copy of base params
		params := make(map[string]interface{})
		for k, v := range baseParams {
			params[k] = v
		}

		// Override with matrix params
		for k, v := range paramSet {
			params[k] = v
		}

		// Run benchmark with this parameter set
		results, shortContextModelFit, longContextModelFit, err := Run(d, params)

		// Calculate LocalScore
		var localScore *float64
		if shortContextModelFit != nil || longContextModelFit != nil {
			modelFits := []*ModelFitResult{shortContextModelFit, longContextModelFit}
			localScore = Calculate(modelFits)
		}

		// Store results with parameter set
		matrixResult := MatrixResult{
			Params:               paramSet,
			OutputFlags:          outputFlags,
			Results:              results,
			ShortContextModelFit: shortContextModelFit,
			LongContextModelFit:  longContextModelFit,
			LocalScore:           localScore,
			Error:                err,
		}

		matrixResults = append(matrixResults, matrixResult)
	}

	return matrixResults, nil
}

// generateParamCombinations generates all possible combinations of parameters from the matrix
func generateParamCombinations(matrix map[string]types.ParameterConfig) []map[string]string {
	if len(matrix) == 0 {
		return nil
	}

	// Extract keys, values, and output flags
	var keys []string
	var valuesList [][]string
	var outputFlags []bool

	for k, config := range matrix {
		if len(config.Values) > 0 {
			keys = append(keys, k)
			valuesList = append(valuesList, config.Values)
			outputFlags = append(outputFlags, config.Output)
		}
	}

	if len(keys) == 0 {
		return nil
	}

	// Generate combinations recursively
	return generateCombinations(keys, valuesList, outputFlags, 0, make(map[string]string), nil)
}

// generateCombinations recursively generates all combinations of parameters
func generateCombinations(
	keys []string,
	valuesList [][]string,
	outputFlags []bool,
	index int,
	current map[string]string,
	result []map[string]string,
) []map[string]string {
	if index == len(keys) {
		// Make a copy of the current combination
		combination := make(map[string]string)
		for k, v := range current {
			combination[k] = v
		}
		return append(result, combination)
	}

	// For each value of the current parameter
	for _, value := range valuesList[index] {
		// Add to current combination
		current[keys[index]] = value

		// Recursively generate combinations for the next parameter
		result = generateCombinations(keys, valuesList, outputFlags, index+1, current, result)
	}

	return result
}
