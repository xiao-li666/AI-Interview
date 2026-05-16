import { createRouter, createWebHistory } from "vue-router";
import HomeView from "../views/HomeView.vue";
import SetupView from "../views/SetupView.vue";
import InterviewView from "../views/InterviewView.vue";
import ReportView from "../views/ReportView.vue";
import HistoryView from "../views/HistoryView.vue";
import AIConfigView from "../views/AIConfigView.vue";
import LoginView from "../views/LoginView.vue";
import RegisterView from "../views/RegisterView.vue";
import SettingsView from "../views/SettingsView.vue";
import { useInterviewState } from "../state/interview";

const routes = [
  { path: "/login", name: "login", component: LoginView, meta: { title: "登录", public: true } },
  { path: "/register", name: "register", component: RegisterView, meta: { title: "注册", public: true } },
  { path: "/", name: "home", component: HomeView, meta: { title: "主页" } },
  { path: "/setup", name: "setup", component: SetupView, meta: { title: "面试配置" } },
  { path: "/interview", name: "interview", component: InterviewView, meta: { title: "模拟面试" } },
  { path: "/report", name: "report", component: ReportView, meta: { title: "复盘报告" } },
  { path: "/history", name: "history", component: HistoryView, meta: { title: "历史记录" } },
  { path: "/ai-config", name: "ai-config", component: AIConfigView, meta: { title: "模型配置" } },
  { path: "/settings", name: "settings", component: SettingsView, meta: { title: "个人设置" } }
];

const router = createRouter({
  history: createWebHistory(),
  routes
});

router.beforeEach(async (to) => {
  const store = useInterviewState();
  const { state } = store;

  if (!state.authReady) {
    await store.initAuth();
  }

  if (to.meta.public) {
    if (state.isAuthenticated) {
      return "/";
    }
    return true;
  }

  if (!state.isAuthenticated) {
    return "/login";
  }

  const hasActiveInterviewSession =
    !!state.sessionId &&
    !!state.sessionDetail?.session &&
    state.sessionDetail.session.id === state.sessionId &&
    state.sessionDetail.session.status !== "completed";

  if (to.path === "/interview" && !to.query.sessionId && !hasActiveInterviewSession) {
    window.alert("请先进行面试配置");
    return "/setup";
  }

  if (to.path === "/report" && !to.query.sessionId && !state.selectedReportSessionId) {
    window.alert("请先选择复盘记录");
    return "/history";
  }

  return true;
});

router.afterEach((to) => {
  document.title = `智能模拟面试 - ${to.meta.title || "页面"}`;
});

export default router;
