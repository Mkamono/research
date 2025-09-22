package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"research/mcp/ask-me/internal/app"
)

var _ app.ChatProvider = (*slack)(nil)

type slack struct {
	oAuthToken string
	channel    string
	channelID  string // 実際のチャンネルID
	client     *http.Client
	replies    map[string][]string
	timeout    time.Duration
}

func NewChatProvider(oAuthToken, channel string) *slack {
	return &slack{
		oAuthToken: oAuthToken,
		channel:    channel,
		client:     &http.Client{Timeout: 30 * time.Second},
		replies:    make(map[string][]string),
		timeout:    24 * time.Hour,
	}
}

func (s *slack) Chat(ctx context.Context, req app.ChatRequest) (app.ChatResponse, error) {
	threadID, err := s.sendMessage(ctx, req.Message, req.ThreadID)
	if err != nil {
		return app.ChatResponse{}, fmt.Errorf("failed to send message: %w", err)
	}

	reply, err := s.waitForReply(ctx, threadID, s.timeout)
	if err != nil {
		return app.ChatResponse{}, fmt.Errorf("failed to wait for reply: %w", err)
	}

	return app.ChatResponse{
		Message:  reply,
		ThreadID: threadID,
	}, nil
}

func (s *slack) GetThreadHistory(ctx context.Context, threadID string) (app.GetThreadHistoryResponse, error) {
	messages, err := s.getThreadMessages(ctx, threadID)
	if err != nil {
		return app.GetThreadHistoryResponse{}, fmt.Errorf("failed to get thread history: %w", err)
	}

	return app.GetThreadHistoryResponse{
		ThreadID: threadID,
		Messages: messages,
	}, nil
}

func (s *slack) sendMessage(ctx context.Context, message string, threadID *string) (string, error) {
	payload := map[string]interface{}{
		"channel": s.channel,
		"text":    message,
	}

	if threadID != nil {
		payload["thread_ts"] = *threadID
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://slack.com/api/chat.postMessage", bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+s.oAuthToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var slackResp struct {
		OK      bool   `json:"ok"`
		TS      string `json:"ts"`
		Error   string `json:"error"`
		Channel string `json:"channel"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&slackResp); err != nil {
		return "", err
	}

	if !slackResp.OK {
		return "", fmt.Errorf("slack API error: %s", slackResp.Error)
	}

	// チャンネルIDを保存
	s.channelID = slackResp.Channel

	if threadID != nil {
		return *threadID, nil
	}
	return slackResp.TS, nil
}

func (s *slack) waitForReply(ctx context.Context, threadID string, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	initialCount := len(s.replies[threadID])

	for {
		select {
		case <-ctx.Done():
			return "", fmt.Errorf("timeout waiting for reply")
		case <-ticker.C:
			messages, err := s.getThreadMessages(ctx, threadID)
			if err != nil {
				continue
			}

			if len(messages) > initialCount {
				s.replies[threadID] = messages
				return messages[len(messages)-1], nil
			}
		}
	}
}

func (s *slack) getThreadMessages(ctx context.Context, threadID string) ([]string, error) {
	// 実際のチャンネルIDを使用（sendMessageで取得済み）
	channelToUse := s.channelID
	if channelToUse == "" {
		channelToUse = strings.TrimPrefix(s.channel, "#")
	}
	url := fmt.Sprintf("https://slack.com/api/conversations.replies?channel=%s&ts=%s", channelToUse, threadID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+s.oAuthToken)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var slackResp struct {
		OK       bool   `json:"ok"`
		Error    string `json:"error"`
		Messages []struct {
			Text    string `json:"text"`
			TS      string `json:"ts"`
			User    string `json:"user"`
			BotID   string `json:"bot_id"`
			Subtype string `json:"subtype"`
		} `json:"messages"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&slackResp); err != nil {
		return nil, err
	}

	if !slackResp.OK {
		return nil, fmt.Errorf("slack API error: %s", slackResp.Error)
	}

	var messages []string
	for _, msg := range slackResp.Messages {
		// ボットメッセージを除外（bot_idがあるか、subtypeがbot_messageの場合）
		if msg.BotID != "" || msg.Subtype == "bot_message" {
			continue
		}
		messages = append(messages, msg.Text)
	}

	return messages, nil
}

func parseSlackTimestamp(ts string) (int64, error) {
	parts := strings.Split(ts, ".")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid timestamp format")
	}

	seconds, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, err
	}

	return seconds, nil
}
