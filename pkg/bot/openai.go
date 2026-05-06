package bot

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"poker-engine/pkg/betting"
)

type OpenAIBot struct {
	APIKey  string
	BaseURL string
	Model   string
	Client  *http.Client
}

func (b *OpenAIBot) Name() string {
	return "openai"
}

func NewOpenAIBot(apiKey string) *OpenAIBot {
	return &OpenAIBot{
		APIKey:  apiKey,
		BaseURL: "https://api.openai.com/v1/chat/completions",
		Model:   "gpt-4o-mini", // 成本低、够用
		Client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (b *OpenAIBot) Act(ctx Context) betting.Action {

	// ✅ 本地兜底
	fallback := Decide(ctx)

	if !shouldCallLLM(ctx) {
		return fallback
	}

	action, err := b.callLLM(ctx)
	if err != nil {
		return fallback
	}

	if !validAction(action) {
		return fallback
	}

	return action
}

// -------------------- 调用 OpenAI --------------------

func (b *OpenAIBot) callLLM(ctx Context) (betting.Action, error) {

	prompt := buildPrompt(ctx)

	body := map[string]interface{}{
		"model": b.Model,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.7,
	}

	data, _ := json.Marshal(body)

	req, err := http.NewRequest("POST", b.BaseURL, bytes.NewBuffer(data))
	if err != nil {
		return betting.Action{}, err
	}

	req.Header.Set("Authorization", "Bearer "+b.APIKey)
	req.Header.Set("Content-Type", "application/json")

	ctxHttp, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req = req.WithContext(ctxHttp)

	resp, err := b.Client.Do(req)
	if err != nil {
		return betting.Action{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return betting.Action{}, err
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			}
		}
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return betting.Action{}, err
	}

	if len(result.Choices) == 0 {
		return betting.Action{}, err
	}

	text := strings.ToLower(strings.TrimSpace(result.Choices[0].Message.Content))

	return parseAction(text, ctx), nil
}
