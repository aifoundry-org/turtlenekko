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
	PromptTokens       int
	CachedPromptTokens int
	CompletionTokens   int
	ResponseTime       time.Duration
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
func (b *Benchmark) RunWithPromptLength(promptLength int, maxCompletionTokens int, postfix string) ([]*CompletionResult, error) {
	slog.Info("Running benchmark",
		"component", "benchmark",
		"prompt_length", promptLength,
		"max_tokens", maxCompletionTokens,
		"url", b.URL)

	results := []*CompletionResult{}

	messages := generateMessages(promptLength, postfix)

	// Parameters with specified prompt length and max tokens
	params := ChatCompletionParams{
		Messages:            messages,
		Temperature:         0.0, // Use deterministic sampling
		TopP:                1.0,
		MaxCompletionTokens: maxCompletionTokens,
		Seed:                42, // Fixed seed for reproducibility
	}

	// Make the actual request to the LLM
	completionResult, err := b.ChatCompletion(params)

	// Small delay between requests to avoid overwhelming the server
	time.Sleep(500 * time.Millisecond)

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

	results = append(results, completionResult)

	// Repeat with the same messages
	// Make the actual request to the LLM
	cachedCompletionResult, err := b.ChatCompletion(params)

	// Small delay between requests to avoid overwhelming the server
	time.Sleep(500 * time.Millisecond)

	if err != nil {
		slog.Error("Benchmark failed",
			"component", "benchmark",
			"prompt_length", promptLength,
			"max_tokens", maxCompletionTokens,
			"error", err)
		return nil, fmt.Errorf("chat completion failed: %v", err)
	}

	// All prompt was cached cached
	cachedCompletionResult.CachedPromptTokens = cachedCompletionResult.PromptTokens
	cachedCompletionResult.PromptTokens = 0

	// Log the detailed timing information
	slog.Info("Benchmark completed (cached)",
		"component", "benchmark",
		"prompt_length", promptLength,
		"max_tokens", maxCompletionTokens,
		"response_time_ms", completionResult.ResponseTime.Milliseconds())

	results = append(results, cachedCompletionResult)

	return results, nil
}

// ModelFitResult contains the fitted parameters for the completion time model
type ModelFitResult struct {
	PromptRate       float64 // ms per prompt token
	CachedPromptRate float64 // ms per cached prompt token
	CompletionRate   float64 // ms per completion token
	RSquared         float64 // goodness of fit (0-1)
}

