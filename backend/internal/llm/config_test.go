package llm

import (
	"testing"
)

func TestLoadConfigUsesJSONKeyOrder(t *testing.T) {
	data := []byte(`{
  "z_last": {"model_name": "z", "token": "t", "base_url": "https://example.com"},
  "_embedding": {"api_key": "k", "base_url": "https://example.com", "model": "emb"},
  "a_first": {"model_name": "a", "token": "t", "base_url": "https://example.com"}
}`)

	cfg, err := parseConfigData(data, "test")
	if err != nil {
		t.Fatalf("parseConfigData: %v", err)
	}
	if cfg.DefaultModel != "z_last" {
		t.Fatalf("DefaultModel = %q, want z_last (first key in JSON)", cfg.DefaultModel)
	}
}

func TestResolveDefaultModel(t *testing.T) {
	cfg := &Config{
		DefaultModel: "alpha",
		Models: map[string]ModelConfig{
			"alpha": {ModelName: "alpha-model"},
			"beta":  {ModelName: "beta-model"},
		},
	}

	key, fallback := ResolveDefaultModel(DefaultModelFirstInJSON, cfg)
	if key != "alpha" || fallback {
		t.Fatalf("first_in_json => (%q, %v), want (alpha, false)", key, fallback)
	}

	key, fallback = ResolveDefaultModel("beta", cfg)
	if key != "beta" || fallback {
		t.Fatalf("beta => (%q, %v), want (beta, false)", key, fallback)
	}

	key, fallback = ResolveDefaultModel("missing", cfg)
	if key != "alpha" || !fallback {
		t.Fatalf("missing => (%q, %v), want (alpha, true)", key, fallback)
	}
}
