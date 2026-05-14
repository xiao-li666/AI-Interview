import { reactive } from "vue";
import { api } from "../lib/api";
import { setupDefaults } from "../data/mock";

const STORAGE_KEY = "ai-interview-app-state";

function loadPersistedState() {
  try {
    const raw = window.localStorage.getItem(STORAGE_KEY);
    if (!raw) {
      return null;
    }

    return JSON.parse(raw);
  } catch {
    return null;
  }
}

const persisted = typeof window !== "undefined" ? loadPersistedState() : null;

const state = reactive({
  config: {
    ...setupDefaults,
    ...(persisted?.config || {})
  },
  planId: persisted?.planId || null,
  sessionId: persisted?.sessionId || null,
  sessionDetail: null,
  currentQuestion: persisted?.currentQuestion || null,
  latestFeedback: null,
  latestAnswer: null,
  report: null,
  aiConfig: null,
  answerHistoryBySession: {},
  history: [],
  health: null,
  loading: {
    setup: false,
    question: false,
    answer: false,
    report: false,
    history: false,
    home: false,
    aiConfig: false,
    aiConfigTest: false
  },
  error: ""
});

function persist() {
  if (typeof window === "undefined") {
    return;
  }

  window.localStorage.setItem(
    STORAGE_KEY,
    JSON.stringify({
      config: state.config,
      planId: state.planId,
      sessionId: state.sessionId,
      currentQuestion: state.currentQuestion
    })
  );
}

function setError(message) {
  state.error = message || "";
}

function clearRuntimeState() {
  state.sessionDetail = null;
  state.currentQuestion = null;
  state.latestFeedback = null;
  state.latestAnswer = null;
  state.report = null;
}

function getAnswerHistory(sessionId = state.sessionId) {
  if (!sessionId) {
    return [];
  }

  return state.answerHistoryBySession[String(sessionId)] || [];
}

async function bootstrapHome() {
  state.loading.home = true;
  setError("");

  try {
    state.health = await api.health();
    state.history = await api.getHistory(state.config.userId);
  } catch (error) {
    setError(error.message);
  } finally {
    state.loading.home = false;
  }
}

async function createInterviewFlow() {
  state.loading.setup = true;
  setError("");

  try {
    const plan = await api.createPlan({
      userId: Number(state.config.userId),
      jobTitle: state.config.jobTitle,
      jobCategory: state.config.jobCategory,
      levelCode: state.config.levelCode,
      interviewType: state.config.interviewType,
      interviewMode: state.config.interviewMode,
      difficultyLevel: state.config.difficultyLevel,
      questionCount: Number(state.config.questionCount),
      interviewerStyle: state.config.interviewerStyle
    });

    const sessionDetail = await api.createSession({
      userId: Number(state.config.userId),
      jobTargetId: 0,
      jobTitle: state.config.jobTitle,
      sessionName: state.config.sessionName,
      roundType: state.config.roundType,
      interviewMode: state.config.interviewMode,
      difficultyLevel: state.config.difficultyLevel,
      questionCount: Number(state.config.questionCount),
      interviewerStyle: state.config.interviewerStyle
    });

    state.planId = plan.id;
    state.sessionId = sessionDetail.session.id;
    state.sessionDetail = sessionDetail;
    state.currentQuestion = null;
    state.latestFeedback = null;
    state.latestAnswer = null;
    state.report = null;
    state.answerHistoryBySession[String(sessionDetail.session.id)] = [];
    persist();

    return sessionDetail;
  } catch (error) {
    setError(error.message);
    throw error;
  } finally {
    state.loading.setup = false;
  }
}

async function loadSession(sessionId = state.sessionId) {
  if (!sessionId) {
    return null;
  }

  try {
    const detail = await api.getSession(sessionId);
    state.sessionId = detail.session.id;
    state.sessionDetail = detail;

    if (state.currentQuestion?.sessionId !== detail.session.id || detail.session.status === "completed") {
      state.currentQuestion = null;
    }

    persist();
    return detail;
  } catch (error) {
    setError(error.message);
    throw error;
  }
}

async function loadNextQuestion(sessionId = state.sessionId) {
  if (!sessionId) {
    return null;
  }

  state.loading.question = true;
  setError("");

  try {
    const question = await api.getNextQuestion(sessionId);
    state.currentQuestion = question;
    await loadSession(sessionId);
    persist();
    return question;
  } catch (error) {
    setError(error.message);
    throw error;
  } finally {
    state.loading.question = false;
  }
}

