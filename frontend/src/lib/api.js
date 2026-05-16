const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || "http://127.0.0.1:8080";
const TOKEN_KEY = "ai-interview-auth-token";

export function getAuthToken() {
  if (typeof window === "undefined") {
    return "";
  }

  return window.localStorage.getItem(TOKEN_KEY) || "";
}

export function setAuthToken(token) {
  if (typeof window === "undefined") {
    return;
  }

  if (!token) {
    window.localStorage.removeItem(TOKEN_KEY);
    return;
  }

  window.localStorage.setItem(TOKEN_KEY, token);
}

async function request(path, options = {}) {
  const headers = {
    ...(options.headers || {})
  };

  if (!(options.body instanceof FormData) && !headers["Content-Type"]) {
    headers["Content-Type"] = "application/json";
  }

  const token = getAuthToken();
  if (token) {
    headers.Authorization = `Bearer ${token}`;
  }

  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...options,
    headers
  });

  const payload = await response.json().catch(() => ({}));
  if (!response.ok || payload.code >= 400) {
    throw new Error(payload.message || "Request failed");
  }

  return payload.data;
}

export const api = {
  register(body) {
    return request("/api/v1/auth/register", {
      method: "POST",
      body: JSON.stringify(body)
    });
  },
  login(body) {
    return request("/api/v1/auth/login", {
      method: "POST",
      body: JSON.stringify(body)
    });
  },
  logout() {
    return request("/api/v1/auth/logout", {
      method: "POST"
    });
  },
  getCurrentUser() {
    return request("/api/v1/auth/me");
  },
  updateCurrentUser(body) {
    return request("/api/v1/auth/me", {
      method: "PUT",
      body: JSON.stringify(body)
    });
  },
  health() {
    return request("/healthz");
  },
  createPlan(body) {
    return request("/api/v1/interview-plans", {
      method: "POST",
      body: JSON.stringify(body)
    });
  },
  createSession(body) {
    return request("/api/v1/interview-sessions", {
      method: "POST",
      body: JSON.stringify(body)
    });
  },
  getSession(sessionId) {
    return request(`/api/v1/interview-sessions/${sessionId}`);
  },
  getSessionAnswers(sessionId) {
    return request(`/api/v1/interview-sessions/${sessionId}/answers`);
  },
  getQuestionReviews(sessionId) {
    return request(`/api/v1/interview-sessions/${sessionId}/question-reviews`);
  },
  getNextQuestion(sessionId) {
    return request(`/api/v1/interview-sessions/${sessionId}/questions/next`, {
      method: "POST"
    });
  },
  submitAnswer(body) {
    return request("/api/v1/interview-answers", {
      method: "POST",
      body: JSON.stringify(body)
    });
  },
  getReport(sessionId) {
    return request(`/api/v1/interview-reports/${sessionId}`);
  },
  getHistory() {
    return request("/api/v1/history");
  },
  getAIConfig() {
    return request("/api/v1/ai-config");
  },
  saveAIConfig(body) {
    return request("/api/v1/ai-config", {
      method: "PUT",
      body: JSON.stringify(body)
    });
  },
  testAIConfig(body) {
    return request("/api/v1/ai-config/test", {
      method: "POST",
      body: JSON.stringify(body)
    });
  },
  async parseResume(file) {
    const formData = new FormData();
    formData.append("file", file);

    return request("/api/v1/resume/parse", {
      method: "POST",
      body: formData
    });
  }
};

export { API_BASE_URL, TOKEN_KEY };
