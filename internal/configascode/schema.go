// Package configascode provides the config-as-code engine for Anvil.
// It defines Go types that mirror the sable-api AgentConfig schema,
// with bidirectional YAML <-> API JSON conversion and local validation.
package configascode

// AgentConfig is the top-level agent configuration.
// Field names use snake_case to match the API JSON and YAML format.
type AgentConfig struct {
	Name                 string             `json:"name" yaml:"name"`
	GreetingInstructions string             `json:"greeting_instructions" yaml:"greeting_instructions"`
	Environment          string             `json:"environment" yaml:"environment"`
	CustomTools          bool               `json:"custom_tools" yaml:"custom_tools"`
	BuiltinTools         map[string]bool    `json:"builtin_tools,omitempty" yaml:"builtin_tools,omitempty"`
	Instructions         *InstructionsConfig `json:"instructions,omitempty" yaml:"instructions,omitempty"`
	LLM                  LLMConfig          `json:"llm" yaml:"llm"`
	STT                  STTConfig          `json:"stt" yaml:"stt"`
	TTS                  TTSConfig          `json:"tts" yaml:"tts"`
	Room                 RoomConfig         `json:"room" yaml:"room"`
	Components           ComponentsConfig   `json:"components" yaml:"components"`
}

// InstructionsConfig holds system prompts and customer stages.
type InstructionsConfig struct {
	SableSystemPrompt    string  `json:"sable_system_prompt" yaml:"sable_system_prompt"`
	CustomerSystemPrompt string  `json:"customer_system_prompt" yaml:"customer_system_prompt"`
	CustomerStages       []Stage `json:"customer_stages" yaml:"customer_stages"`
	ProactiveVisionPrompt string `json:"proactive_vision_prompt" yaml:"proactive_vision_prompt"`
}

// Stage is a named instruction stage.
type Stage struct {
	Name        string `json:"name" yaml:"name"`
	Instruction string `json:"instruction" yaml:"instruction"`
}

// LLMConfig configures the language model.
type LLMConfig struct {
	Model         string              `json:"model" yaml:"model"`
	Modalities    []string            `json:"modalities" yaml:"modalities"`
	Temperature   float64             `json:"temperature" yaml:"temperature"`
	TurnDetection TurnDetectionConfig `json:"turn_detection" yaml:"turn_detection"`
	Speed         *float64            `json:"speed,omitempty" yaml:"speed,omitempty"`
}

// TurnDetectionConfig configures server-side voice activity detection.
type TurnDetectionConfig struct {
	Type              string `json:"type" yaml:"type"`
	Threshold         float64 `json:"threshold" yaml:"threshold"`
	SilenceDurationMs int    `json:"silence_duration_ms" yaml:"silence_duration_ms"`
}

// STTConfig configures speech-to-text.
type STTConfig struct {
	Provider string           `json:"provider" yaml:"provider"`
	OpenAI   OpenAISTTConfig  `json:"openai" yaml:"openai"`
	Deepgram DeepgramSTTConfig `json:"deepgram" yaml:"deepgram"`
}

// OpenAISTTConfig configures the OpenAI STT provider.
type OpenAISTTConfig struct {
	Enabled bool   `json:"enabled" yaml:"enabled"`
	Model   string `json:"model" yaml:"model"`
}

// DeepgramSTTConfig configures the Deepgram STT provider.
type DeepgramSTTConfig struct {
	Enabled        bool   `json:"enabled" yaml:"enabled"`
	Model          string `json:"model" yaml:"model"`
	Language       string `json:"language" yaml:"language"`
	InterimResults bool   `json:"interim_results" yaml:"interim_results"`
	Punctuate      bool   `json:"punctuate" yaml:"punctuate"`
	SmartFormat    bool   `json:"smart_format" yaml:"smart_format"`
}

// TTSConfig configures text-to-speech.
type TTSConfig struct {
	Provider string  `json:"provider" yaml:"provider"`
	VoiceID  *string `json:"voice_id" yaml:"voice_id"`
	Model    *string `json:"model" yaml:"model"`
}

