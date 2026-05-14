import { createRouter, createWebHistory } from "vue-router";
import HomeView from "../views/HomeView.vue";
import SetupView from "../views/SetupView.vue";
import InterviewView from "../views/InterviewView.vue";
import ReportView from "../views/ReportView.vue";
import HistoryView from "../views/HistoryView.vue";
import AIConfigView from "../views/AIConfigView.vue";

const routes = [
  {
    path: "/",
    name: "home",
    component: HomeView,
    meta: { title: "首页" }
  },
  {
    path: "/setup",
    name: "setup",
    component: SetupView,
    meta: { title: "面试配置" }
  },
  {
    path: "/interview",
    name: "interview",
    component: InterviewView,
    meta: { title: "模拟面试" }
  },
  {
    path: "/report",
    name: "report",
    component: ReportView,
    meta: { title: "复盘报告" }
  },
  {
    path: "/history",
    name: "history",
    component: HistoryView,
    meta: { title: "历史记录" }
  },
  {
    path: "/ai-config",
    name: "ai-config",
    component: AIConfigView,
    meta: { title: "模型配置" }
  }
];

const router = createRouter({
  history: createWebHistory(),
  routes
});

router.afterEach((to) => {
  document.title = `AI 模拟面试 - ${to.meta.title || "页面"}`;
});

export default router;
