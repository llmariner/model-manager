package id

import "strings"

// ToLLMarinerModelID converts a model ID to the LLMariner format.
func ToLLMarinerModelID(id string) string {
	// HuggingFace uses '/' as a separator, but Ollama does not accept. Use '-' instead for now.
	// TODO(kenji): Revisit this.
	return strings.ReplaceAll(id, "/", "-")
}
