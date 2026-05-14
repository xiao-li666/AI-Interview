<template>
  <div class="page-grid">
    <section v-if="state.error" class="message-banner message-error">
      {{ state.error }}
    </section>

    <section v-if="state.loading.report" class="message-banner">正在加载复盘报告...</section>

    <template v-else-if="state.report">
      <section class="stats-grid report-stats-grid">
        <article class="metric-card">
          <p class="metric-label">综合得分</p>
          <p class="metric-value">{{ Math.round(state.report.overallScore) }}</p>
        </article>
        <article
          v-for="item in state.report.dimensionScores"
          :key="item.dimensionCode"
          class="metric-card"
        >
          <p class="metric-label">{{ toDimensionLabel(item.dimensionCode) }}</p>
          <p class="metric-value">{{ Math.round(item.score) }}</p>
          <p class="metric-hint">{{ item.comment }}</p>
        </article>
      </section>

      <section class="panel">
        <div class="panel-header">
          <h3>综合总结</h3>
        </div>
        <div class="report-copy">
          <p>{{ state.report.performanceSummary }}</p>
          <p><strong>亮点：</strong>{{ state.report.strengthsSummary }}</p>
          <p><strong>薄弱项：</strong>{{ state.report.weaknessSummary }}</p>
          <p><strong>下一步建议：</strong>{{ state.report.recommendedActions }}</p>
        </div>
      </section>

      <section class="report-grid">
        <article class="panel">
          <div class="panel-header">
            <h3>维度图表</h3>
          </div>
          <div class="chart-list">
            <div
              v-for="item in state.report.dimensionScores"
              :key="item.dimensionCode"
              class="chart-row"
            >
              <div class="chart-row-head">
                <span>{{ toDimensionLabel(item.dimensionCode) }}</span>
                <strong>{{ item.score.toFixed(2) }}</strong>
              </div>
              <div class="chart-bar">
                <div class="chart-bar-fill" :style="{ width: `${Math.min(item.score, 100)}%` }"></div>
              </div>
              <p class="chart-comment">{{ item.comment }}</p>
            </div>
          </div>
        </article>

        <article class="panel">
          <div class="panel-header">
            <h3>下一轮训练计划</h3>
          </div>
          <div class="tag-list">
            <span
              v-for="item in state.report.nextPlan?.focusAreas || []"
              :key="item"
              class="tag"
            >
              {{ focusAreaText(item) }}
            </span>
          </div>
          <p class="report-copy single-line-copy">
            推荐题量：{{ state.report.nextPlan?.recommendedQuestionCount || "--" }}
          </p>
        </article>
      </section>
    </template>

    <section v-else class="panel">
      <div class="panel-header">
        <h3>暂无报告</h3>
      </div>
      <p class="lead">先完成一场面试，这里才会显示对应的复盘报告。</p>
    </section>
  </div>
</template>

<script setup>
import { onMounted } from "vue";
import { useRoute } from "vue-router";
import { toDimensionLabel } from "../data/mock";
import { useInterviewState } from "../state/interview";

const route = useRoute();
const { state, selectSession, loadReport } = useInterviewState();

function focusAreaText(value) {
  const map = {
    high_concurrency: "高并发设计",
    cache_consistency: "缓存一致性",
    project_depth: "项目深度表达"
  };

  return map[value] || value;
}

onMounted(() => {
  const sessionId = Number(route.query.sessionId || state.sessionId);
  if (!sessionId) {
    return;
  }

  selectSession(sessionId);
  loadReport(sessionId);
});
</script>
