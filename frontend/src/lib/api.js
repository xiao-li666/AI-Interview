const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || "http://127.0.0.1:8080";

async function request(path, options = {}) {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    headers: {
      "Content-Type": "application/json",
      ...(options.headers || {})
    },
    ...options
  });

  const payload = await response.json().catch(() => ({}));
  if (!response.ok || payload.code >= 400) {
    throw new Error(payload.message || "Request failed");
  }

  return payload.data;
}

export const api = {
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
  getHistory(userId) {
    return request(`/api/v1/history?userId=${userId}`);
  },
  getAIConfig(userId) {
    return request(`/api/v1/ai-config?userId=${userId}`);
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
  }
};

export { API_BASE_URL };
