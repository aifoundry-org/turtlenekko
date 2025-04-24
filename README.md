# Turtlenekko

A benchmarking tool for measuring and comparing performance of LLM inference runtimes.

## Overview

Turtlenekko is a benchmarking tool designed to measure the performance of Large
Language Models (LLMs) through OpenAI API compatible chat completion endpoints.
It provides reports on token processing speeds.

Many LLM benchmarking tools assume specific inference engines and only measure hardware 
performance (often for a short list of preselected models). While this is good
for apples-to-apples comparisons of the most popular software stacks on
assorted hardware, it falls short when it comes to less popular software/hardware
combinations.

The main goal of Turtlenekko is benchmarking arbitrary inference runtimes
as long as they expose an OpenAI API compatible(-ish) chat completions API. This
enables users to evaluate less popular inference runtimes, especially those
that employ non-mainstream hardware, such as custom accelerators.

Another Turtlenekko goal is exploring the space of different configurations.
Nothing is fixed, and you can run different models against the same runtime. Why
not measure how adding more threads improves performance? Does enabling some
optimization (like flash attention) actually help? Turtlenekko can do all of
the above (and anything else you can parameterize about the runtime) in a single
run.

Data like this can help select the optimal configuration for your deployment, guide
inference engine optimization efforts, or even act as a performance regression
checker for your inference provider.


Key capabilities:
- Inference runtime agnostic - works with any LLM server that exposes an OpenAI API
  compatible chat completion endpoint
- Supports both local and remote LLM deployments
- Conducts parametrized benchmarks across multiple dimensions (models, CPU
  allocations, batch sizes, etc.)
- Tests all combinations of parameters through a flexible parameter matrix
- Generates detailed performance reports in multiple formats (JSON, CSV, text)
- Originally developed to benchmark
  [NekkoAPI](https://github.com/aifoundry-org/NekkoAPI), but works with any
  compatible LLM server
- Repeats benchmark samples several times until result is reliable enough
- Measures KV cache reuse (making a call with the same prompt is expected
  to result in negligible prompt processing times)

Limitations:
- Doesn't support text completion API
- Doesn't measure performance of concurrent requests (thus ignoring benefits of
  continuous batching)
- Doesn't work correctly on model architectures that support dynamic
  attention where inference speed depends on the content of the prompt.
  This limitation is caused by pure gibberish being used as prompts.
- Slow. Like really slow. Turtlenekko itself is not slow, but
  multiple samples per run with multiple runs per matrix can
  take quite some time.


TODO:
- Measure performance of concurrent requests
- Use client-side token counter when tokenizer is known (should increase
  precision and lower benchmarking duration)
- Add support for text completions endpoint
- Add `exclude` and other filtering constructs to the parameter matrix
- Implement web API to run benchmarks with precise control

## Installation

```bash
go install github.com/aifoundry-org/turtlenekko@latest
```

## Usage

Turtlenekko uses a configuration file to define benchmark parameters. You can create a default configuration file with:

```bash
turtlenekko init
```

Run a benchmark with:

```bash
turtlenekko benchmark --config config.yaml --format json
```

Available output formats:
- `json`: Structured JSON output for programmatic consumption and integration with other tools
- `text`: Human-readable text output for quick analysis
- `csv`: CSV format for spreadsheet analysis and data visualization

#### Output Format Details

##### JSON Format

The JSON output provides detailed benchmark results in a structured format:

```json
[
  {
    "params": {
      "model": "llama3-7b",
      "threads": "8"
    },
    "short_context_prompt_tokens_per_sec": 2380.95,
    "short_context_cached_prompt_tokens_per_sec": 12500.00,
    "short_context_completion_tokens_per_sec": 7.96,
    "short_context_r_squared": 0.99,
    "long_context_prompt_tokens_per_sec": 1123.60,
    "long_context_cached_prompt_tokens_per_sec": 8333.33,
    "long_context_completion_tokens_per_sec": 5.34,
    "long_context_r_squared": 0.99,
    "localscore_estimate": 20.95
  },
  {
    "params": {
      "model": "mistral-7b",
      "threads": "4"
    },
    "short_context_prompt_tokens_per_sec": 1960.78,
    "short_context_cached_prompt_tokens_per_sec": 10000.00,
    "short_context_completion_tokens_per_sec": 10.17,
    "short_context_r_squared": 0.99,
    "long_context_prompt_tokens_per_sec": 952.38,
    "long_context_cached_prompt_tokens_per_sec": 7142.86,
    "long_context_completion_tokens_per_sec": 6.89,
    "long_context_r_squared": 0.99,
    "localscore_estimate": 21.88
  }
]
```

