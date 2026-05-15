package openai

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
		return nil, errors.New("openai api key is required")
	}

	baseURL := strings.TrimSpace(cfg.BaseURL)
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1/responses"
	}

	modelName := strings.TrimSpace(cfg.Model)
	if modelName == "" {
		modelName = "gpt-5.5"
	}

	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 60 * time.Second
	}

	return &Provider{
		apiKey:  cfg.APIKey,
		baseURL: baseURL,
		model:   modelName,
		client: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

func (p *Provider) Name() string {
	return "openai"
}

func (p *Provider) GenerateQuestions(ctx context.Context, input ai.GenerateQuestionsInput) ([]model.InterviewQuestion, error) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"questions": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"questionType": map[string]any{
							"type": "string",
							"enum": []string{"self_intro", "project", "system_design", "go_runtime", "behavior"},
						},
						"assessmentDimension": map[string]any{
							"type": "string",
							"enum": []string{"communication", "depth", "accuracy", "clarity"},
						},
						"prompt": map[string]any{
							"type": "string",
						},
						"expectedPoints": map[string]any{
							"type": "array",
							"items": map[string]any{
								"type": "string",
							},
						},
					},
					"required":             []string{"questionType", "assessmentDimension", "prompt", "expectedPoints"},
					"additionalProperties": false,
				},
			},
		},
		"required":             []string{"questions"},
		"additionalProperties": false,
	}

	systemPrompt := "你是一名中文技术面试官。请结合岗位信息、候选人简历、目标公司近几年常见题型和面试轮次，生成贴近真实场景的中文面试题。题目必须同时体现岗位匹配度、简历深挖和公司风格。expectedPoints 只返回简洁中文短语，不要英文，不要解释。"

	var payload struct {
		Questions []struct {
			QuestionType        string   `json:"questionType"`
			AssessmentDimension string   `json:"assessmentDimension"`
			Prompt              string   `json:"prompt"`
			ExpectedPoints      []string `json:"expectedPoints"`
		} `json:"questions"`
	}
	if err := p.requestStructuredJSON(ctx, systemPrompt, input, "interview_question_set", schema, &payload); err != nil {
		return nil, err
	}

	now := time.Now()
	questions := make([]model.InterviewQuestion, 0, len(payload.Questions))
	for index, item := range payload.Questions {
		questions = append(questions, model.InterviewQuestion{
			QuestionNo:          index + 1,
			QuestionType:        item.QuestionType,
			AssessmentDimension: item.AssessmentDimension,
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
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"overallScore":          map[string]any{"type": "number"},
			"accuracyScore":         map[string]any{"type": "number"},
			"clarityScore":          map[string]any{"type": "number"},
			"depthScore":            map[string]any{"type": "number"},
			"communicationScore":    map[string]any{"type": "number"},
			"strengths":             map[string]any{"type": "string"},
			"issues":                map[string]any{"type": "string"},
			"improvementSuggestion": map[string]any{"type": "string"},
			"followUpQuestion":      map[string]any{"type": "string"},
		},
		"required": []string{
			"overallScore",
			"accuracyScore",
			"clarityScore",
			"depthScore",
			"communicationScore",
			"strengths",
			"issues",
			"improvementSuggestion",
			"followUpQuestion",
		},
		"additionalProperties": false,
	}

	systemPrompt := "你是一名严谨但友好的中文技术面试官。请根据题目、考察维度、回答内容和上下文，输出结构化评分与点评。所有分数范围为 0 到 100，所有评价必须使用简体中文，结论要具体。"

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
	if err := p.requestStructuredJSON(ctx, systemPrompt, input, "answer_feedback", schema, &payload); err != nil {
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
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"overallScore":       map[string]any{"type": "number"},
			"performanceSummary": map[string]any{"type": "string"},
			"strengthsSummary":   map[string]any{"type": "string"},
			"weaknessSummary":    map[string]any{"type": "string"},
			"recommendedActions": map[string]any{"type": "string"},
			"nextPlan": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"focusAreas": map[string]any{
						"type":  "array",
						"items": map[string]any{"type": "string"},
					},
					"recommendedQuestionCount": map[string]any{"type": "number"},
				},
				"required":             []string{"focusAreas", "recommendedQuestionCount"},
				"additionalProperties": false,
			},
			"dimensionScores": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"dimensionCode": map[string]any{
							"type": "string",
							"enum": []string{"accuracy", "clarity", "depth", "communication"},
						},
						"score":   map[string]any{"type": "number"},
						"comment": map[string]any{"type": "string"},
					},
					"required":             []string{"dimensionCode", "score", "comment"},
					"additionalProperties": false,
				},
			},
		},
		"required": []string{
			"overallScore",
			"performanceSummary",
			"strengthsSummary",
			"weaknessSummary",
			"recommendedActions",
			"nextPlan",
			"dimensionScores",
		},
		"additionalProperties": false,
	}

	systemPrompt := "你是一名中文技术面试复盘教练。请基于整场问答历史、逐题评分和题目内容生成中文复盘报告。总结要具体、有诊断价值，训练建议要能直接执行，不要输出 Markdown。"

	var payload struct {
		OverallScore       float64                      `json:"overallScore"`
		PerformanceSummary string                       `json:"performanceSummary"`
		StrengthsSummary   string                       `json:"strengthsSummary"`
		WeaknessSummary    string                       `json:"weaknessSummary"`
		RecommendedActions string                       `json:"recommendedActions"`
		NextPlan           map[string]any               `json:"nextPlan"`
		DimensionScores    []model.ReportDimensionScore `json:"dimensionScores"`
	}
	if err := p.requestStructuredJSON(ctx, systemPrompt, input, "interview_report", schema, &payload); err != nil {
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

func (p *Provider) GenerateQuestionReviews(ctx context.Context, input ai.GenerateQuestionReviewsInput) ([]model.QuestionReviewItem, error) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"items": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"questionId": map[string]any{"type": "number"},
						"reviewType": map[string]any{
							"type": "string",
							"enum": []string{"correct_answer", "methodology"},
						},
						"title":   map[string]any{"type": "string"},
						"content": map[string]any{"type": "string"},
					},
					"required":             []string{"questionId", "reviewType", "title", "content"},
					"additionalProperties": false,
				},
			},
		},
		"required":             []string{"items"},
		"additionalProperties": false,
	}

	systemPrompt := "你是一名中文技术面试复盘教练。请根据每道题的题型、题目、考察点、用户回答和反馈，生成精简题目解析。知识点题返回 reviewType=correct_answer，并直接给出精简正确答案；主观题和开放题返回 reviewType=methodology，并给出回答方向或方法论。content 控制在 120 字以内。"

	var payload struct {
		Items []struct {
			QuestionID int64  `json:"questionId"`
			ReviewType string `json:"reviewType"`
			Title      string `json:"title"`
			Content    string `json:"content"`
		} `json:"items"`
	}
	if err := p.requestStructuredJSON(ctx, systemPrompt, input, "question_reviews", schema, &payload); err != nil {
		return nil, err
	}

	answerMap := make(map[int64]model.SessionAnswerItem, len(input.AnswerHistory))
	for _, item := range input.AnswerHistory {
		answerMap[item.QuestionID] = item
	}

	reviews := make([]model.QuestionReviewItem, 0, len(payload.Items))
	for _, item := range payload.Items {
		answerItem, ok := answerMap[item.QuestionID]
		if !ok {
			continue
		}

		reviews = append(reviews, model.QuestionReviewItem{
			QuestionID:          answerItem.QuestionID,
			QuestionNo:          answerItem.QuestionNo,
			QuestionType:        answerItem.QuestionType,
			AssessmentDimension: answerItem.AssessmentDimension,
			Prompt:              answerItem.Prompt,
			ExpectedPoints:      answerItem.ExpectedPoints,
			ReviewType:          strings.TrimSpace(item.ReviewType),
			Title:               strings.TrimSpace(item.Title),
			Content:             strings.TrimSpace(item.Content),
		})
	}

	return reviews, nil
}

