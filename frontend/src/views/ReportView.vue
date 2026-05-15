<template>
  <div class="page-grid report-page">
    <section v-if="state.error" class="message-banner message-error">
      {{ state.error }}
    </section>

    <section v-if="state.loading.report" class="message-banner">正在加载复盘报告...</section>

    <template v-else-if="state.report">
      <section class="stats-grid report-stats-grid report-score-grid">
        <article class="metric-card report-score-card report-score-overall" :class="scoreClass(state.report.overallScore)">
          <div class="report-score-head">
            <p class="metric-label">综合得分</p>
            <span class="report-score-badge">{{ scoreToneText(state.report.overallScore) }}</span>
          </div>
          <div class="report-score-main">
            <p class="metric-value">{{ Math.round(state.report.overallScore) }}</p>
            <span class="report-score-face">{{ scoreEmoji(state.report.overallScore) }}</span>
          </div>
        </article>

        <article
          v-for="item in state.report.dimensionScores"
          :key="item.dimensionCode"
          class="metric-card report-score-card"
          :class="scoreClass(item.score)"
        >
          <div class="report-score-head">
            <p class="metric-label">{{ toDimensionLabel(item.dimensionCode) }}</p>
            <span class="report-score-dot"></span>
          </div>
          <div class="report-score-main">
            <p class="metric-value">{{ Math.round(item.score) }}</p>
          </div>
          <p class="metric-hint report-score-comment">{{ item.comment }}</p>
        </article>
      </section>

      <section class="panel">
        <div class="panel-header">
          <h3>综合总结</h3>
        </div>
        <div class="report-copy">
          <p>{{ state.report.performanceSummary }}</p>
          <p><strong>亮点：</strong>{{ state.report.strengthsSummary }}</p>
          <p><strong>薄弱项：</strong>{{ state.report.weaknessSummary }}</p>
          <p><strong>下一步建议：</strong>{{ state.report.recommendedActions }}</p>
        </div>
      </section>

      <section class="panel report-question-panel">
        <div class="panel-header">
          <div>
            <h3>题目回看</h3>
          </div>
          <span class="status-chip status-pending">{{ answerHistory.length }} 题</span>
        </div>

        <div v-if="state.loading.questionReviews" class="message-banner">正在生成题目解析...</div>

        <div v-if="answerHistory.length" class="question-review-list">
          <article
            v-for="item in answerHistory"
            :key="`${item.questionId}-${item.answer.submittedAt}`"
            class="question-review-card"
          >
            <button
              type="button"
              class="question-review-toggle"
              :class="{ 'is-open': isQuestionOpen(item.questionId) }"
              @click="toggleQuestion(item.questionId)"
            >
              <div class="question-review-summary">
                <span class="question-review-order">Q{{ item.questionNo }}</span>
                <div class="question-review-copy">
                  <strong>{{ item.prompt }}</strong>
                  <span>
                    {{ toQuestionTypeLabel(item.questionType) }} · {{ toDimensionLabel(item.assessmentDimension) }}
                  </span>
                </div>
              </div>
              <div class="question-review-actions">
                <span class="question-review-score" :class="scoreClass(item.feedback?.overallScore || 0)">
                  {{ item.feedback ? Math.round(item.feedback.overallScore) : "--" }}
                </span>
                <span class="question-review-arrow">{{ isQuestionOpen(item.questionId) ? "收起" : "展开" }}</span>
              </div>
            </button>

            <div v-if="isQuestionOpen(item.questionId)" class="question-review-body">
              <div class="question-review-block">
                <p class="question-review-title">面试官想听什么</p>
                <p>{{ summarizeExpectedPoints(item.expectedPoints) }}</p>
              </div>

              <div class="question-review-block question-review-guide">
                <p class="question-review-title">{{ reviewMeta(item.questionId).title }}</p>
                <p>{{ reviewMeta(item.questionId).content }}</p>
              </div>

              <div class="question-review-inline-grid">
                <div class="question-review-block">
                  <p class="question-review-title">你的回答</p>
                  <p>{{ item.answer.answerText }}</p>
                </div>
                <div class="question-review-block">
                  <p class="question-review-title">本题建议</p>
                  <p>{{ compactSuggestion(item) }}</p>
                </div>
              </div>
            </div>
          </article>
        </div>

        <div v-else class="empty-state compact-empty">
          <p>当前还没有可展示的题目记录。完成答题后，这里会显示每道题的回看内容。</p>
        </div>
      </section>

      <section class="report-grid">
        <article class="panel">
          <div class="panel-header">
            <h3>维度图表</h3>
          </div>
          <div class="chart-list">
            <div v-for="item in state.report.dimensionScores" :key="item.dimensionCode" class="chart-row">
              <div class="chart-row-head">
                <span>{{ toDimensionLabel(item.dimensionCode) }}</span>
                <strong>{{ item.score.toFixed(2) }}</strong>
              </div>
              <div class="chart-bar">
                <div
                  class="chart-bar-fill"
                  :class="scoreClass(item.score)"
                  :style="{ width: `${Math.min(item.score, 100)}%` }"
                ></div>
              </div>
              <p class="chart-comment">{{ item.comment }}</p>
            </div>
          </div>
        </article>

        <article class="panel">
          <div class="panel-header">
            <h3>下一轮训练计划</h3>
          </div>
          <div class="tag-list">
            <span v-for="item in state.report.nextPlan?.focusAreas || []" :key="item" class="tag">
              {{ focusAreaText(item) }}
            </span>
          </div>
          <p class="report-copy single-line-copy">
            推荐题量：{{ state.report.nextPlan?.recommendedQuestionCount || "--" }}
          </p>
        </article>
      </section>
    </template>

    <section v-else class="panel">
      <div class="panel-header">
        <h3>暂无报告</h3>
      </div>
      <p class="lead">先完成一场面试，这里才会显示对应的复盘报告。</p>
    </section>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { toDimensionLabel, toQuestionTypeLabel } from "../data/mock";
