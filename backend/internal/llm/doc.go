// Package llm provides a provider-agnostic client for OpenAI-compatible
// chat-completions endpoints, used for email style analysis and
// sensitivity classification.
//
// The client targets the OpenAI /v1/chat/completions schema, which is the
// de-facto standard exposed by local runtimes (Ollama, vLLM, llama-server)
// and by Mistral La Plateforme. Switching providers requires only changing
// configuration (base URL, API key, model) — no code changes:
//
//	// Local Ollama (default)
//	c := llm.New(llm.Config{})
//
//	// Mistral La Plateforme (EU-hosted)
//	c := llm.New(llm.Config{
//	    BaseURL: "https://api.mistral.ai/v1",
//	    APIKey:  os.Getenv(llm.EnvAPIKey),
//	    Model:   "mistral-small-latest",
//	})
//
// Handlers should depend on the Client interface so a fake can be
// substituted in tests.
package llm
