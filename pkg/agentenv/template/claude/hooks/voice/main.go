// nolint
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/hasansino/go42x/pkg/claudehook"
)

const (
	hookName           = "voice-tts-hook"
	elevenLabsAPIBase  = "https://api.elevenlabs.io/v1"
	defaultHttpTimeout = 10 * time.Second
	defaultBufferSize  = 4096
	defaultVoiceID     = "21m00Tcm4TlvDq8ikWAM"
	ModelHighQuality   = "high_quality"
	ModelBalanced      = "balanced"
	ModelLowLatency    = "low_latency"
)

var (
	elevenLabsAPIEndpoint = getEnvOrDefault(os.Getenv("GO42X_ELEVEN_LABS_API_ENDPOINT"), elevenLabsAPIBase)
	elevenLabsAPITimeout  = getEnvOrDefault(os.Getenv("GO42X_ELEVEN_LABS_API_TIMEOUT"), "")
	elevenLabsAPIKey      = getEnvOrDefault(os.Getenv("GO42X_ELEVEN_LABS_API_KEY"), "")
	elevenLabsBufferSize  = getEnvOrDefault(os.Getenv("GO42X_ELEVEN_LABS_BUFFER_SIZE"), "")
	elevenLabsVoiceID     = getEnvOrDefault(os.Getenv("GO42X_ELEVEN_LABS_VOICE_ID"), defaultVoiceID)
	elevenLabsModel       = getEnvOrDefault(os.Getenv("GO42X_ELEVEN_LABS_PRESET"), ModelBalanced)
)

// Default voice presets for different use cases
var voicePresets = map[string]VoiceConfig{
	ModelHighQuality: {
		VoiceID:         defaultVoiceID,
		ModelID:         "eleven_multilingual_v2",
		Stability:       0.75,
		Similarity:      0.85,
		Style:           0.25,
		SpeakerBoost:    true,
		OutputFormat:    "mp3_44100_128",
		SampleRate:      44100,
		BitRate:         "128k",
		OptimizeLatency: 0,
	},
	ModelBalanced: {
		VoiceID:         defaultVoiceID,
		ModelID:         "eleven_turbo_v2_5",
		Stability:       0.5,
		Similarity:      0.75,
		Style:           0.0,
		SpeakerBoost:    true,
		OutputFormat:    "mp3_22050_32",
		SampleRate:      22050,
		BitRate:         "32k",
		OptimizeLatency: 2,
	},
	ModelLowLatency: {
		VoiceID:         defaultVoiceID,
		ModelID:         "eleven_turbo_v2_5",
		Stability:       0.5,
		Similarity:      0.75,
		Style:           0.0,
		SpeakerBoost:    false,
		OutputFormat:    "mp3_22050_32",
		SampleRate:      22050,
		BitRate:         "32k",
		OptimizeLatency: 4,
	},
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := slog.New(slog.DiscardHandler)
	logPath := filepath.Join(claudehook.GetProjectDir(), "claude_tts_hook.log")
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err == nil {
		logger = slog.New(slog.NewTextHandler(file, &slog.HandlerOptions{AddSource: false}))
		defer file.Close()
	}

	hook, err := claudehook.New(
		hookName,
		newVoiceProcessor(logger),
		claudehook.WithLogger(logger),
	)
	if err != nil {
		logger.Error("Failed to create hook", "error", err)
		os.Exit(1)
	}

	if err := hook.Run(ctx); err != nil {
		logger.Error("Hook execution failed", "error", err)
		os.Exit(1)
	}
}

// ---- AudioPlayer

// AudioPlayer handles audio playback
type AudioPlayer struct {
	logger *slog.Logger
}

// NewAudioPlayer creates a new audio player
func NewAudioPlayer(logger *slog.Logger) *AudioPlayer {
	return &AudioPlayer{logger: logger}
}

// Play plays audio from stream
func (p *AudioPlayer) Play(stream io.Reader) error {
	player, args := p.detectAudioPlayer()
	if player == "" {
		return fmt.Errorf("no audio player found")
	}

	p.logger.Info("Using audio player", "player", player)

	// Handle players that need temp files
	if player == "afplay" {
		return p.playWithTempFile(player, stream)
	}

	// Stream to stdin-capable players
	return p.streamToPlayer(player, args, stream)
}

