package mock

import (
	"context"
	"fmt"
	"strings"
	"time"

	"ai-interview/backend/internal/ai"
	"ai-interview/backend/internal/model"
)

type Provider struct{}

func NewProvider() *Provider {
	return &Provider{}
}

func (p *Provider) Name() string {
	return "mock"
}

func (p *Provider) GenerateQuestions(_ context.Context, input ai.GenerateQuestionsInput) ([]model.InterviewQuestion, error) {
	templates := []struct {
		questionType string
		dimension    string
		prompt       string
		points       []string
	}{
		{
			questionType: "self_intro",
			dimension:    "communication",
			prompt:       "请先做一个两分钟以内的自我介绍，并重点说明你最近一段项目经历里最能代表你的技术决策。",
			points:       []string{"背景概述", "核心项目", "技术价值"},
		},
		{
			questionType: "project",
			dimension:    "depth",
			prompt:       "你最近做过的 Go 项目里，最复杂的一次故障排查是什么？你如何定位、修复并复盘？",
			points:       []string{"问题背景", "定位路径", "修复动作", "复盘结果"},
		},
		{
			questionType: "system_design",
			dimension:    "accuracy",
			prompt:       "如果要设计一个高并发抢购系统，你会怎么处理库存扣减、削峰、幂等和最终一致性？",
			points:       []string{"库存模型", "限流削峰", "幂等设计", "一致性方案"},
		},
		{
			questionType: "go_runtime",
			dimension:    "accuracy",
			prompt:       "结合实际问题说明一个 goroutine 泄漏通常怎么出现，线上如何观测和处理？",
			points:       []string{"成因", "监控指标", "排查工具", "修复策略"},
		},
		{
			questionType: "behavior",
			dimension:    "clarity",
			prompt:       "如果你发现需求变更会显著影响交付周期，但产品和业务方很着急，你会怎么沟通？",
			points:       []string{"风险表达", "方案权衡", "对齐机制"},
		},
	}

	total := input.QuestionCount
	if total <= 0 {
		total = 3
	}
	if total > len(templates) {
		total = len(templates)
	}

	now := time.Now()
	questions := make([]model.InterviewQuestion, 0, total)
	for index := 0; index < total; index++ {
		item := templates[index]
		questions = append(questions, model.InterviewQuestion{
			QuestionNo:          index + 1,
			QuestionType:        item.questionType,
			AssessmentDimension: item.dimension,
			Prompt:              item.prompt,
			ExpectedPoints:      item.points,
			Source:              p.Name(),
			IsFollowUp:          false,
			CreatedAt:           now,
		})
	}

	return questions, nil
}

func (p *Provider) EvaluateAnswer(_ context.Context, input ai.EvaluateAnswerInput) (model.AnswerFeedback, error) {
	lengthScore := 74.0
	if len([]rune(input.AnswerText)) > 120 {
		lengthScore = 84
	}

	return model.AnswerFeedback{
		OverallScore:          lengthScore,
		AccuracyScore:         lengthScore - 2,
		ClarityScore:          lengthScore + 3,
		DepthScore:            lengthScore - 4,
		CommunicationScore:    lengthScore + 2,
		Strengths:             "回答结构比较清楚，能够覆盖主要问题链路。",
		Issues:                "关键实现细节还不够具体，缺少量化结果和线上案例。",
		ImprovementSuggestion: "建议采用“背景-方案-权衡-结果”的表达模板，并补足具体技术细节。",
		FollowUpQuestion:      fmt.Sprintf("如果让你继续展开，请补充你在“%s”里的关键设计取舍。", truncate(input.AnswerText, 20)),
		CreatedAt:             time.Now(),
		Source:                p.Name(),
	}, nil
}

func (p *Provider) GenerateReport(_ context.Context, input ai.GenerateReportInput) (model.InterviewReport, error) {
	overallScore, accuracyScore, clarityScore, depthScore, communicationScore := aggregateScores(input.AnswerHistory)

	return model.InterviewReport{
		SessionID:          input.Session.ID,
		OverallScore:       overallScore,
		PerformanceSummary: "整体回答覆盖面比较完整，主问题链路清楚，但实现细节还可以再具体一些。",
		StrengthsSummary:   "回答结构比较稳定，能够清晰覆盖核心问题链路。",
		WeaknessSummary:    "在缓存一致性、幂等控制和故障恢复等关键细节上，深度还不够充分。",
		RecommendedActions: "下一轮建议重点训练高并发设计、缓存一致性、重试补偿，以及项目表达的完整度。",
		NextPlan: map[string]any{
			"focusAreas":               []string{"高并发设计", "缓存一致性", "项目深度表达"},
			"recommendedQuestionCount": 6,
		},
		DimensionScores: []model.ReportDimensionScore{
			{DimensionCode: "accuracy", Score: accuracyScore, Comment: "技术准确性整体稳定。"},
			{DimensionCode: "clarity", Score: clarityScore, Comment: "回答结构清晰，表达比较顺畅。"},
			{DimensionCode: "depth", Score: depthScore, Comment: "技术深度还有提升空间。"},
			{DimensionCode: "communication", Score: communicationScore, Comment: "沟通表达自然，节奏较好。"},
		},
		CreatedAt: time.Now(),
		Source:    p.Name(),
	}, nil
}