// RoomConfig configures the LiveKit room.
type RoomConfig struct {
	NoiseCancellation    bool `json:"noise_cancellation" yaml:"noise_cancellation"`
	VideoEnabled         bool `json:"video_enabled" yaml:"video_enabled"`
	TranscriptionEnabled bool `json:"transcription_enabled" yaml:"transcription_enabled"`
}

// ComponentsConfig groups all optional component configurations.
type ComponentsConfig struct {
	Browser             BrowserConfig             `json:"browser" yaml:"browser"`
	Vision              VisionConfig              `json:"vision" yaml:"vision"`
	Transcription       TranscriptionConfig       `json:"transcription" yaml:"transcription"`
	TranscriptionLogger TranscriptionLoggerConfig `json:"transcription_logger" yaml:"transcription_logger"`
	Memory              MemoryConfig              `json:"memory" yaml:"memory"`
	RAG                 RAGConfig                 `json:"rag" yaml:"rag"`
	PIP                 PIPConfig                 `json:"pip" yaml:"pip"`
}

// BrowserConfig configures browser-based interaction.
type BrowserConfig struct {
	Enabled         bool    `json:"enabled" yaml:"enabled"`
	EnableStreaming  bool    `json:"enable_streaming" yaml:"enable_streaming"`
	MCPServerDir    *string `json:"mcp_server_dir" yaml:"mcp_server_dir"`
}

// VisionConfig configures the vision component.
type VisionConfig struct {
	Enabled   bool    `json:"enabled" yaml:"enabled"`
	Proactive bool    `json:"proactive" yaml:"proactive"`
	Threshold float64 `json:"threshold" yaml:"threshold"`
	Cooldown  int     `json:"cooldown" yaml:"cooldown"`
	DebugLogs bool    `json:"debug_logs" yaml:"debug_logs"`
}

// TranscriptionConfig configures live transcription.
type TranscriptionConfig struct {
	Enabled                 bool    `json:"enabled" yaml:"enabled"`
	UtteranceFinalizeDelay  float64 `json:"utterance_finalize_delay" yaml:"utterance_finalize_delay"`
	TranscriptPrefix        *string `json:"transcript_prefix" yaml:"transcript_prefix"`
}

// TranscriptionLoggerConfig configures transcript logging.
type TranscriptionLoggerConfig struct {
	Enabled  bool    `json:"enabled" yaml:"enabled"`
	ModuleID *string `json:"module_id" yaml:"module_id"`
	OrgID    *string `json:"org_id" yaml:"org_id"`
}

// MemoryConfig configures conversation memory.
type MemoryConfig struct {
	Enabled             bool   `json:"enabled" yaml:"enabled"`
	SummarizationModel  string `json:"summarization_model" yaml:"summarization_model"`
}

// RAGConfig configures retrieval-augmented generation.
type RAGConfig struct {
	Enabled             bool    `json:"enabled" yaml:"enabled"`
	PineconeIndexName   *string `json:"pinecone_index_name" yaml:"pinecone_index_name"`
	MinScore            float64 `json:"min_score" yaml:"min_score"`
	IndexPath           *string `json:"index_path" yaml:"index_path"`
	RRFK                int     `json:"rrf_k" yaml:"rrf_k"`
	SearchTopN          int     `json:"search_top_n" yaml:"search_top_n"`
	SearchK             int     `json:"search_k" yaml:"search_k"`
	EmbeddingsModel     string  `json:"embeddings_model" yaml:"embeddings_model"`
	EmbeddingsDimension int     `json:"embeddings_dimension" yaml:"embeddings_dimension"`
	ResultLimit         int     `json:"result_limit" yaml:"result_limit"`
	NoResultsMessage    string  `json:"no_results_message" yaml:"no_results_message"`
}

// PIPConfig configures picture-in-picture mode.
type PIPConfig struct {
	Enabled bool `json:"enabled" yaml:"enabled"`
}

// ConfigFile is a YAML config file wrapper that includes metadata
// alongside the agent configuration.
type ConfigFile struct {
	Agent   string      `json:"agent,omitempty" yaml:"agent,omitempty"`
	OrgID   int         `json:"org_id,omitempty" yaml:"org_id,omitempty"`
	Config  AgentConfig `json:"config" yaml:"config"`
}
