package benchmark

import (
	"math"
)

// Constants for LocalScore calculation
const (
	// Average prompt tokens across all test scenarios
	// Derived from https://github.com/Mozilla-Ocho/llamafile/blob/e6daab04b51482009bf598a7cdaddeed8a1ba197/localscore/localscore.cpp#L387
	// (1024 + 4096 + 2048 + 2048 + 1024 + 1280 + 384 + 64 + 16) / 9 = 1331.56
	AvgPromptTokens = 1331.56

	// Scaling factor for the final score
	ScalingFactor = 10.0
)

// Calculate computes the effective estimated LocalScore based on benchmark results
// https://github.com/Mozilla-Ocho/llamafile/blob/e6daab04b51482009bf598a7cdaddeed8a1ba197/localscore/localscore.cpp#L331
// Estimated LocalScore = (prompt_tps * gen_tps * (1000/ttft_ms))^(1/3) * 10
func Calculate(modelFits []*ModelFitResult) *float64 {
	var result *float64 = nil

	// Calculate average prompt and generation TPS across all contexts
	var promptTPS, genTPS float64
	var validContexts int

	// Add metrics from all valid model fits
	for _, modelFit := range modelFits {
		if modelFit != nil && modelFit.PromptRate > 0 && modelFit.CompletionRate > 0 {
			promptTPS += 1000.0 / modelFit.PromptRate
			genTPS += 1000.0 / modelFit.CompletionRate
			validContexts++
		}
	}

	// Calculate average if we have valid data
	if validContexts > 0 {
		promptTPS /= float64(validContexts)
		genTPS /= float64(validContexts)

		// Calculate effective TTFT in milliseconds
		// ttft_ms = avg_prompt_tokens / prompt_tps * 1000
		ttftMS := AvgPromptTokens / promptTPS * 1000.0

		// Calculate the geometric mean and apply scaling factor
		// score = (prompt_tps * gen_tps * (1000/ttft_ms))^(1/3) * 10
		score := math.Pow(promptTPS*genTPS*(1000.0/ttftMS), 1.0/3.0) * ScalingFactor

		// Round to 2 decimal places
		scoreValue := math.Round(score*100) / 100
		result = &scoreValue
	}

	return result
}
