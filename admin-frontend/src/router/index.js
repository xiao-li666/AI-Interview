import { createRouter, createWebHistory } from "vue-router";
import DashboardView from "../views/DashboardView.vue";
import UsersView from "../views/UsersView.vue";
import SessionsView from "../views/SessionsView.vue";
import ResumesView from "../views/ResumesView.vue";
import AIConfigsView from "../views/AIConfigsView.vue";
import LoginView from "../views/LoginView.vue";
import { useAdminState } from "../state/admin";

const routes = [
  { path: "/login", name: "login", component: LoginView, meta: { public: true, title: "管理员登录" } },
  { path: "/", name: "dashboard", component: DashboardView, meta: { title: "后台概览" } },
  { path: "/users", name: "users", component: UsersView, meta: { title: "用户管理" } },
  { path: "/sessions", name: "sessions", component: SessionsView, meta: { title: "面试记录" } },
  { path: "/resumes", name: "resumes", component: ResumesView, meta: { title: "简历查看" } },
  { path: "/ai-configs", name: "ai-configs", component: AIConfigsView, meta: { title: "AI 配置" } }
];

const router = createRouter({
  history: createWebHistory(),
  routes
});

router.beforeEach(async (to) => {
  const store = useAdminState();
  if (!store.state.authReady) {
    await store.initAuth();
  }

  if (to.meta.public) {
    if (store.state.isAuthenticated) {
      return "/";
    }
    return true;
  }

  if (!store.state.isAuthenticated) {
    return "/login";
  }

  return true;
});

router.afterEach((to) => {
  document.title = `AI 面试管理后台 - ${to.meta.title || "页面"}`;
});

export default router;