// detectAudioPlayer finds available audio player
func (p *AudioPlayer) detectAudioPlayer() (string, []string) {
	players := []struct {
		cmd  string
		args []string
	}{
		{"ffplay", []string{"-nodisp", "-autoexit", "-loglevel", "quiet", "-"}},
		{"mpv", []string{"--no-video", "--really-quiet", "-"}},
		{"vlc", []string{"--intf", "dummy", "--play-and-exit", "-"}},
		{"sox", []string{"-q", "-t", "mp3", "-"}},
		{"aplay", []string{"-q", "-"}},
	}

	for _, player := range players {
		if _, err := exec.LookPath(player.cmd); err == nil {
			return player.cmd, player.args
		}
	}

	// Check for macOS afplay
	if runtime.GOOS == "darwin" {
		if _, err := exec.LookPath("afplay"); err == nil {
			return "afplay", nil
		}
	}

	return "", nil
}

// streamToPlayer streams audio to player stdin
func (p *AudioPlayer) streamToPlayer(cmd string, args []string, stream io.Reader) error {
	player := exec.Command(cmd, args...)
	stdin, err := player.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	if err := player.Start(); err != nil {
		return fmt.Errorf("failed to start player: %w", err)
	}

	bufferSize := defaultBufferSize
	if val, err := strconv.Atoi(elevenLabsBufferSize); err == nil && val > 0 {
		bufferSize = val
	}

	// Stream audio data
	go func() {
		defer stdin.Close()
		buf := make([]byte, bufferSize)
		totalBytes := 0

		for {
			n, err := stream.Read(buf)
			if n > 0 {
				stdin.Write(buf[:n])
				totalBytes += n
			}
			if err == io.EOF {
				p.logger.Debug("Stream complete", "bytes", totalBytes)
				break
			}
			if err != nil {
				p.logger.Error("Stream error", "error", err)
				break
			}
		}
	}()

	return player.Wait()
}

// playWithTempFile plays audio using temp file
func (p *AudioPlayer) playWithTempFile(cmd string, stream io.Reader) error {
	tmpFile, err := os.CreateTemp("", "tts_*.mp3")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write stream to temp file
	written, err := io.Copy(tmpFile, stream)
	tmpFile.Close()
	if err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	p.logger.Debug("Written to temp file", "bytes", written, "file", tmpFile.Name())

	// Play temp file
	player := exec.Command(cmd, tmpFile.Name())
	return player.Run()
}

// ---- TTSService

// TTSRequest represents the ElevenLabs API request
type TTSRequest struct {
	Text          string         `json:"text"`
	ModelID       string         `json:"model_id"`
	VoiceSettings *VoiceSettings `json:"voice_settings,omitempty"`
}

// VoiceSettings for fine-tuning voice output
type VoiceSettings struct {
	Stability       float32 `json:"stability"`
	SimilarityBoost float32 `json:"similarity_boost"`
	Style           float32 `json:"style,omitempty"`
	UseSpeakerBoost bool    `json:"use_speaker_boost,omitempty"`
}

// VoiceConfig holds configuration for TTS
type VoiceConfig struct {
	VoiceID         string
	ModelID         string
	Stability       float32
	Similarity      float32
	Style           float32
	SpeakerBoost    bool
	OutputFormat    string
	SampleRate      int
	BitRate         string
	OptimizeLatency int
}

// TTSService handles text-to-speech operations
type TTSService struct {
	apiKey string
	config VoiceConfig
	logger *slog.Logger
	client *http.Client
}

