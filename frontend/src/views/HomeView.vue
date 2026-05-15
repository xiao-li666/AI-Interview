<template>
  <div class="page-grid home-page">
    <section class="hero-card dashboard-hero">
      <div class="hero-visual" aria-hidden="true">
        <div class="hero-glow"></div>
        <div class="hero-orb hero-orb-a"></div>
        <div class="hero-orb hero-orb-b"></div>
        <div class="hero-rings">
          <span></span>
          <span></span>
          <span></span>
        </div>
        <div class="hero-hill hero-hill-a"></div>
        <div class="hero-hill hero-hill-b"></div>
      </div>

      <div class="hero-overlay">
        <div class="hero-copy">
          <p class="eyebrow">沉浸式练习空间</p>
          <h2>让每一场模拟面试，都更安静、更顺滑，也更接近真实沟通。</h2>
          <p class="lead">
            从配置岗位、进入答题，到查看历史记录与复盘报告，整条训练链路已经打通。
            你可以把这里当作一个稳定的日常练习台，随时开始，随时回看。
          </p>
        </div>

        <div class="hero-floating-actions">
          <RouterLink to="/setup" class="primary-button">配置面试</RouterLink>
          <RouterLink to="/history" class="ghost-button">历史记录</RouterLink>
        </div>
      </div>
    </section>

    <section class="stats-grid dashboard-stats-grid">
      <article class="metric-card status-card" :class="healthStatus.className">
        <div class="status-card-head">
          <p class="metric-label">服务状态</p>
          <span class="status-dot"></span>
        </div>
        <p class="metric-value">{{ healthStatus.label }}</p>
      </article>

      <article class="metric-card overview-card">
        <p class="metric-label">历史场次</p>
        <p class="metric-value">{{ homeStats.sessionCount }}</p>
      </article>

      <article class="metric-card score-status-card" :class="scoreStatus.className">
        <div class="score-status-face">{{ scoreStatus.face }}</div>
        <div class="score-status-copy">
          <p class="metric-label">最近得分</p>
          <p class="metric-value">{{ homeStats.latestScore }}</p>
          <p class="score-status-note">{{ scoreStatus.note }}</p>
        </div>
      </article>
    </section>

    <section class="panel flow-panel">
      <div class="panel-header">
        <h3>使用流程</h3>
        <span v-if="state.loading.home" class="status-chip status-pending">加载中</span>
        <span v-else class="status-chip status-success">已就绪</span>
      </div>

      <div v-if="state.error" class="message-banner message-error">
        {{ state.error }}
      </div>

      <div class="flow-diagram">
        <article v-for="step in flowSteps" :key="step.index" class="flow-step">
          <div class="flow-step-index">{{ step.index }}</div>
          <div class="flow-step-body">
            <div class="flow-step-head">
              <h4>{{ step.title }}</h4>
              <span class="flow-step-tag">{{ step.tag }}</span>
            </div>
            <p>{{ step.description }}</p>
            <span class="flow-step-tip">操作建议：{{ step.tip }}</span>
          </div>
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

const flowSteps = [
  {
    index: "01",
    title: "面试配置",
    tag: "创建会话",
    description: "先确定岗位、难度、轮次、题量和面试风格，系统会据此创建一场新的模拟面试。",
    tip: "进入“面试配置”填写基础信息，确认后生成本次会话。"
  },
  {
    index: "02",
    title: "模拟面试",
    tag: "逐题作答",
    description: "开始后系统会按顺序出题。每次作答提交后，当前答案与反馈都会同步保存，方便后续复盘。",
    tip: "进入“模拟面试”逐题回答，保持连续完成可以获得更完整的评估。"
  },
  {
    index: "03",
    title: "历史记录",
    tag: "选择记录",
    description: "所有完成或进行中的会话都会保存在历史记录中，你可以从这里快速回看具体场次。",
    tip: "先在“历史记录”中选择要查看的场次，再进入复盘报告。"
  },
  {
    index: "04",
    title: "复盘报告",
    tag: "总结提升",
    description: "系统会结合整场问答内容，给出总结、维度评分和改进建议，帮助你定位优势与短板。",
    tip: "确认已选中对应记录后，再打开“复盘报告”查看完整分析。"
  }
];

const homeStats = computed(() => {
  const latestCompleted = state.history.find((item) => item.overallScore != null);

  return {
    sessionCount: String(state.history.length || 0),
    latestScore:
      latestCompleted?.overallScore != null ? String(Math.round(latestCompleted.overallScore)) : "--"
  };
});

const healthStatus = computed(() => {
  if (state.loading.home && !state.health) {
    return { label: "检测中", className: "is-neutral" };
  }

  if (state.health?.status === "ok") {
    return { label: "正常", className: "is-good" };
  }

  return { label: "异常", className: "is-bad" };
});

const scoreStatus = computed(() => {
  const score = Number(homeStats.value.latestScore);

  if (Number.isNaN(score)) {
    return { face: "•", className: "is-neutral", note: "还没有可展示的成绩" };
  }

  if (score < 60) {
    return { face: "😭", className: "is-bad", note: "需要重点补强" };
  }

  if (score <= 80) {
    return { face: "😐", className: "is-warn", note: "整体稳定，仍有提升空间" };
  }

  return { face: "😊", className: "is-good", note: "状态很好，继续保持" };
});

onMounted(() => {
  bootstrapHome();
});
</script>