Each object in the array represents one benchmark run with:
- `params`: The parameters used for this run (only those with `output: true`)
- Short context metrics (few hundred tokens):
  - `short_context_prompt_tokens_per_sec`: Prompt tokens processed per second
  - `short_context_cached_prompt_tokens_per_sec`: Cached prompt tokens processed per second (KV cache reuse)
  - `short_context_completion_tokens_per_sec`: Completion tokens generated per second
  - `short_context_r_squared`: Statistical measure of how well the model fits the data (0-1)
- Long context metrics (around 3000 tokens):
  - `long_context_prompt_tokens_per_sec`: Prompt tokens processed per second
  - `long_context_cached_prompt_tokens_per_sec`: Cached prompt tokens processed per second (KV cache reuse)
  - `long_context_completion_tokens_per_sec`: Completion tokens generated per second
  - `long_context_r_squared`: Statistical measure of how well the model fits the data (0-1)
- `localscore_estimate`: Estimated LocalScore - a composite performance score
  based on average prompt speed, generation speed, and responsiveness across both
  contexts

##### CSV Format

The CSV output is ideal for importing into spreadsheet applications:

```
model,threads,short_context_prompt_tokens_per_sec,short_context_cached_prompt_tokens_per_sec,short_context_completion_tokens_per_sec,short_context_r_squared,long_context_prompt_tokens_per_sec,long_context_cached_prompt_tokens_per_sec,long_context_completion_tokens_per_sec,long_context_r_squared,localscore_estimate
llama3-7b,8,2380.95,12500.00,7.96,0.99,1123.60,8333.33,5.34,0.99,20.95
mistral-7b,4,1960.78,10000.00,10.17,0.99,952.38,7142.86,6.89,0.99,21.88
```

The CSV includes:
- All parameters marked with `output: true` in the configuration
- All performance metrics in a tabular format
- Headers for easy identification of columns

##### Text Format
The text output provides a human-readable summary of each benchmark run:

```
=== Matrix Combination 1 ===
Parameters:
  model: llama3-7b
  threads: 8

Short Context Results:
  Prompt processing: 2380.95 tokens/sec
  Cached prompt processing: 12500.00 tokens/sec
  Completion generation: 7.96 tokens/sec
  Model fit quality (R²): 0.99

Long Context Results:
  Prompt processing: 1123.60 tokens/sec
  Cached prompt processing: 8333.33 tokens/sec
  Completion generation: 5.34 tokens/sec
  Model fit quality (R²): 0.99

Localscore Estimate: 20.95
```

### Drivers

Turtlenekko supports different drivers to manage the LLM runtime environment:

#### 1. Dummy Driver

The dummy driver doesn't set up any environment and simply connects to an
already running LLM server.

**Configuration Example:**

```yaml
driver: "dummy"
matrix:
  url:
    values: ["http://localhost:8000/v1/chat/completions"]
    output: true
  model:
    values: ["llama3"]
    output: true
```

**Parameters:**
- `url`: The endpoint URL of the LLM server (required)
- `model`: The model name to use (required)

#### 2. Local Command Driver

The local_cmd driver executes shell commands to start and stop the LLM server
before and after benchmarking. This is useful for testing different server
configurations or when you need to manage the server lifecycle.

**Configuration Example:**

```yaml
driver: "local_cmd"
matrix:
  url:
    values: ["http://localhost:8000/v1/chat/completions"]
    output: false
  model:
    values: ["/models/llama3-7b.gguf", "/models/mistral-7b.gguf"]
    output: true
  threads:
    values: ["4", "8"]
    output: true
  setup_cmd:
    values: ["docker run -d --rm -p 8000:8000 -v ~/models:/models -e THREADS={{.threads}} -e MODEL_PATH={{.model}} llm-server:latest"]
    output: false
  teardown_cmd:
    values: ["docker stop $(docker ps -q --filter ancestor=llm-server:latest)"]
    output: false
```

**Parameters:**
- `url`: The endpoint URL of the LLM server (required)
- `model`: The model name or path to use (required)
- `setup_cmd`: Command to run before benchmarking (supports Go templates for parameter interpolation)
- `teardown_cmd`: Command to run after benchmarking (supports Go templates)
- Any additional parameters you want to test in your matrix

