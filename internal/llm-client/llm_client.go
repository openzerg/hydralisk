package llmclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/openzerg/hydralisk/internal/core/types"
)

type Client struct {
	config *types.LLMConfig
	client *http.Client
}

func NewClient(config *types.LLMConfig) *Client {
	return &Client{
		config: config,
		client: &http.Client{},
	}
}

func (c *Client) Complete(ctx context.Context, messages []*types.LLMMessage, tools []*types.ToolDefinition, options map[string]interface{}) (*types.ChatCompletionResponse, error) {
	msgs := make([]types.LLMMessage, len(messages))
	for i, m := range messages {
		msgs[i] = *m
	}

	chatTools := make([]*types.ChatCompletionTool, len(tools))
	for i, t := range tools {
		params := make(map[string]interface{})
		for k, v := range t.Function.Parameters.Properties {
			params[k] = v
		}
		chatTools[i] = &types.ChatCompletionTool{
			Type: t.Type,
			Function: &types.ChatCompletionToolFunction{
				Name:        t.Function.Name,
				Description: t.Function.Description,
				Parameters:  params,
			},
		}
	}

	req := &types.ChatCompletionRequest{
		Model:    c.config.Model,
		Messages: msgs,
		Tools:    chatTools,
	}

	if c.config.MaxTokens != nil {
		req.MaxTokens = c.config.MaxTokens
	}
	if c.config.Temperature != nil {
		req.Temperature = c.config.Temperature
	}
	if c.config.TopP != nil {
		req.TopP = c.config.TopP
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/chat/completions", c.config.BaseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if c.config.APIKey != "" {
		httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result types.ChatCompletionResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) Stream(ctx context.Context, messages []*types.LLMMessage, tools []*types.ToolDefinition, options map[string]interface{}) (<-chan *types.StreamChunk, error) {
	ch := make(chan *types.StreamChunk)

	go func() {
		defer close(ch)
		// Simplified implementation - just call Complete and return as single chunk
		resp, err := c.Complete(ctx, messages, tools, options)
		if err != nil {
			return
		}

		for _, choice := range resp.Choices {
			var role *string
			if choice.Message != nil {
				r := string(choice.Message.Role)
				role = &r
			}
			ch <- &types.StreamChunk{
				ID:      resp.ID,
				Model:   resp.Model,
				Created: resp.Created,
				Choices: []*types.StreamChunkChoice{
					{
						Index: choice.Index,
						Delta: &types.StreamChunkDelta{
							Role:    role,
							Content: choice.Message.Content,
						},
						FinishReason: choice.FinishReason,
					},
				},
			}
		}
	}()

	return ch, nil
}

func (c *Client) UpdateConfig(config map[string]interface{}) {
	if baseURL, ok := config["base_url"].(string); ok {
		c.config.BaseURL = baseURL
	}
	if apiKey, ok := config["api_key"].(string); ok {
		c.config.APIKey = apiKey
	}
	if model, ok := config["model"].(string); ok {
		c.config.Model = model
	}
}

func (c *Client) GetConfig() *types.LLMConfig {
	return c.config
}

func (c *Client) CountTokens(messages []*types.LLMMessage) int {
	count := 0
	for _, msg := range messages {
		if msg.Content != nil {
			count += len(*msg.Content)
		}
	}
	return count
}