async function ensureQuestion(sessionId = state.sessionId) {
  const detail = await loadSession(sessionId);
  if (!detail) {
    return null;
  }

  if (detail.session.status === "completed") {
    return null;
  }

  if (state.currentQuestion?.sessionId === detail.session.id) {
    return state.currentQuestion;
  }

  return loadNextQuestion(sessionId);
}

async function loadSessionAnswers(sessionId = state.sessionId) {
  if (!sessionId) {
    return [];
  }

  try {
    const items = await api.getSessionAnswers(sessionId);
    state.answerHistoryBySession[String(sessionId)] = items;
    return items;
  } catch (error) {
    setError(error.message);
    throw error;
  }
}

async function submitAnswer(answerText) {
  if (!state.sessionId || !state.currentQuestion) {
    throw new Error("当前没有可提交的题目");
  }

  state.loading.answer = true;
  setError("");

  try {
    const result = await api.submitAnswer({
      sessionId: Number(state.sessionId),
      questionId: Number(state.currentQuestion.id),
      answerText,
      answerDurationSeconds: 60
    });

    state.latestAnswer = result.answer;
    state.latestFeedback = result.feedback;
    await loadSession(state.sessionId);
    await loadSessionAnswers(state.sessionId);

    if (state.sessionDetail?.session?.status === "completed") {
      await loadReport(state.sessionId);
      state.currentQuestion = null;
      persist();
      return { completed: true };
    }

    await loadNextQuestion(state.sessionId);
    return { completed: false };
  } catch (error) {
    setError(error.message);
    throw error;
  } finally {
    state.loading.answer = false;
  }
}

async function loadReport(sessionId = state.sessionId) {
  if (!sessionId) {
    return null;
  }

  state.loading.report = true;
  setError("");

  try {
    const report = await api.getReport(sessionId);
    state.report = report;
    return report;
  } catch (error) {
    setError(error.message);
    throw error;
  } finally {
    state.loading.report = false;
  }
}

async function loadHistory() {
  state.loading.history = true;
  setError("");

  try {
    const items = await api.getHistory(state.config.userId);
    state.history = items;
    return items;
  } catch (error) {
    setError(error.message);
    throw error;
  } finally {
    state.loading.history = false;
  }
}

async function loadAIConfig(userId = state.config.userId) {
  state.loading.aiConfig = true;
  setError("");

  try {
    const config = await api.getAIConfig(Number(userId));
    state.aiConfig = config;
    return config;
  } catch (error) {
    if (error.message === "resource not found") {
      state.aiConfig = null;
      return null;
    }
    setError(error.message);
    throw error;
  } finally {
    state.loading.aiConfig = false;
  }
}

async function saveAIConfig(payload) {
  state.loading.aiConfig = true;
  setError("");

  try {
    const config = await api.saveAIConfig(payload);
    state.aiConfig = config;
    return config;
  } catch (error) {
    setError(error.message);
    throw error;
  } finally {
    state.loading.aiConfig = false;
  }
}

async function testAIConfig(payload) {
  state.loading.aiConfigTest = true;
  setError("");

  try {
    return await api.testAIConfig(payload);
  } catch (error) {
    setError(error.message);
    throw error;
  } finally {
    state.loading.aiConfigTest = false;
  }
}

function updateConfig(patch) {
  state.config = {
    ...state.config,
    ...patch
  };
  persist();
}

function selectSession(sessionId) {
  if (state.sessionId !== sessionId) {
    state.currentQuestion = null;
  }

  state.sessionId = sessionId;
  persist();
}

function resetInterviewState() {
  state.planId = null;
  state.sessionId = null;
  clearRuntimeState();
  persist();
}

export function useInterviewState() {
  return {
    state,
    setError,
    updateConfig,
    bootstrapHome,
    createInterviewFlow,
    loadSession,
    loadNextQuestion,
    ensureQuestion,
    loadSessionAnswers,
    submitAnswer,
    loadReport,
    loadHistory,
    loadAIConfig,
    saveAIConfig,
    testAIConfig,
    getAnswerHistory,
    selectSession,
    resetInterviewState
  };
}
