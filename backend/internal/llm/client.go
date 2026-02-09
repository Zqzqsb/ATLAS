// Package llm provides LLM configuration loading and model creation.
//
// This package is a thin wrapper around langchaingo, handling:
//   - Loading model configs from llm_config.json
//   - Creating langchaingo llms.Model instances
//
// All consumers should use llms.Model directly (the langchaingo interface).
package llm
