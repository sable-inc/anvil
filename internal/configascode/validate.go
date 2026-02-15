package configascode

import (
	"fmt"
	"strings"
)

// ValidationError represents a single validation issue at a specific path.
type ValidationError struct {
	Path    string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Path, e.Message)
}

// ValidationErrors collects multiple validation issues.
type ValidationErrors struct {
	Errors []ValidationError
}

func (e *ValidationErrors) Error() string {
	if len(e.Errors) == 0 {
		return "validation passed"
	}
	msgs := make([]string, len(e.Errors))
	for i, err := range e.Errors {
		msgs[i] = err.Error()
	}
	return strings.Join(msgs, "\n")
}

// Validate performs local validation of an AgentConfig, mirroring
// the sable-api Zod superRefine rules. Returns nil if valid.
func Validate(cfg *AgentConfig) error {
	var errs []ValidationError

	// Required fields.
	if cfg.Name == "" {
		errs = append(errs, ValidationError{Path: "name", Message: "name is required"})
	}
	if cfg.GreetingInstructions == "" {
		errs = append(errs, ValidationError{Path: "greeting_instructions", Message: "greeting_instructions is required"})
	}
	if !isValidEnvironment(cfg.Environment) {
		errs = append(errs, ValidationError{Path: "environment", Message: "must be one of: development, staging, production"})
	}
	if len(cfg.LLM.Modalities) == 0 {
		errs = append(errs, ValidationError{Path: "llm.modalities", Message: "at least one modality is required"})
	}

	// TTS cross-field rules.
	errs = append(errs, validateTTS(&cfg.TTS)...)

	// STT cross-field rules.
	errs = append(errs, validateSTT(&cfg.STT)...)

	// Component cross-field rules.
	errs = append(errs, validateComponents(&cfg.Components)...)

	if len(errs) == 0 {
		return nil
	}
	return &ValidationErrors{Errors: errs}
}

func validateTTS(tts *TTSConfig) []ValidationError {
	switch tts.Provider {
	case "elevenlabs":
		return validateTTSElevenLabs(tts)
	case "none":
		return validateTTSNullFields(tts, "none")
	case "openai":
		return validateTTSNullFields(tts, "openai")
	default:
		return nil
	}
}

func validateTTSElevenLabs(tts *TTSConfig) []ValidationError {
	var errs []ValidationError
	if tts.VoiceID == nil || *tts.VoiceID == "" {
		errs = append(errs, ValidationError{
			Path: "tts.voice_id", Message: "voice_id is required when provider is elevenlabs",
		})
	}
	if tts.Model == nil || *tts.Model == "" {
		errs = append(errs, ValidationError{
			Path: "tts.model", Message: "model is required when provider is elevenlabs",
		})
	}
	return errs
}

func validateTTSNullFields(tts *TTSConfig, provider string) []ValidationError {
	var errs []ValidationError
	if tts.VoiceID != nil && *tts.VoiceID != "" {
		errs = append(errs, ValidationError{
			Path: "tts.voice_id", Message: "voice_id must be null when provider is " + provider,
		})
	}
	if tts.Model != nil && *tts.Model != "" {
		errs = append(errs, ValidationError{
			Path: "tts.model", Message: "model must be null when provider is " + provider,
		})
	}
	return errs
}

func validateSTT(stt *STTConfig) []ValidationError {
	var errs []ValidationError

	switch stt.Provider {
	case "openai":
		if !stt.OpenAI.Enabled {
			errs = append(errs, ValidationError{
				Path: "stt.openai.enabled", Message: "openai.enabled must be true when stt.provider is openai",
			})
		}
	case "deepgram":
		if !stt.Deepgram.Enabled {
			errs = append(errs, ValidationError{
				Path: "stt.deepgram.enabled", Message: "deepgram.enabled must be true when stt.provider is deepgram",
			})
		}
	case "both":
		if !stt.OpenAI.Enabled || !stt.Deepgram.Enabled {
			errs = append(errs, ValidationError{
				Path: "stt.provider", Message: "openai.enabled and deepgram.enabled must be true when stt.provider is both",
			})
		}
	}

	return errs
}

func validateComponents(c *ComponentsConfig) []ValidationError {
	var errs []ValidationError

	// Vision: proactive requires enabled.
	if c.Vision.Proactive && !c.Vision.Enabled {
		errs = append(errs, ValidationError{
			Path: "components.vision.enabled", Message: "vision.enabled must be true when vision.proactive is enabled",
		})
	}

	// Browser: streaming requires enabled.
	if !c.Browser.Enabled && c.Browser.EnableStreaming {
		errs = append(errs, ValidationError{
			Path:    "components.browser.enable_streaming",
			Message: "enable_streaming must be false when browser is disabled",
		})
	}

	// RAG: enabled requires an index source.
	if c.RAG.Enabled {
		hasIndex := c.RAG.IndexPath != nil && *c.RAG.IndexPath != ""
		hasPinecone := c.RAG.PineconeIndexName != nil && *c.RAG.PineconeIndexName != ""
		if !hasIndex && !hasPinecone {
			errs = append(errs, ValidationError{
				Path:    "components.rag.index_path",
				Message: "provide index_path or pinecone_index_name when RAG is enabled",
			})
		}
	}

	// RAG: embeddings dimension must match model.
	if c.RAG.EmbeddingsModel == "text-embedding-3-small" && c.RAG.EmbeddingsDimension != 1536 {
		errs = append(errs, ValidationError{
			Path:    "components.rag.embeddings_dimension",
			Message: "embeddings_dimension must be 1536 for text-embedding-3-small",
		})
	}

	return errs
}

func isValidEnvironment(env string) bool {
	switch env {
	case "development", "staging", "production":
		return true
	default:
		return false
	}
}
