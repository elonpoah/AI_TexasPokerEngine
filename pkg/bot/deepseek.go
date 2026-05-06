package bot

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"poker-engine/pkg/betting"
)

type DeepSeekBot struct {
	APIKey  string
	BaseURL string
	Model   string

	Client *http.Client
}

func (b *DeepSeekBot) Name() string {
	return "deepseek"
}

// 构造
func NewDeepSeekBot(apiKey string) *DeepSeekBot {
	return &DeepSeekBot{
		APIKey:  apiKey,
		BaseURL: "https://api.deepseek.com/v1/chat/completions",
		Model:   "deepseek-chat",
		Client: &http.Client{
			Timeout: 5 * time.Second, // 防卡死
		},
	}
}

// 对外入口（统一调用）
func (b *DeepSeekBot) Act(ctx Context) betting.Action {

	// ⭐ 本地策略兜底（永远有返回）
	fallback := Decide(ctx)

	// ⭐ 是否值得调用 LLM（节省成本 + 稳定）
	// if !shouldCallLLM(ctx) {
	// 	return fallback
	// }

	action, err := b.callLLM(ctx)
	if err != nil {
		return fallback
	}

	// ⭐ 校验 AI 输出（防止乱来）
	if !validAction(action) {
		return fallback
	}

	return action
}

func shouldCallLLM(ctx Context) bool {

	// 极强 or 极弱，用规则即可
	if ctx.Strength > 0.85 || ctx.Strength < 0.15 {
		return false
	}

	// 人少更需要博弈
	if ctx.ActivePlayers <= 3 {
		return true
	}

	// 默认概率触发
	return true
}

// -------------------- 调用 DeepSeek --------------------

func (b *DeepSeekBot) callLLM(ctx Context) (betting.Action, error) {

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

	// ⭐ 加 context 超时（更安全）
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

// -------------------- 解析 AI 输出 --------------------

func parseAction(text string, ctx Context) betting.Action {

	action := betting.Action{
		SeatID: ctx.SeatID,
	}

	switch {

	case strings.HasPrefix(text, "fold"):
		action.Type = betting.Fold

	case strings.HasPrefix(text, "check"):
		action.Type = betting.Check

	case strings.HasPrefix(text, "call"):
		action.Type = betting.Call

	case strings.HasPrefix(text, "allin"):
		action.Type = betting.AllIn

	case strings.HasPrefix(text, "bet"):
		action.Type = betting.Bet
		parts := strings.Split(text, " ")
		if len(parts) >= 2 {
			if amt, err := strconv.Atoi(parts[1]); err == nil {
				action.Amount = amt
				return action
			}
		}
		// fallback：默认加注
		action.Amount = ctx.MinRaise * 2

	case strings.HasPrefix(text, "raise"):
		action.Type = betting.Raise

		// 提取金额
		parts := strings.Split(text, " ")
		if len(parts) >= 2 {
			if amt, err := strconv.Atoi(parts[1]); err == nil {
				action.Amount = amt
				return action
			}
		}

		// fallback：默认加注
		action.Amount = ctx.MaxBet + ctx.Pot/3

	default:
		// AI胡说 → fold
		action.Type = betting.Fold
	}

	return action
}

// -------------------- 校验 --------------------

func validAction(a betting.Action) bool {

	switch a.Type {
	case betting.Fold,
		betting.Call,
		betting.Check,
		betting.Bet,
		betting.Raise,
		betting.AllIn:
		return true
	}

	return false
}
