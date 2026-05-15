<template>
  <div class="page-grid">
    <section class="panel">
      <div class="panel-header">
        <div>
          <h3>面试进度</h3>
          <p>当前会话会实时同步后端状态，答题历史会从后端 MySQL 实时读取。</p>
        </div>
        <span class="pill">{{ progressLabel }}</span>
      </div>

      <div class="progress-track">
        <div class="progress-fill" :style="{ width: `${progressPercent}%` }"></div>
      </div>

      <div class="summary-list summary-grid">
        <div class="summary-row">
          <span>会话名称</span>
          <strong>{{ state.sessionDetail?.session?.sessionName || "--" }}</strong>
        </div>
        <div class="summary-row">
          <span>当前状态</span>
          <strong>{{ statusLabel }}</strong>
        </div>
        <div class="summary-row">
          <span>已拉题数</span>
          <strong>{{ state.sessionDetail?.session?.currentQuestionNo || 0 }}</strong>
        </div>
        <div class="summary-row">
          <span>题目总数</span>
          <strong>{{ state.sessionDetail?.session?.questionCount || 0 }}</strong>
        </div>
      </div>
    </section>

    <div class="interview-layout">
      <section class="panel question-list-panel">
        <div class="panel-header">
          <h3>题目列表</h3>
        </div>

        <div v-if="questionItems.length" class="question-list">
          <button
            v-for="item in questionItems"
            :key="item.id"
            class="question-list-item"
            :class="{
              'is-current': state.currentQuestion?.id === item.id,
              'is-answered': isAnswered(item.id)
            }"
            @click="focusQuestion(item)"
          >
            <span class="question-order">Q{{ item.questionNo }}</span>
            <div class="question-list-copy">
              <strong>{{ toQuestionTypeLabel(item.questionType) }}</strong>
              <span>{{ toDimensionLabel(item.assessmentDimension) }}</span>
            </div>
          </button>
        </div>

        <div v-else class="empty-state compact-empty">
          <p>创建会话后，这里会显示本场面试题目。</p>
        </div>
      </section>

      <section class="panel question-panel">
        <div class="panel-header">
          <h3>当前题目</h3>
          <span class="status-chip status-pending">{{ sessionModeLabel }}</span>
        </div>

        <div v-if="state.error" class="message-banner message-error">
          {{ state.error }}
        </div>

        <div v-if="state.loading.question" class="message-banner">正在加载下一题...</div>

        <template v-else-if="state.currentQuestion">
          <div class="question-meta">
            <span>{{ toQuestionTypeLabel(state.currentQuestion.questionType) }}</span>
            <span>{{ toDimensionLabel(state.currentQuestion.assessmentDimension) }}</span>
          </div>

          <p class="question-text">{{ state.currentQuestion.prompt }}</p>

          <div class="timeline-box">
            <p class="timeline-title">回答要点</p>
            <div class="tag-list">
              <span v-for="point in state.currentQuestion.expectedPoints" :key="point" class="tag">
                {{ point }}
              </span>
            </div>
          </div>
        </template>

        <div v-else class="empty-state">
          <p>当前还没有可展示的题目。</p>
          <button class="ghost-button" @click="loadQuestion">加载题目</button>
        </div>
      </section>

      <section class="panel answer-panel">
        <div class="panel-header">
          <h3>答题区</h3>
          <span class="eyebrow">{{ sessionModeLabel }}</span>
        </div>

        <textarea
          v-model="answerText"
          rows="14"
          :disabled="!state.currentQuestion || state.loading.answer"
          placeholder="输入你的回答，提交后会拿到即时反馈并进入下一题。"
        />

        <div class="button-row">
          <button class="ghost-button" :disabled="!answerText" @click="resetDraft">清空草稿</button>
          <button
            class="primary-button"
            :disabled="!state.currentQuestion || !answerText.trim() || state.loading.answer"
            @click="handleSubmit"
          >
            {{ state.loading.answer ? "提交中..." : "提交回答" }}
          </button>
        </div>

        <div v-if="state.latestFeedback" class="feedback-card">
          <div class="panel-header">
            <h3>最新反馈</h3>
            <span class="status-chip status-success">{{ Math.round(state.latestFeedback.overallScore) }}</span>
          </div>
          <p><strong>亮点：</strong>{{ state.latestFeedback.strengths }}</p>
          <p><strong>问题：</strong>{{ state.latestFeedback.issues }}</p>
          <p><strong>建议：</strong>{{ state.latestFeedback.improvementSuggestion }}</p>
          <p><strong>追问：</strong>{{ state.latestFeedback.followUpQuestion }}</p>
        </div>
      </section>

      <section class="panel rubric-panel">
        <div class="panel-header">
          <h3>评分维度</h3>
        </div>

        <div class="score-list">
          <div v-for="item in scoreWeights" :key="item.label" class="score-row">
            <span>{{ item.label }}</span>
            <strong>{{ item.value }}</strong>
          </div>
        </div>
      </section>
    </div>

    <section class="panel">
      <div class="panel-header">
        <h3>答案历史</h3>
        <p>这里展示的是当前会话已经落库的回答历史和反馈结果。</p>
      </div>

      <div v-if="answerHistory.length" class="answer-history-list">
        <article
          v-for="item in answerHistory"
          :key="`${item.questionId}-${item.submittedAt}`"
          class="answer-history-card"
        >
          <div class="panel-header">
            <div>
              <h3>Q{{ item.questionNo }} · {{ toQuestionTypeLabel(item.questionType) }}</h3>
              <p>{{ toDimensionLabel(item.assessmentDimension) }}</p>
            </div>
            <span class="status-chip status-success">{{ Math.round(item.feedback.overallScore) }}</span>
          </div>
          <p class="history-question">{{ item.prompt }}</p>
          <div class="history-answer-block">
            <p class="history-answer-title">你的回答</p>
            <p>{{ item.answerText }}</p>
          </div>
          <div class="history-feedback-grid">
            <div>
              <strong>亮点</strong>
              <p>{{ item.feedback.strengths }}</p>
            </div>
            <div>
              <strong>建议</strong>
              <p>{{ item.feedback.improvementSuggestion }}</p>
            </div>
          </div>
        </article>
      </div>

      <div v-else class="empty-state compact-empty">
        <p>提交回答后，这里会展示本场面试已写入 MySQL 的答题历史。</p>
      </div>
    </section>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import {
  scoreWeights,
  toDimensionLabel,
  toQuestionTypeLabel,
  toStatusLabel
} from "../data/mock";
import { useInterviewState } from "../state/interview";