// NewTTSService creates a new TTS service instance
func NewTTSService(logger *slog.Logger) *TTSService {
	config := voicePresets[elevenLabsModel]

	if elevenLabsVoiceID != "" {
		config.VoiceID = elevenLabsVoiceID
	}

	timeout := defaultHttpTimeout
	if elevenLabsAPITimeout != "" {
		if t, err := strconv.Atoi(elevenLabsAPITimeout); err == nil && t > 0 {
			timeout = time.Duration(t) * time.Second
		}
	}

	return &TTSService{
		apiKey: elevenLabsAPIKey,
		config: config,
		logger: logger,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// ProcessText converts text to speech
func (s *TTSService) ProcessText(text string) error {
	if text == "" {
		s.logger.Debug("No text to process")
		return nil
	}

	s.logger.Info("Processing text", "characters", len(text))

	// Create TTS request
	audioStream, err := s.streamTTS(text)
	if err != nil {
		return fmt.Errorf("TTS streaming failed: %w", err)
	}
	defer audioStream.Close()

	// Play audio
	player := NewAudioPlayer(s.logger)
	return player.Play(audioStream)
}

// streamTTS makes the API request and returns audio stream
func (s *TTSService) streamTTS(text string) (io.ReadCloser, error) {
	url := fmt.Sprintf("%s/text-to-speech/%s/stream", elevenLabsAPIEndpoint, s.config.VoiceID)

	// Prepare request body
	reqBody := TTSRequest{
		Text:    text,
		ModelID: s.config.ModelID,
		VoiceSettings: &VoiceSettings{
			Stability:       s.config.Stability,
			SimilarityBoost: s.config.Similarity,
			Style:           s.config.Style,
			UseSpeakerBoost: s.config.SpeakerBoost,
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "audio/mpeg")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("xi-api-key", s.apiKey)

	// Add query parameters based on model
	q := req.URL.Query()
	q.Add("output_format", s.config.OutputFormat)

	// Only add optimize_streaming_latency for turbo models
	if strings.Contains(s.config.ModelID, "turbo") && s.config.OptimizeLatency > 0 {
		q.Add("optimize_streaming_latency", fmt.Sprintf("%d", s.config.OptimizeLatency))
	}

	req.URL.RawQuery = q.Encode()

	s.logger.Debug("Making TTS request", "url", url)
	s.logger.Debug("TTS config", "model", s.config.ModelID, "format", s.config.OutputFormat)

	// Execute request
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	s.logger.Info("TTS request successful")
	return resp.Body, nil
}

// ---- VoiceProcessor

// VoiceProcessor implements the claudehook.Processor interface
type VoiceProcessor struct {
	logger   *slog.Logger
	tts      *TTSService
	parser   *claudehook.TranscriptParser
	textProc *claudehook.TextProcessor
}

// newVoiceProcessor creates a new voice processor
func newVoiceProcessor(logger *slog.Logger) *VoiceProcessor {
	return &VoiceProcessor{
		logger:   logger,
		parser:   claudehook.NewTranscriptParser(),
		textProc: claudehook.NewTextProcessor(true, true),
	}
}

// Process implements the hook processing logic
func (vp *VoiceProcessor) Process(input *claudehook.Input) error {
	// Initialize TTS service with logger from context
	vp.tts = NewTTSService(vp.logger)

	// Extract text from transcript
	text, err := vp.extractText(input)
	if err != nil {
		return fmt.Errorf("failed to extract text: %w", err)
	}

	if text == "" {
		vp.logger.Info("No text to speak")
		return nil
	}

	vp.logger.Info("Text to speak", "characters", len(text))

	// Process text to speech
	if err := vp.tts.ProcessText(text); err != nil {
		return fmt.Errorf("TTS processing failed: %w", err)
	}

	return nil
}

// extractText gets text from transcript
func (vp *VoiceProcessor) extractText(input *claudehook.Input) (string, error) {
	if input.TranscriptPath == "" {
		vp.logger.Debug("No transcript path provided")
		return "", nil
	}

	vp.logger.Debug("Reading transcript", "path", input.TranscriptPath)
	text, err := vp.parser.ExtractLastAssistantMessage(input.TranscriptPath)
	if err != nil {
		return "", fmt.Errorf("failed to read transcript: %w", err)
	}

	// Process text to remove markdown and code blocks
	text = vp.textProc.Process(text)

	return text, nil
}

// ----

func getEnvOrDefault(envVar, defaultVal string) string {
	if val := os.Getenv(envVar); val != "" {
		return val
	}
	return defaultVal
}
