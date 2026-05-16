const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || "http://127.0.0.1:8080";
const TOKEN_KEY = "ai-interview-admin-token";

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

function buildQuery(params = {}) {
  const query = new URLSearchParams();
  Object.entries(params).forEach(([key, value]) => {
    if (value == null || value === "") {
      return;
    }
    query.set(key, String(value));
  });
  const raw = query.toString();
  return raw ? `?${raw}` : "";
}

export const api = {
  login(body) {
    return request("/api/v1/admin/auth/login", {
      method: "POST",
      body: JSON.stringify(body)
    });
  },
  logout() {
    return request("/api/v1/admin/auth/logout", { method: "POST" });
  },
  getCurrentAdmin() {
    return request("/api/v1/admin/auth/me");
  },
  getOverview() {
    return request("/api/v1/admin/overview");
  },
  getUsers(params) {
    return request(`/api/v1/admin/users${buildQuery(params)}`);
  },
  getUserDetail(userId) {
    return request(`/api/v1/admin/users/${userId}`);
  },
  updateUserStatus(userId, body) {
    return request(`/api/v1/admin/users/${userId}/status`, {
      method: "PUT",
      body: JSON.stringify(body)
    });
  },
  resetUserPassword(userId, body) {
    return request(`/api/v1/admin/users/${userId}/reset-password`, {
      method: "POST",
      body: JSON.stringify(body)
    });
  },
  getSessions(params) {
    return request(`/api/v1/admin/interview-sessions${buildQuery(params)}`);
  },
  getSessionDetail(sessionId) {
    return request(`/api/v1/admin/interview-sessions/${sessionId}`);
  },
  getSessionAnswers(sessionId) {
    return request(`/api/v1/admin/interview-sessions/${sessionId}/answers`);
  },
  getSessionReport(sessionId) {
    return request(`/api/v1/admin/interview-reports/${sessionId}`);
  },
  getResumes(params) {
    return request(`/api/v1/admin/resumes${buildQuery(params)}`);
  },
  getResumeDetail(resumeId) {
    return request(`/api/v1/admin/resumes/${resumeId}`);
  },
  getAIConfigs(params) {
    return request(`/api/v1/admin/ai-configs${buildQuery(params)}`);
  }
};