const route = useRoute();
const router = useRouter();
const {
  state,
  selectSession,
  ensureQuestion,
  submitAnswer,
  loadSession,
  loadSessionAnswers,
  loadReport,
  getAnswerHistory,
  setError
} = useInterviewState();

const answerText = ref("");

const progressLabel = computed(() => {
  const current = state.sessionDetail?.session?.currentQuestionNo || 0;
  const total = state.sessionDetail?.session?.questionCount || 0;
  return `${current}/${total}`;
});

const progressPercent = computed(() => {
  const current = state.sessionDetail?.session?.currentQuestionNo || 0;
  const total = state.sessionDetail?.session?.questionCount || 0;
  if (!total) {
    return 0;
  }

  return Math.min(100, Math.round((current / total) * 100));
});

const sessionModeLabel = computed(() => {
  if (state.sessionDetail?.session?.interviewMode === "text") {
    return "文字模式";
  }

  return state.sessionDetail?.session?.interviewMode || "--";
});

const statusLabel = computed(() => toStatusLabel(state.sessionDetail?.session?.status));
const questionItems = computed(() => state.sessionDetail?.questions || []);
const answerHistory = computed(() => getAnswerHistory());

function isAnswered(questionId) {
  return answerHistory.value.some((item) => item.questionId === questionId);
}

function focusQuestion(item) {
  if (state.currentQuestion?.id === item.id || isAnswered(item.id)) {
    return;
  }

  state.currentQuestion = item;
}

async function syncSessionFromRoute() {
  const sessionId = Number(route.query.sessionId || state.sessionId);
  if (!sessionId) {
    window.alert("请先进行面试配置");
    setError("请先进行面试配置");
    router.replace("/setup");
    return;
  }

  selectSession(sessionId);
  await loadSession(sessionId);
  await loadSessionAnswers(sessionId);

  if (state.sessionDetail?.session?.status === "completed") {
    await loadReport(sessionId);
    router.replace({
      path: "/report",
      query: { sessionId: String(sessionId) }
    });
    return;
  }

  await ensureQuestion(sessionId);
}

async function loadQuestion() {
  await syncSessionFromRoute();
}

function resetDraft() {
  answerText.value = "";
}

async function handleSubmit() {
  try {
    const result = await submitAnswer(answerText.value);
    answerText.value = "";

    if (result.completed) {
      router.push({
        path: "/report",
        query: { sessionId: String(state.sessionId) }
      });
    }
  } catch {
    // store already captured the error
  }
}

onMounted(() => {
  syncSessionFromRoute();
});
</script>
