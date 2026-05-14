package deepseek

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"ai-interview/backend/internal/ai"
	"ai-interview/backend/internal/model"
)

type Config struct {
	APIKey  string
	BaseURL string
	Model   string
	Timeout time.Duration
}

type Provider struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

func NewProvider(cfg Config) (*Provider, error) {
	if strings.TrimSpace(cfg.APIKey) == "" {
		return nil, errors.New("deepseek api key is required")
	}

	baseURL := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if baseURL == "" {
		baseURL = "https://api.deepseek.com"
	}

	modelName := strings.TrimSpace(cfg.Model)
	if modelName == "" {
		modelName = "deepseek-v4-flash"
	}

	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 60 * time.Second
	}

	return &Provider{
		apiKey:  cfg.APIKey,
		baseURL: baseURL + "/chat/completions",
		model:   modelName,
		client: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

func (p *Provider) Name() string {
	return "deepseek"
}

func (p *Provider) GenerateQuestions(ctx context.Context, input ai.GenerateQuestionsInput) ([]model.InterviewQuestion, error) {
	systemPrompt := `你是一名中文技术面试官。请根据岗位、轮次、难度和风格，输出一组高质量中文面试题。

你必须返回 JSON，格式如下：
{
  "questions": [
    {
      "questionType": "self_intro|project|system_design|go_runtime|behavior",
      "assessmentDimension": "communication|depth|accuracy|clarity",
      "prompt": "中文题目",
      "expectedPoints": ["中文要点1", "中文要点2"]
    }
  ]
}

要求：
1. questions 数量等于 questionCount。
2. 只能返回 JSON，不要返回解释。
3. expectedPoints 必须是中文短语数组。`

	var payload struct {
		Questions []struct {
			QuestionType        string   `json:"questionType"`
			AssessmentDimension string   `json:"assessmentDimension"`
			Prompt              string   `json:"prompt"`
			ExpectedPoints      []string `json:"expectedPoints"`
		} `json:"questions"`
	}
	if err := p.requestJSON(ctx, systemPrompt, input, &payload); err != nil {
		return nil, err
	}

	now := time.Now()
	questions := make([]model.InterviewQuestion, 0, len(payload.Questions))
	for index, item := range payload.Questions {
		questions = append(questions, model.InterviewQuestion{
			QuestionNo:          index + 1,
			QuestionType:        strings.TrimSpace(item.QuestionType),
			AssessmentDimension: strings.TrimSpace(item.AssessmentDimension),
			Prompt:              strings.TrimSpace(item.Prompt),
			ExpectedPoints:      compactStrings(item.ExpectedPoints),
			Source:              p.Name(),
			IsFollowUp:          false,
			CreatedAt:           now,
		})
	}

	return questions, nil
}

func (p *Provider) EvaluateAnswer(ctx context.Context, input ai.EvaluateAnswerInput) (model.AnswerFeedback, error) {
	systemPrompt := `你是一名严谨但友好的中文技术面试官。请根据题目、考察维度、回答内容和上下文输出结构化评分与点评。

你必须返回 JSON，格式如下：
{
  "overallScore": 0,
  "accuracyScore": 0,
  "clarityScore": 0,
  "depthScore": 0,
  "communicationScore": 0,
  "strengths": "中文",
  "issues": "中文",
  "improvementSuggestion": "中文",
  "followUpQuestion": "中文"
}

要求：
1. 所有分数范围 0 到 100。
2. 所有评价都必须是简体中文。
3. 只能返回 JSON，不要返回解释。`

	var payload struct {
		OverallScore          float64 `json:"overallScore"`
		AccuracyScore         float64 `json:"accuracyScore"`
		ClarityScore          float64 `json:"clarityScore"`
		DepthScore            float64 `json:"depthScore"`
		CommunicationScore    float64 `json:"communicationScore"`
		Strengths             string  `json:"strengths"`
		Issues                string  `json:"issues"`
		ImprovementSuggestion string  `json:"improvementSuggestion"`
		FollowUpQuestion      string  `json:"followUpQuestion"`
	}
	if err := p.requestJSON(ctx, systemPrompt, input, &payload); err != nil {
		return model.AnswerFeedback{}, err
	}

	return model.AnswerFeedback{
		OverallScore:          clampScore(payload.OverallScore),
		AccuracyScore:         clampScore(payload.AccuracyScore),
		ClarityScore:          clampScore(payload.ClarityScore),
		DepthScore:            clampScore(payload.DepthScore),
		CommunicationScore:    clampScore(payload.CommunicationScore),
		Strengths:             strings.TrimSpace(payload.Strengths),
		Issues:                strings.TrimSpace(payload.Issues),
		ImprovementSuggestion: strings.TrimSpace(payload.ImprovementSuggestion),
		FollowUpQuestion:      strings.TrimSpace(payload.FollowUpQuestion),
		CreatedAt:             time.Now(),
		Source:                p.Name(),
	}, nil
}

func (p *Provider) GenerateReport(ctx context.Context, input ai.GenerateReportInput) (model.InterviewReport, error) {
	systemPrompt := `你是一名中文技术面试复盘教练。请基于整场问答历史、逐题评分和题目内容生成结构化复盘报告。

你必须返回 JSON，格式如下：
{
  "overallScore": 0,
  "performanceSummary": "中文",
  "strengthsSummary": "中文",
  "weaknessSummary": "中文",
  "recommendedActions": "中文",
  "nextPlan": {
    "focusAreas": ["中文", "中文"],
    "recommendedQuestionCount": 0
  },
  "dimensionScores": [
    {
      "dimensionCode": "accuracy|clarity|depth|communication",
      "score": 0,
      "comment": "中文"
    }
  ]
}

要求：
1. 所有分数范围 0 到 100。
2. 所有输出都必须是简体中文。
3. 只能返回 JSON，不要返回解释。`

	var payload struct {
		OverallScore       float64                      `json:"overallScore"`
		PerformanceSummary string                       `json:"performanceSummary"`
		StrengthsSummary   string                       `json:"strengthsSummary"`
		WeaknessSummary    string                       `json:"weaknessSummary"`
		RecommendedActions string                       `json:"recommendedActions"`
		NextPlan           map[string]any               `json:"nextPlan"`
		DimensionScores    []model.ReportDimensionScore `json:"dimensionScores"`
	}
	if err := p.requestJSON(ctx, systemPrompt, input, &payload); err != nil {
		return model.InterviewReport{}, err
	}

	return model.InterviewReport{
		SessionID:          input.Session.ID,
		OverallScore:       clampScore(payload.OverallScore),
		PerformanceSummary: strings.TrimSpace(payload.PerformanceSummary),
		StrengthsSummary:   strings.TrimSpace(payload.StrengthsSummary),
		WeaknessSummary:    strings.TrimSpace(payload.WeaknessSummary),
		RecommendedActions: strings.TrimSpace(payload.RecommendedActions),
		NextPlan:           sanitizeNextPlan(payload.NextPlan),
		DimensionScores:    sanitizeDimensionScores(payload.DimensionScores),
		CreatedAt:          time.Now(),
		Source:             p.Name(),
	}, nil
}

func (p *Provider) requestJSON(ctx context.Context, systemPrompt string, input any, out any) error {
	inputJSON, err := json.MarshalIndent(input, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal provider input: %w", err)
	}

	requestBody := map[string]any{
		"model": p.model,
		"messages": []map[string]any{
			{
				"role":    "system",
				"content": systemPrompt,
			},
			{
				"role":    "user",
				"content": "请根据下面的 JSON 上下文完成任务，并返回合法 json：\n\n" + string(inputJSON),
			},
		},
		"response_format": map[string]any{
			"type": "json_object",
		},
	}

	payload, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("marshal deepseek request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("create deepseek request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("request deepseek chat completions api: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read deepseek response: %w", err)
	}

	if resp.StatusCode >= 300 {
		return fmt.Errorf("deepseek api returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	content, err := extractContent(body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal([]byte(content), out); err != nil {
		return fmt.Errorf("decode deepseek json output: %w; raw output: %s", err, content)
	}

	return nil
}

func extractContent(body []byte) (string, error) {
	var response struct {
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("decode deepseek response envelope: %w", err)
	}

	if response.Error != nil && strings.TrimSpace(response.Error.Message) != "" {
		return "", errors.New(response.Error.Message)
	}
	if len(response.Choices) == 0 {
		return "", fmt.Errorf("deepseek response did not contain choices: %s", strings.TrimSpace(string(body)))
	}

	content := strings.TrimSpace(response.Choices[0].Message.Content)
	if content == "" {
		return "", fmt.Errorf("deepseek response returned empty content: %s", strings.TrimSpace(string(body)))
	}

	return content, nil
}

func clampScore(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 100 {
		return 100
	}
	return value
}

func compactStrings(items []string) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		out = append(out, item)
	}
	return out
}

