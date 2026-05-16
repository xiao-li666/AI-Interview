import { reactive } from "vue";
import { api, getAuthToken, setAuthToken } from "../lib/api";
import { companyOptions, setupDefaults } from "../data/mock";

const STORAGE_KEY = "ai-interview-app-state";

function getInitialOpenTabs() {
  if (typeof window === "undefined") {
    return ["/"];
  }

  const path = window.location.pathname || "/";
  if (path === "/login" || path === "/register") {
    return [];
  }
  return [path];
}

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
const persistedConfig = persisted?.config
  ? {
      ...persisted.config,
      resumeText: "",
      resumeFileName: ""
    }
  : null;

const state = reactive({
  user: persisted?.user || null,
  isAuthenticated: Boolean(getAuthToken()),
  authReady: false,
  config: {
    ...setupDefaults,
    ...(persistedConfig || {})
  },
  companyOptions: normalizeCompanyOptions(persisted?.companyOptions),
  planId: persisted?.planId || null,
  sessionId: persisted?.sessionId || null,
  selectedReportSessionId: persisted?.selectedReportSessionId || null,
  sessionDetail: null,
  currentQuestion: persisted?.currentQuestion || null,
  latestFeedback: null,
  latestAnswer: null,
  report: null,
  aiConfig: null,
  openTabs: getInitialOpenTabs(),
  answerHistoryBySession: {},
  questionReviewsBySession: {},
  history: [],
  health: null,
  loading: {
    auth: false,
    register: false,
    login: false,
    profile: false,
    setup: false,
    question: false,
    answer: false,
    report: false,
    history: false,
    home: false,
    questionReviews: false,
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
      user: state.user,
      config: {
        ...state.config,
        resumeText: "",
        resumeFileName: ""
      },
      companyOptions: state.companyOptions,
      planId: state.planId,
      sessionId: state.sessionId,
      selectedReportSessionId: state.selectedReportSessionId,
      currentQuestion: state.currentQuestion
    })
  );
}

function clearPersistedState() {
  if (typeof window === "undefined") {
    return;
  }

  window.localStorage.removeItem(STORAGE_KEY);
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
  state.history = [];
  state.answerHistoryBySession = {};
  state.questionReviewsBySession = {};
}

function getAnswerHistory(sessionId = state.sessionId) {
  if (!sessionId) {
    return [];
  }

  return state.answerHistoryBySession[String(sessionId)] || [];
}

function getQuestionReviews(sessionId = state.sessionId) {
  if (!sessionId) {
    return [];
  }

  return state.questionReviewsBySession[String(sessionId)] || [];
}

async function initAuth() {
  if (!state.isAuthenticated) {
    state.authReady = true;
    return null;
  }

  state.loading.auth = true;
  try {
    const user = await api.getCurrentUser();
    state.user = user;
    state.isAuthenticated = true;
    persist();
    return user;
  } catch {
    logoutLocal();
    return null;
  } finally {
    state.loading.auth = false;
    state.authReady = true;
  }
}

async function register(payload) {
  state.loading.register = true;
  setError("");

  try {
    const session = await api.register(payload);
    setAuthToken(session.token);
    state.user = session.user;
    state.isAuthenticated = true;
    persist();
    return session;
  } catch (error) {
    setError(error.message);
    throw error;
  } finally {
    state.loading.register = false;
  }
}

async function login(payload) {
  state.loading.login = true;
  setError("");

  try {
    const session = await api.login(payload);
    setAuthToken(session.token);
    state.user = session.user;
    state.isAuthenticated = true;
    persist();
    return session;
  } catch (error) {
    setError(error.message);
    throw error;
  } finally {
    state.loading.login = false;
  }
}

async function logout() {
  try {
    if (state.isAuthenticated) {
      await api.logout();
    }
  } catch {
    // ignore logout errors
  } finally {
    logoutLocal();
  }
}

async function updateCurrentUser(payload) {
  state.loading.profile = true;
  setError("");

  try {
    const user = await api.updateCurrentUser(payload);
    state.user = user;
    persist();
    return user;
  } catch (error) {
    setError(error.message);
    throw error;
  } finally {
    state.loading.profile = false;
  }
}