// fitCompletionTimeModel fits the model: completion_time = a * prompt_tokens + b * cached_prompt_tokens + c * completion_tokens
// to the measured data using linear regression (ordinary least squares)
func fitCompletionTimeModel(results []*CompletionResult) *ModelFitResult {
	if len(results) < 2 {
		slog.Warn("Not enough results for model fitting", "component", "benchmark", "count", len(results))
		return &ModelFitResult{
			PromptRate:       0,
			CachedPromptRate: 0,
			CompletionRate:   0,
			RSquared:         0,
		}
	}

	// Count valid results and log input data
	validResults := 0
	slog.Info("Model fitting input data:", "component", "benchmark")

	// Prepare data for linear regression
	var X [][]float64 // Features: [prompt_tokens, cached_prompt_tokens, completion_tokens]
	var y []float64   // Target: response_time_ms

	for i, r := range results {
		if r == nil {
			continue
		}
		validResults++

		// Log data point
		slog.Info("Data point",
			"component", "benchmark",
			"index", i,
			"prompt_tokens", r.PromptTokens,
			"cached_prompt_tokens", r.CachedPromptTokens,
			"completion_tokens", r.CompletionTokens,
			"response_time_ms", r.ResponseTime.Milliseconds())

		// Add to regression data
		X = append(X, []float64{
			float64(r.PromptTokens),
			float64(r.CachedPromptTokens),
			float64(r.CompletionTokens),
		})
		y = append(y, float64(r.ResponseTime.Milliseconds()))
	}

	slog.Info("Starting linear regression", "component", "benchmark", "valid_results", validResults)

	// Calculate means
	meanX := make([]float64, 3)
	meanY := 0.0

	for i := 0; i < len(X); i++ {
		for j := 0; j < 3; j++ {
			meanX[j] += X[i][j]
		}
		meanY += y[i]
	}

	for j := 0; j < 3; j++ {
		meanX[j] /= float64(len(X))
	}
	meanY /= float64(len(y))

	// Calculate coefficients using normal equations
	// (X^T * X)^(-1) * X^T * y

	// First, calculate X^T * X (3x3 matrix)
	xtx := make([][]float64, 3)
	for i := range xtx {
		xtx[i] = make([]float64, 3)
	}

	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			for k := 0; k < len(X); k++ {
				xtx[i][j] += X[k][i] * X[k][j]
			}
		}
	}

	// Calculate X^T * y (3x1 vector)
	xty := make([]float64, 3)
	for i := 0; i < 3; i++ {
		for k := 0; k < len(X); k++ {
			xty[i] += X[k][i] * y[k]
		}
	}

	// Solve the system of equations using Gaussian elimination
	// We're solving: xtx * [a, b, c] = xty

	// Check if the matrix is invertible (non-zero determinant)
	// For simplicity, we'll just check if any column is all zeros
	for j := 0; j < 3; j++ {
		allZeros := true
		for i := 0; i < 3; i++ {
			if math.Abs(xtx[i][j]) > 1e-10 {
				allZeros = false
				break
			}
		}
		if allZeros {
			slog.Warn("Matrix is singular, using fallback values", "component", "benchmark", "column", j)
			// Use reasonable fallback values based on the data
			a := 3.0  // ~3ms per prompt token
			b := 0.01 // ~0.01ms per cached token
			c := 25.0 // ~25ms per completion token

			slog.Info("Using fallback model parameters",
				"component", "benchmark",
				"prompt_rate_ms_per_token", a,
				"cached_prompt_rate_ms_per_token", b,
				"completion_rate_ms_per_token", c)

			return &ModelFitResult{
				PromptRate:       a,
				CachedPromptRate: b,
				CompletionRate:   c,
				RSquared:         0.5, // Reasonable default
			}
		}
	}

	// Augment the matrix for Gaussian elimination
	augmented := make([][]float64, 3)
	for i := range augmented {
		augmented[i] = make([]float64, 4)
		for j := 0; j < 3; j++ {
			augmented[i][j] = xtx[i][j]
		}
		augmented[i][3] = xty[i]
	}

	// Gaussian elimination
	for i := 0; i < 3; i++ {
		// Find pivot
		maxRow := i
		for j := i + 1; j < 3; j++ {
			if math.Abs(augmented[j][i]) > math.Abs(augmented[maxRow][i]) {
				maxRow = j
			}
		}

		// Swap rows
		augmented[i], augmented[maxRow] = augmented[maxRow], augmented[i]

		// Check for numerical stability
		if math.Abs(augmented[i][i]) < 1e-10 {
			slog.Warn("Matrix is nearly singular, using fallback values", "component", "benchmark")
			// Use reasonable fallback values
			a := 3.0  // ~3ms per prompt token
			b := 0.01 // ~0.01ms per cached token
			c := 25.0 // ~25ms per completion token

			return &ModelFitResult{
				PromptRate:       a,
				CachedPromptRate: b,
				CompletionRate:   c,
				RSquared:         0.5, // Reasonable default
			}
		}

		// Scale row
		pivot := augmented[i][i]
		for j := i; j < 4; j++ {
			augmented[i][j] /= pivot
		}

		// Eliminate other rows
		for j := 0; j < 3; j++ {
			if j != i {
				factor := augmented[j][i]
				for k := i; k < 4; k++ {
					augmented[j][k] -= factor * augmented[i][k]
				}
			}
		}
	}

	// Extract coefficients
	a := augmented[0][3]
	b := augmented[1][3]
	c := augmented[2][3]

	// Ensure coefficients are non-negative
	a = math.Max(0.01, a)  // Minimum 0.01ms per token
	b = math.Max(0.001, b) // Minimum 0.001ms per token
	c = math.Max(0.1, c)   // Minimum 0.1ms per token

	slog.Info("Linear regression results",
		"component", "benchmark",
		"prompt_rate_ms_per_token", a,
		"cached_prompt_rate_ms_per_token", b,
		"completion_rate_ms_per_token", c)

	// Calculate R-squared
	totalSumSquares := 0.0
	residualSumSquares := 0.0

	// Log predictions vs actual values
	slog.Info("Model predictions:", "component", "benchmark")

	for i, r := range results {
		if r == nil {
			continue
		}

		y := float64(r.ResponseTime.Milliseconds())
		yPred := a*float64(r.PromptTokens) + b*float64(r.CachedPromptTokens) + c*float64(r.CompletionTokens)

		totalSumSquares += math.Pow(y-meanY, 2)
		residualSumSquares += math.Pow(y-yPred, 2)

		slog.Info("Prediction",
			"component", "benchmark",
			"index", i,
			"actual_ms", y,
			"predicted_ms", yPred,
			"error_ms", y-yPred,
			"prompt_tokens", r.PromptTokens,
			"cached_prompt_tokens", r.CachedPromptTokens,
			"completion_tokens", r.CompletionTokens)
	}

	rSquared := 0.0
	if totalSumSquares > 0 {
		rSquared = 1.0 - (residualSumSquares / totalSumSquares)
	}

	slog.Info("R-squared calculation",
		"component", "benchmark",
		"total_sum_squares", totalSumSquares,
		"residual_sum_squares", residualSumSquares,
		"r_squared", rSquared)

	// Convert rates from ms/token to tokens/sec for easier interpretation
	promptRate := 1000.0 / a
	cachedPromptRate := 1000.0 / b
	completionRate := 1000.0 / c

	slog.Info("Final model metrics",
		"component", "benchmark",
		"prompt_tokens_per_sec", promptRate,
		"cached_prompt_tokens_per_sec", cachedPromptRate,
		"completion_tokens_per_sec", completionRate,
		"r_squared", rSquared)

	return &ModelFitResult{
		PromptRate:       a,
		CachedPromptRate: b,
		CompletionRate:   c,
		RSquared:         rSquared,
	}
}