func sanitizeNextPlan(plan map[string]any) map[string]any {
	if plan == nil {
		return map[string]any{
			"focusAreas":               []string{},
			"recommendedQuestionCount": 0,
		}
	}

	rawAreas, ok := plan["focusAreas"]
	if ok {
		switch items := rawAreas.(type) {
		case []any:
			areas := make([]string, 0, len(items))
			for _, item := range items {
				if text, textOK := item.(string); textOK && strings.TrimSpace(text) != "" {
					areas = append(areas, strings.TrimSpace(text))
				}
			}
			plan["focusAreas"] = areas
		case []string:
			plan["focusAreas"] = compactStrings(items)
		}
	}

	if value, ok := plan["recommendedQuestionCount"].(float64); ok {
		plan["recommendedQuestionCount"] = int(value)
	}

	return plan
}

func sanitizeDimensionScores(items []model.ReportDimensionScore) []model.ReportDimensionScore {
	out := make([]model.ReportDimensionScore, 0, len(items))
	for _, item := range items {
		item.DimensionCode = strings.TrimSpace(item.DimensionCode)
		item.Comment = strings.TrimSpace(item.Comment)
		item.Score = clampScore(item.Score)
		if item.DimensionCode == "" {
			continue
		}
		out = append(out, item)
	}

	return out
}