import { useInterviewState } from "../state/interview";

const route = useRoute();
const router = useRouter();
const {
  state,
  selectReportSession,
  loadReport,
  loadSessionAnswers,
  loadQuestionReviews,
  getAnswerHistory,
  getQuestionReviews
} = useInterviewState();

const openQuestionIds = ref([]);
const reportSessionId = computed(() => Number(route.query.sessionId || state.selectedReportSessionId || 0));

const answerHistory = computed(() => getAnswerHistory(reportSessionId.value));
const questionReviews = computed(() => getQuestionReviews(reportSessionId.value));
const questionReviewMap = computed(() => {
  const map = new Map();
  for (const item of questionReviews.value) {
    map.set(item.questionId, item);
  }
  return map;
});

function focusAreaText(value) {
  const map = {
    high_concurrency: "高并发设计",
    cache_consistency: "缓存一致性",
    project_depth: "项目深度表达"
  };

  return map[value] || value;
}

function scoreClass(score) {
  if (score < 60) {
    return "is-bad";
  }
  if (score <= 80) {
    return "is-warn";
  }
  return "is-good";
}

function scoreEmoji(score) {
  if (score < 60) {
    return "😭";
  }
  if (score <= 80) {
    return "😐";
  }
  return "😊";
}

function scoreToneText(score) {
  if (score < 60) {
    return "需预警";
  }
  if (score <= 80) {
    return "可提升";
  }
  return "状态好";
}

function isQuestionOpen(questionId) {
  return openQuestionIds.value.includes(questionId);
}

function toggleQuestion(questionId) {
  if (isQuestionOpen(questionId)) {
    openQuestionIds.value = openQuestionIds.value.filter((item) => item !== questionId);
    return;
  }

  openQuestionIds.value = [...openQuestionIds.value, questionId];
}

function summarizeExpectedPoints(points) {
  if (!points?.length) {
    return "建议先给结论，再补关键原因和实际场景。";
  }

  return `建议优先覆盖：${points.slice(0, 3).join("、")}。`;
}

function reviewMeta(questionId) {
  const item = questionReviewMap.value.get(questionId);
  if (!item) {
    return {
      title: "题目解析",
      content: "当前题目的 AI 解析暂未生成，请稍后刷新重试。"
    };
  }

  return {
    title: item.title || (item.reviewType === "correct_answer" ? "正确答案" : "回答方向"),
    content: item.content || "当前题目的 AI 解析暂未生成，请稍后刷新重试。"
  };
}

function compactSuggestion(item) {
  if (item.feedback?.improvementSuggestion) {
    return item.feedback.improvementSuggestion;
  }

  return "回答可以再聚焦一点，优先说结论、关键步骤和最终结果。";
}

onMounted(async () => {
  const sessionId = reportSessionId.value;
  if (!sessionId) {
    window.alert("请先选择复盘记录");
    router.replace("/history");
    return;
  }

  selectReportSession(sessionId);
  await Promise.all([loadReport(sessionId), loadSessionAnswers(sessionId), loadQuestionReviews(sessionId)]);
});
</script>