function logoutLocal() {
  setAuthToken("");
  state.user = null;
  state.isAuthenticated = false;
  state.authReady = true;
  state.aiConfig = null;
  resetInterviewState();
  clearPersistedState();
}

async function bootstrapHome() {
  state.loading.home = true;
  setError("");

  try {
    state.health = await api.health();
    state.history = await api.getHistory();
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
      jobTitle: state.config.jobTitle,
      jobCategory: state.config.jobCategory,
      levelCode: state.config.levelCode,
      interviewType: state.config.interviewType,
      companyName: state.config.companyName,
      resumeFileName: state.config.resumeFileName,
      resumeText: state.config.resumeText,
      interviewMode: state.config.interviewMode,
      difficultyLevel: state.config.difficultyLevel,
      questionCount: Number(state.config.questionCount),
      interviewerStyle: state.config.interviewerStyle
    });

    const sessionDetail = await api.createSession({
      jobTargetId: 0,
      jobTitle: state.config.jobTitle,
      jobCategory: state.config.jobCategory,
      levelCode: state.config.levelCode,
      interviewType: state.config.interviewType,
      companyName: state.config.companyName,
      resumeFileName: state.config.resumeFileName,
      resumeText: state.config.resumeText,
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

async function loadQuestionReviews(sessionId = state.sessionId) {
  if (!sessionId) {
    return [];
  }

  state.loading.questionReviews = true;
  setError("");

  try {
    const items = await api.getQuestionReviews(sessionId);
    state.questionReviewsBySession[String(sessionId)] = items;
    return items;
  } catch (error) {
    setError(error.message);
    throw error;
  } finally {
    state.loading.questionReviews = false;
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
    state.selectedReportSessionId = sessionId;
    persist();
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
    const items = await api.getHistory();
    state.history = items;
    return items;
  } catch (error) {
    setError(error.message);
    throw error;
  } finally {
    state.loading.history = false;
  }
}

async function loadAIConfig() {
  state.loading.aiConfig = true;
  setError("");

  try {
    const config = await api.getAIConfig();
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

function addCompanyOption(companyName) {
  const name = String(companyName || "").trim();
  if (!name) {
    throw new Error("公司名称不能为空");
  }

  if (state.companyOptions.some((item) => item === name)) {
    throw new Error("公司名称不能重复");
  }

  state.companyOptions = [...state.companyOptions, name];
  persist();
}

function selectSession(sessionId) {
  if (state.sessionId !== sessionId) {
    state.currentQuestion = null;
  }

  state.sessionId = sessionId;
  persist();
}

function selectReportSession(sessionId) {
  state.selectedReportSessionId = sessionId || null;
  persist();
}

function resetInterviewState() {
  state.planId = null;
  state.sessionId = null;
  state.selectedReportSessionId = null;
  clearRuntimeState();
  persist();
}

function openTab(path) {
  if (!path || path === "/login" || path === "/register") {
    return;
  }

  const currentTabs = state.openTabs.length ? state.openTabs : ["/"];
  const existingIndex = currentTabs.indexOf(path);

  if (existingIndex >= 0) {
    state.openTabs = currentTabs.slice(0, existingIndex + 1);
    persist();
    return;
  }

  state.openTabs = [...currentTabs, path];
  persist();
}

function closeTab(path) {
  if (path === "/") {
    return;
  }

  state.openTabs = state.openTabs.filter((item) => item !== path);
  persist();
}

export function useInterviewState() {
  return {
    state,
    setError,
    initAuth,
    register,
    login,
    logout,
    updateCurrentUser,
    updateConfig,
    addCompanyOption,
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
    getQuestionReviews,
    loadQuestionReviews,
    selectSession,
    selectReportSession,
    resetInterviewState,
    openTab,
    closeTab
  };
}

function normalizeCompanyOptions(items) {
  const source = Array.isArray(items) && items.length ? items : companyOptions;
  const unique = [];

  for (const item of source) {
    const name = String(item || "").trim();
    if (!name || unique.includes(name)) {
      continue;
    }
    unique.push(name);
  }

  return unique;
}
