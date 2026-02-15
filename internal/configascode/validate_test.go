package configascode_test

import (
	"errors"
	"testing"

	"github.com/sable-inc/anvil/internal/configascode"
)

func validConfig() configascode.AgentConfig {
	voiceID := "voice123"
	model := "eleven_turbo_v2"
	pinecone := "my-index"
	return configascode.AgentConfig{
		Name:                 "Test Agent",
		GreetingInstructions: "Hello!",
		Environment:          "development",
		LLM: configascode.LLMConfig{
			Model:      "gpt-4o-realtime-preview",
			Modalities: []string{"text", "audio"},
			TurnDetection: configascode.TurnDetectionConfig{
				Type: "server_vad", Threshold: 0.5, SilenceDurationMs: 500,
			},
		},
		STT: configascode.STTConfig{
			Provider: "both",
			OpenAI:   configascode.OpenAISTTConfig{Enabled: true, Model: "gpt-4o-mini-transcribe-2025-12-15"},
			Deepgram: configascode.DeepgramSTTConfig{Enabled: true, Model: "nova-3", Language: "multi"},
		},
		TTS: configascode.TTSConfig{
			Provider: "elevenlabs",
			VoiceID:  &voiceID,
			Model:    &model,
		},
		Room: configascode.RoomConfig{TranscriptionEnabled: true},
		Components: configascode.ComponentsConfig{
			RAG: configascode.RAGConfig{
				Enabled:             true,
				PineconeIndexName:   &pinecone,
				EmbeddingsModel:     "text-embedding-3-small",
				EmbeddingsDimension: 1536,
				NoResultsMessage:    "No results found.",
			},
			Memory: configascode.MemoryConfig{SummarizationModel: "gpt-4o-mini"},
		},
	}
}

func TestValidate_ValidConfig(t *testing.T) {
	cfg := validConfig()
	if err := configascode.Validate(&cfg); err != nil {
		t.Errorf("expected valid config, got: %v", err)
	}
}

func TestValidate_MissingName(t *testing.T) {
	cfg := validConfig()
	cfg.Name = ""
	err := configascode.Validate(&cfg)
	if err == nil {
		t.Fatal("expected error for missing name")
	}
	assertContainsPath(t, err, "name")
}

func TestValidate_InvalidEnvironment(t *testing.T) {
	cfg := validConfig()
	cfg.Environment = "invalid"
	err := configascode.Validate(&cfg)
	if err == nil {
		t.Fatal("expected error for invalid environment")
	}
	assertContainsPath(t, err, "environment")
}

func TestValidate_ElevenLabsMissingVoiceID(t *testing.T) {
	cfg := validConfig()
	cfg.TTS.Provider = "elevenlabs"
	cfg.TTS.VoiceID = nil
	err := configascode.Validate(&cfg)
	if err == nil {
		t.Fatal("expected error for missing voice_id with elevenlabs")
	}
	assertContainsPath(t, err, "tts.voice_id")
}

func TestValidate_NoneProviderWithVoiceID(t *testing.T) {
	cfg := validConfig()
	cfg.TTS.Provider = "none"
	voice := "should-be-nil"
	cfg.TTS.VoiceID = &voice
	cfg.TTS.Model = nil
	err := configascode.Validate(&cfg)
	if err == nil {
		t.Fatal("expected error for voice_id with none provider")
	}
	assertContainsPath(t, err, "tts.voice_id")
}

func TestValidate_OpenAIProviderWithVoiceID(t *testing.T) {
	cfg := validConfig()
	cfg.TTS.Provider = "openai"
	voice := "should-be-nil"
	cfg.TTS.VoiceID = &voice
	cfg.TTS.Model = nil
	err := configascode.Validate(&cfg)
	if err == nil {
		t.Fatal("expected error for voice_id with openai provider")
	}
	assertContainsPath(t, err, "tts.voice_id")
}

func TestValidate_STTProviderMismatch(t *testing.T) {
	cfg := validConfig()
	cfg.STT.Provider = "openai"
	cfg.STT.OpenAI.Enabled = false
	err := configascode.Validate(&cfg)
	if err == nil {
		t.Fatal("expected error for disabled openai with openai provider")
	}
	assertContainsPath(t, err, "stt.openai.enabled")
}

func TestValidate_STTBothRequiresBothEnabled(t *testing.T) {
	cfg := validConfig()
	cfg.STT.Provider = "both"
	cfg.STT.OpenAI.Enabled = true
	cfg.STT.Deepgram.Enabled = false
	err := configascode.Validate(&cfg)
	if err == nil {
		t.Fatal("expected error for both provider with deepgram disabled")
	}
	assertContainsPath(t, err, "stt.provider")
}

func TestValidate_VisionProactiveRequiresEnabled(t *testing.T) {
	cfg := validConfig()
	cfg.Components.Vision.Proactive = true
	cfg.Components.Vision.Enabled = false
	err := configascode.Validate(&cfg)
	if err == nil {
		t.Fatal("expected error for proactive vision without enabled")
	}
	assertContainsPath(t, err, "components.vision.enabled")
}

func TestValidate_BrowserStreamingRequiresEnabled(t *testing.T) {
	cfg := validConfig()
	cfg.Components.Browser.Enabled = false
	cfg.Components.Browser.EnableStreaming = true
	err := configascode.Validate(&cfg)
	if err == nil {
		t.Fatal("expected error for streaming without browser enabled")
	}
	assertContainsPath(t, err, "components.browser.enable_streaming")
}

func TestValidate_RAGEnabledRequiresIndex(t *testing.T) {
	cfg := validConfig()
	cfg.Components.RAG.Enabled = true
	cfg.Components.RAG.PineconeIndexName = nil
	cfg.Components.RAG.IndexPath = nil
	err := configascode.Validate(&cfg)
	if err == nil {
		t.Fatal("expected error for enabled RAG without index")
	}
	assertContainsPath(t, err, "components.rag.index_path")
}

func TestValidate_RAGEmbeddingsDimension(t *testing.T) {
	cfg := validConfig()
	cfg.Components.RAG.EmbeddingsModel = "text-embedding-3-small"
	cfg.Components.RAG.EmbeddingsDimension = 768
	err := configascode.Validate(&cfg)
	if err == nil {
		t.Fatal("expected error for wrong embeddings dimension")
	}
	assertContainsPath(t, err, "components.rag.embeddings_dimension")
}

func assertContainsPath(t *testing.T, err error, path string) {
	t.Helper()
	var ve *configascode.ValidationErrors
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationErrors, got %T", err)
	}
	for _, e := range ve.Errors {
		if e.Path == path {
			return
		}
	}
	t.Errorf("expected validation error at path %q, got: %v", path, err)
}
