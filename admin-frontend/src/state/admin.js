import { reactive } from "vue";
import { api, getAuthToken, setAuthToken } from "../lib/api";

const state = reactive({
  admin: null,
  isAuthenticated: Boolean(getAuthToken()),
  authReady: false,
  error: "",
  overview: null,
  users: { items: [], total: 0, page: 1, pageSize: 10, keyword: "", status: "" },
  selectedUser: null,
  sessions: { items: [], total: 0, page: 1, pageSize: 10, keyword: "", status: "" },
  selectedSession: null,
  sessionAnswers: [],
  sessionReport: null,
  resumes: { items: [], total: 0, page: 1, pageSize: 10, keyword: "" },
  selectedResume: null,
  aiConfigs: { items: [], total: 0, page: 1, pageSize: 10, keyword: "", status: "" },
  loading: {
    auth: false,
    login: false,
    overview: false,
    users: false,
    userDetail: false,
    userAction: false,
    sessions: false,
    sessionDetail: false,
    resumes: false,
    resumeDetail: false,
    aiConfigs: false
  }
});

function setError(message) {
  state.error = message || "";
}

async function initAuth() {
  if (!state.isAuthenticated) {
    state.authReady = true;
    return null;
  }

  state.loading.auth = true;
  try {
    const admin = await api.getCurrentAdmin();
    state.admin = admin;
    state.isAuthenticated = true;
    return admin;
  } catch {
    logoutLocal();
    return null;
  } finally {
    state.loading.auth = false;
    state.authReady = true;
  }
}

async function login(payload) {
  state.loading.login = true;
  setError("");
  try {
    const session = await api.login(payload);
    setAuthToken(session.token);
    state.admin = session.admin;
    state.isAuthenticated = true;
    state.authReady = true;
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
    // ignore
  } finally {
    logoutLocal();
  }
}

function logoutLocal() {
  setAuthToken("");
  state.admin = null;
  state.isAuthenticated = false;
  state.authReady = true;
  state.overview = null;
  state.selectedUser = null;
  state.selectedSession = null;
  state.sessionAnswers = [];
  state.sessionReport = null;
  state.selectedResume = null;
}

async function loadOverview() {
  state.loading.overview = true;
  setError("");
  try {
    state.overview = await api.getOverview();
    return state.overview;
  } catch (error) {
    setError(error.message);
    throw error;
  } finally {
    state.loading.overview = false;
  }
}

async function loadUsers(patch = {}) {
  state.loading.users = true;
  setError("");
  try {
    state.users = { ...state.users, ...patch };
    const data = await api.getUsers(state.users);
    state.users = { ...state.users, ...data };
    return data;
  } catch (error) {
    setError(error.message);
    throw error;
  } finally {
    state.loading.users = false;
  }
}

async function loadUserDetail(userId) {
  state.loading.userDetail = true;
  setError("");
  try {
    state.selectedUser = await api.getUserDetail(userId);
    return state.selectedUser;
  } catch (error) {
    setError(error.message);
    throw error;
  } finally {
    state.loading.userDetail = false;
  }
}

async function updateUserStatus(userId, status) {
  state.loading.userAction = true;
  setError("");
  try {
    const user = await api.updateUserStatus(userId, { status });
    if (state.selectedUser?.user?.id === userId) {
      state.selectedUser = { ...state.selectedUser, user };
    }
    await loadUsers();
    return user;
  } catch (error) {
    setError(error.message);
    throw error;
  } finally {
    state.loading.userAction = false;
  }
}

async function resetUserPassword(userId, newPassword) {
  state.loading.userAction = true;
  setError("");
  try {
    return await api.resetUserPassword(userId, { newPassword });
  } catch (error) {
    setError(error.message);
    throw error;
  } finally {
    state.loading.userAction = false;
  }
}

async function loadSessions(patch = {}) {
  state.loading.sessions = true;
  setError("");
  try {
    state.sessions = { ...state.sessions, ...patch };
    const data = await api.getSessions(state.sessions);
    state.sessions = { ...state.sessions, ...data };
    return data;
  } catch (error) {
    setError(error.message);
    throw error;
  } finally {
    state.loading.sessions = false;
  }
}

async function loadSessionDetail(sessionId) {
  state.loading.sessionDetail = true;
  setError("");
  try {
    const [detail, answers, report] = await Promise.all([
      api.getSessionDetail(sessionId),
      api.getSessionAnswers(sessionId),
      api.getSessionReport(sessionId).catch(() => null)
    ]);
    state.selectedSession = detail;
    state.sessionAnswers = answers;
    state.sessionReport = report;
    return detail;
  } catch (error) {
    setError(error.message);
    throw error;
  } finally {
    state.loading.sessionDetail = false;
  }
}

async function loadResumes(patch = {}) {
  state.loading.resumes = true;
  setError("");
  try {
    state.resumes = { ...state.resumes, ...patch };
    const data = await api.getResumes(state.resumes);
    state.resumes = { ...state.resumes, ...data };
    return data;
  } catch (error) {
    setError(error.message);
    throw error;
  } finally {
    state.loading.resumes = false;
  }
}

async function loadResumeDetail(resumeId) {
  state.loading.resumeDetail = true;
  setError("");
  try {
    state.selectedResume = await api.getResumeDetail(resumeId);
    return state.selectedResume;
  } catch (error) {
    setError(error.message);
    throw error;
  } finally {
    state.loading.resumeDetail = false;
  }
}

async function loadAIConfigs(patch = {}) {
  state.loading.aiConfigs = true;
  setError("");
  try {
    state.aiConfigs = { ...state.aiConfigs, ...patch };
    const data = await api.getAIConfigs(state.aiConfigs);
    state.aiConfigs = { ...state.aiConfigs, ...data };
    return data;
  } catch (error) {
    setError(error.message);
    throw error;
  } finally {
    state.loading.aiConfigs = false;
  }
}

export function useAdminState() {
  return {
    state,
    setError,
    initAuth,
    login,
    logout,
    loadOverview,
    loadUsers,
    loadUserDetail,
    updateUserStatus,
    resetUserPassword,
    loadSessions,
    loadSessionDetail,
    loadResumes,
    loadResumeDetail,
    loadAIConfigs
  };
}