**Template Variables:**
The setup_cmd and teardown_cmd support Go template variables that are replaced
with the current parameter values. For example, `{{.threads}}` will be replaced
with the current value of the "threads" parameter.

### Parameter Matrix

The `matrix` section defines parameters to test in all possible combinations:

```yaml
matrix:
  parameter1:
    values: ["value1", "value2"]
    output: true  # Include in results output
  parameter2:
    values: ["valueA", "valueB"]
    output: false  # Don't include in results output
```

Each parameter can be specified as:
1. A simple array: `param: ["value1", "value2"]`
2. An object with values and output flag: `param: {values: ["value1", "value2"], output: true}`

The `output` flag controls whether the parameter appears in the benchmark results.

## Methodology

Turtlenekko uses a statistical approach to measure LLM performance metrics that
cannot be directly controlled through the OpenAI API interface:

### Measurement Challenges

When benchmarking LLMs through a chat completion API, several challenges exist:
- We cannot precisely control the exact number of tokens processed
- The API returns total response time, but doesn't break down processing stages
- We have to actively prevent the runtime from using KV cache for subsequent
  requests to get representative results

### Regression-Based Approach

To overcome these limitations, Turtlenekko:

1. **Samples Multiple Data Points**: Runs benchmarks with varying prompt lengths and completion token limits
2. **Measures at Different Context Lengths**: Takes measurements at both short context (few hundred tokens) and long context (around 10,000 tokens) to capture performance degradation as context grows
3. **Randomizes Prompts**: Generates random prompts to prevent KV cache reuse
4. **Collects Measurements**: For each run, records:
   - Prompt token count (as reported by the API)
   - Completion token count (as reported by the API)
   - Total response time
5. **Fits Linear Regression Models**: Uses the equation:
   ```
   response_time = prompt_rate * prompt_tokens + cached_prompt_rate * cached_prompt_tokens + completion_rate * completion_tokens
   ```
   Separate models are fitted for short and long contexts.
6. **Calculates Key Metrics**:
   - **Prompt Processing Rate**: Time per prompt token (milliseconds) for both short and long contexts
   - **Cached Prompt Processing Rate**: Time per cached prompt token (milliseconds) when KV cache is reused
   - **Completion Generation Rate**: Time per completion token (milliseconds) for both short and long contexts
   - **R-squared value**: Indicates how well each model fits the data (0-1)

This approach allows Turtlenekko to:
- Separate the time spent on processing the input prompt from the time spent generating the completion
- Measure how performance degrades as context length increases

### Estimated LocalScore

Turtlenekko calculates a composite performance metric called estimated
LocalScore, inspired by the scoring system used in the [LocalScore
benchmark](https://www.localscore.ai/). The formula is:

```
LocalScore = (prompt_tps * gen_tps * (1000/ttft_ms))^(1/3) * 10
```

Where:
- `prompt_tps`: Average prompt tokens processed per second across both short and long contexts
- `gen_tps`: Average completion tokens generated per second across both short and long contexts
- `ttft_ms`: Time to first token in milliseconds (calculated as average prompt tokens / prompt_tps * 1000)

The LocalScore is the geometric mean of these three metrics (with TTFT inverted
since lower is better), multiplied by 10 for readability. This provides a single
number that balances:

1. Prompt processing speed
2. Generation speed
3. Responsiveness (via TTFT)

A higher estimated LocalScore indicates better overall performance. By averaging
across both short and long contexts, the score reflects the model's performance
across the entire context window range.

Turtlenekko can calculate LocalScore estimates for you (`--localscore` command line argument, default: true).
If you wan't numbers that are somewhat comparable to the official LocalScore scores,
there are premade congurations in the `examples/localscore` folder to benchmark NekkoAPI runtime
against the models used by LocalScore tool. Just run:

```sh
# To run localscore benchmarks you have to clone Turtlenekko repository first:
git clone https://github.com/aifoundry-org/turtlenekko.git
cd turtlenekko

# Run the benchmarks:
make localscore-tiny
# or
make localscore-small
# or
make localscore-medium
```

These will download corresponding models from Hugging Face and run the benchmarks.

[Note]: models are downloaded without authentication, rate limits may apply.


## License

[Apache License 2.0](LICENSE)

