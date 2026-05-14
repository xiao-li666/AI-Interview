<template>
  <div class="page-grid">
    <section class="hero-card">
      <p class="eyebrow">实时链路</p>
      <h2>现在这版不只是原型，而是一条真的能跑通的模拟面试流程。</h2>
      <p class="lead">
        当前前端已经接通 Go 后端。你可以创建面试计划、开始模拟面试、提交回答、生成复盘报告，并从 MySQL
        里查看历史记录。
      </p>
      <div class="hero-actions">
        <RouterLink to="/setup" class="primary-button">去配置面试</RouterLink>
        <RouterLink to="/history" class="ghost-button">查看历史记录</RouterLink>
      </div>
    </section>

    <section class="stats-grid">
      <article class="metric-card">
        <p class="metric-label">接口状态</p>
        <p class="metric-value">{{ homeStats.apiStatus }}</p>
        <p class="metric-hint">来自 `/healthz` 检查</p>
      </article>
      <article class="metric-card">
        <p class="metric-label">历史场次</p>
        <p class="metric-value">{{ homeStats.sessionCount }}</p>
        <p class="metric-hint">来自 `GET /api/v1/history`</p>
      </article>
      <article class="metric-card">
        <p class="metric-label">最近得分</p>
        <p class="metric-value">{{ homeStats.latestScore }}</p>
        <p class="metric-hint">最近一次已完成报告</p>
      </article>
    </section>

    <section class="panel">
      <div class="panel-header">
        <h3>当前流程状态</h3>
        <span v-if="state.loading.home" class="status-chip status-pending">加载中</span>
        <span v-else class="status-chip status-success">已就绪</span>
      </div>

      <div v-if="state.error" class="message-banner message-error">
        {{ state.error }}
      </div>

      <div class="feature-list">
        <article class="feature-card">
          <span class="feature-index">01</span>
          <h4>面试配置</h4>
          <p>先配置岗位、轮次、难度和题量，再创建一场真实会话。</p>
        </article>
        <article class="feature-card">
          <span class="feature-index">02</span>
          <h4>模拟面试</h4>
          <p>从后端实时拉题，逐题作答，并在过程中拿到单题反馈。</p>
        </article>
        <article class="feature-card">
          <span class="feature-index">03</span>
          <h4>复盘报告</h4>
          <p>读取维度评分、改进建议和已保存到 MySQL 的完整历史。</p>
        </article>
      </div>
    </section>
  </div>
</template>

<script setup>
import { computed, onMounted } from "vue";
import { RouterLink } from "vue-router";
import { useInterviewState } from "../state/interview";

const { state, bootstrapHome } = useInterviewState();

const homeStats = computed(() => {
  const latestCompleted = state.history.find((item) => item.overallScore != null);

  return {
    apiStatus: state.health?.status === "ok" ? "正常" : "--",
    sessionCount: String(state.history.length || 0),
    latestScore:
      latestCompleted?.overallScore != null ? String(Math.round(latestCompleted.overallScore)) : "--"
  };
});

onMounted(() => {
  bootstrapHome();
});
</script>