// BenchmarkConfig represents a single benchmark configuration
type BenchmarkConfig struct {
	PromptLength int
	MaxTokens    int
}

// Constants for benchmark quality control
const (
	MinAcceptableRSquared  = 0.99 // Minimum acceptable R-squared value
	MaxBenchmarkIterations = 3    // Maximum number of iterations to try
)

// runContextBenchmark runs benchmarks for a specific context size (short or long)
func (b *Benchmark) runContextBenchmark(contextType string, configs []BenchmarkConfig, postfix string) ([]*CompletionResult, *ModelFitResult, error) {
	slog.Info(fmt.Sprintf("Running %s context benchmarks", contextType), "component", "benchmark")

	// Map to store the fastest result for each token count combination
	// Key format: "promptTokens:cachedPromptTokens:completionTokens"
	bestResults := make(map[string]*CompletionResult)

	// Track which configs have been run
	configsRun := make(map[string]bool)

	// Run up to MaxBenchmarkIterations
	for iteration := 1; iteration <= MaxBenchmarkIterations; iteration++ {
		slog.Info(fmt.Sprintf("Starting %s context benchmark iteration %d/%d",
			contextType, iteration, MaxBenchmarkIterations), "component", "benchmark")

		// Run benchmarks for configurations that haven't been run yet
		for _, config := range configs {
			// Create a key for this config
			configKey := fmt.Sprintf("%d:%d", config.PromptLength, config.MaxTokens)

			// Skip if we've already run this config in a previous iteration
			if iteration > 1 && configsRun[configKey] {
				continue
			}

			// Mark this config as run
			configsRun[configKey] = true

			results, err := b.RunWithPromptLength(config.PromptLength, config.MaxTokens, postfix)

			if err != nil {
				slog.Error(fmt.Sprintf("%s context benchmark failed", contextType),
					"component", "benchmark",
					"prompt_length", config.PromptLength,
					"max_tokens", config.MaxTokens,
					"error", err)
			} else {
				// Process each result and keep only the fastest for each token combination
				for _, result := range results {
					if result == nil {
						continue
					}

					// Create a key based on token counts
					key := fmt.Sprintf("%d:%d:%d",
						result.PromptTokens,
						result.CachedPromptTokens,
						result.CompletionTokens)

					// Check if we already have a result for this token combination
					existing, exists := bestResults[key]
					if !exists || result.ResponseTime < existing.ResponseTime {
						// This is either the first result for this combination or faster than the previous one
						bestResults[key] = result
						slog.Info("New best result for token combination",
							"component", "benchmark",
							"iteration", iteration,
							"context_type", contextType,
							"prompt_tokens", result.PromptTokens,
							"cached_prompt_tokens", result.CachedPromptTokens,
							"completion_tokens", result.CompletionTokens,
							"response_time_ms", result.ResponseTime.Milliseconds())
					}
				}
			}

			// After each config, check if we have enough data for a good fit
			if len(bestResults) >= 8 { // Need at least 8 data points for a meaningful fit
				// Convert map to slice for model fitting
				var currentResults []*CompletionResult
				for _, result := range bestResults {
					currentResults = append(currentResults, result)
				}

				// Try to fit the model with current results
				currentFit := fitCompletionTimeModel(currentResults)

				slog.Info(fmt.Sprintf("Intermediate %s model fit after %d configs", contextType, len(configsRun)),
					"component", "benchmark",
					"iteration", iteration,
					"r_squared", currentFit.RSquared,
					"configs_run", len(configsRun))

				// If R-squared is good enough, we can stop
				if currentFit.RSquared >= MinAcceptableRSquared {
					slog.Info(fmt.Sprintf("Achieved acceptable R-squared for %s context", contextType),
						"component", "benchmark",
						"iteration", iteration,
						"r_squared", currentFit.RSquared)

					return currentResults, currentFit, nil
				}
			}
		}

		// At the end of each iteration, check if we need to continue
		// Convert map to slice for model fitting
		var contextResults []*CompletionResult
		for _, result := range bestResults {
			contextResults = append(contextResults, result)
		}

		slog.Info(fmt.Sprintf("Completed iteration %d for %s context with %d results",
			iteration, contextType, len(contextResults)), "component", "benchmark")

		// If this is the last iteration or we don't have enough results, return what we have
		if iteration == MaxBenchmarkIterations || len(contextResults) < 4 {
			var modelFit *ModelFitResult
			if len(contextResults) >= 4 {
				modelFit = fitCompletionTimeModel(contextResults)
				slog.Info(fmt.Sprintf("Final %s model fit after %d iterations", contextType, iteration),
					"component", "benchmark",
					"r_squared", modelFit.RSquared)
			} else {
				slog.Warn(fmt.Sprintf("Not enough data points for %s model fit", contextType),
					"component", "benchmark",
					"data_points", len(contextResults))
			}
			return contextResults, modelFit, nil
		}

		// Otherwise, log that we're going to try another iteration
		slog.Info(fmt.Sprintf("R-squared not acceptable for %s context, running another iteration", contextType),
			"component", "benchmark",
			"iteration", iteration,
			"min_acceptable", MinAcceptableRSquared)
	}

	// This should never be reached, but just in case
	var contextResults []*CompletionResult
	for _, result := range bestResults {
		contextResults = append(contextResults, result)
	}

	var modelFit *ModelFitResult
	if len(contextResults) >= 4 {
		modelFit = fitCompletionTimeModel(contextResults)
	}

	return contextResults, modelFit, nil
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

	// Define benchmark configurations
	shortContextConfigs := []BenchmarkConfig{
		{PromptLength: 100, MaxTokens: 1},
		{PromptLength: 100, MaxTokens: 100},
		{PromptLength: 500, MaxTokens: 1},
		{PromptLength: 500, MaxTokens: 100},
	}

	longContextConfigs := []BenchmarkConfig{
		{PromptLength: 9000, MaxTokens: 1},
		{PromptLength: 9000, MaxTokens: 100},
		{PromptLength: 10000, MaxTokens: 1},
		{PromptLength: 10000, MaxTokens: 100},
	}

	// Run benchmarks for each context size
	shortContextResults, shortContextModelFit, _ := b.runContextBenchmark("short", shortContextConfigs, postfix)
	longContextResults, longContextModelFit, _ := b.runContextBenchmark("long", longContextConfigs, postfix)

	// Combine all results
	allResults := append(shortContextResults, longContextResults...)

	// Check if all benchmarks failed
	if len(shortContextResults) == 0 && len(longContextResults) == 0 {
		return allResults, nil, nil, fmt.Errorf("all benchmark configurations failed")
	}

	// Log summary of results
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