func (p *Provider) GenerateQuestionReviews(_ context.Context, input ai.GenerateQuestionReviewsInput) ([]model.QuestionReviewItem, error) {
	items := make([]model.QuestionReviewItem, 0, len(input.AnswerHistory))
	for _, item := range input.AnswerHistory {
		reviewType := "methodology"
		title := "回答方向"
		content := buildMockMethodology(item)
		if isKnowledgeQuestionType(item.QuestionType) {
			reviewType = "correct_answer"
			title = "正确答案"
			content = buildMockKnowledgeAnswer(item)
		}

		items = append(items, model.QuestionReviewItem{
			QuestionID:          item.QuestionID,
			QuestionNo:          item.QuestionNo,
			QuestionType:        item.QuestionType,
			AssessmentDimension: item.AssessmentDimension,
			Prompt:              item.Prompt,
			ExpectedPoints:      item.ExpectedPoints,
			ReviewType:          reviewType,
			Title:               title,
			Content:             content,
		})
	}

	return items, nil
}

func aggregateScores(items []model.SessionAnswerItem) (float64, float64, float64, float64, float64) {
	if len(items) == 0 {
		return 0, 0, 0, 0, 0
	}

	var overall float64
	var accuracy float64
	var clarity float64
	var depth float64
	var communication float64
	var count float64

	for _, item := range items {
		if item.Feedback == nil {
			continue
		}

		count++
		overall += item.Feedback.OverallScore
		accuracy += item.Feedback.AccuracyScore
		clarity += item.Feedback.ClarityScore
		depth += item.Feedback.DepthScore
		communication += item.Feedback.CommunicationScore
	}

	if count == 0 {
		return 0, 0, 0, 0, 0
	}

	return overall / count, accuracy / count, clarity / count, depth / count, communication / count
}

func truncate(text string, size int) string {
	runes := []rune(text)
	if len(runes) <= size {
		return text
	}

	return string(runes[:size]) + "..."
}

func isKnowledgeQuestionType(questionType string) bool {
	return questionType == "go_runtime" || questionType == "system_design"
}

func buildMockKnowledgeAnswer(item model.SessionAnswerItem) string {
	if strings.Contains(item.Prompt, "切片") {
		return "Go 切片是对底层数组的描述符，包含指针、长度和容量。扩容时小容量通常倍增，容量变大后增长比例会下降；扩容可能触发底层数组拷贝。"
	}

	if strings.Contains(item.Prompt, "goroutine") || strings.Contains(item.Prompt, "泄漏") {
		return "goroutine 泄漏常见于阻塞发送、阻塞接收、未退出循环或 context 未取消。排查时常看 pprof、goroutine 数量和阻塞栈，再用超时、取消和资源关闭治理。"
	}

	if strings.Contains(item.Prompt, "高并发") || strings.Contains(item.Prompt, "库存") {
		return "高并发题通常先稳流量，再保数据正确。入口限流削峰，中间异步化平滑请求，库存扣减用原子操作或条件更新，再用幂等、补偿和最终一致性兜底。"
	}

	if len(item.ExpectedPoints) > 0 {
		return "这题的核心答案应覆盖：" + strings.Join(limitStrings(item.ExpectedPoints, 3), "、") + "。"
	}

	return "这类知识点题建议先给结论，再补关键原理、场景和边界条件。"
}

func buildMockMethodology(item model.SessionAnswerItem) string {
	switch item.QuestionType {
	case "behavior":
		return "这类题建议按 STAR 来答：背景、任务、行动、结果，每段一到两句即可。"
	case "project":
		return "建议按“项目背景-你的职责-关键决策-结果复盘”来讲，重点突出你的判断和取舍。"
	default:
		return "这类开放题建议先给结论，再拆 2 到 3 个关键点展开，最后补结果或复盘。"
	}
}

func limitStrings(items []string, limit int) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		out = append(out, item)
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out
}