func (p *Provider) requestStructuredJSON(ctx context.Context, systemPrompt string, input any, schemaName string, schema map[string]any, out any) error {
	inputJSON, err := json.MarshalIndent(input, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal provider input: %w", err)
	}

	requestBody := map[string]any{
		"model": p.model,
		"input": []map[string]any{
			{
				"role":    "system",
				"content": systemPrompt,
			},
			{
				"role":    "user",
				"content": "请根据下面的 JSON 上下文完成任务，并只返回符合 schema 的 JSON。\n\n" + string(inputJSON),
			},
		},
		"text": map[string]any{
			"format": map[string]any{
				"type":   "json_schema",
				"name":   schemaName,
				"strict": true,
				"schema": schema,
			},
		},
	}

	payload, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("marshal openai request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("create openai request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("request openai responses api: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read openai response: %w", err)
	}

	if resp.StatusCode >= 300 {
		return fmt.Errorf("openai responses api returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	outputText, err := extractOutputText(body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal([]byte(outputText), out); err != nil {
		return fmt.Errorf("decode structured output: %w; raw output: %s", err, outputText)
	}

	return nil
}

func extractOutputText(body []byte) (string, error) {
	var response struct {
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
		Output []struct {
			Type    string `json:"type"`
			Content []struct {
				Type    string `json:"type"`
				Text    string `json:"text"`
				Refusal string `json:"refusal"`
			} `json:"content"`
		} `json:"output"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("decode openai response envelope: %w", err)
	}

	if response.Error != nil && strings.TrimSpace(response.Error.Message) != "" {
		return "", errors.New(response.Error.Message)
	}

	var builder strings.Builder
	for _, item := range response.Output {
		if item.Type != "message" {
			continue
		}

		for _, content := range item.Content {
			switch content.Type {
			case "output_text":
				builder.WriteString(content.Text)
			case "refusal":
				return "", fmt.Errorf("model refusal: %s", strings.TrimSpace(content.Refusal))
			}
		}
	}

	text := strings.TrimSpace(builder.String())
	if text == "" {
		return "", fmt.Errorf("openai response did not contain output_text: %s", strings.TrimSpace(string(body)))
	}

	return text, nil
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
