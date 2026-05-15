<template>
  <div class="page-grid">
    <section class="panel">
      <div class="panel-header">
        <h3>历史记录</h3>
        <p>这里展示的是已经写入 MySQL 的面试会话记录。</p>
      </div>

      <div v-if="state.error" class="message-banner message-error">
        {{ state.error }}
      </div>

      <div v-if="state.loading.history" class="message-banner">正在加载历史记录...</div>

      <div v-else-if="state.history.length" class="history-table">
        <div class="history-head">
          <span>创建时间</span>
          <span>岗位</span>
          <span>轮次</span>
          <span>状态</span>
          <span>得分</span>
        </div>

        <button
          v-for="item in state.history"
          :key="item.sessionId"
          class="history-row history-button"
          @click="openSession(item)"
        >
          <span>{{ formatDate(item.createdAt) }}</span>
          <span>{{ item.jobTitle }}</span>
          <span>{{ toRoundTypeLabel(item.roundType) }}</span>
          <span>{{ toStatusLabel(item.status) }}</span>
          <strong>{{ item.overallScore != null ? Math.round(item.overallScore) : "--" }}</strong>
        </button>
      </div>

      <div v-else class="empty-state">
        <p>目前还没有已保存的面试记录。</p>
      </div>
    </section>
  </div>
</template>

<script setup>
import { onMounted } from "vue";
import { useRouter } from "vue-router";
import { toRoundTypeLabel, toStatusLabel } from "../data/mock";
import { useInterviewState } from "../state/interview";

const router = useRouter();
const { state, loadHistory, selectSession, selectReportSession } = useInterviewState();

function formatDate(value) {
  if (!value) {
    return "--";
  }

  return new Date(value).toLocaleString();
}

function openSession(item) {
  selectReportSession(item.sessionId);

  if (item.status === "completed") {
    router.push({
      path: "/report",
      query: { sessionId: String(item.sessionId) }
    });
    return;
  }

  selectSession(item.sessionId);
  router.push({
    path: "/interview",
    query: { sessionId: String(item.sessionId) }
  });
}

onMounted(() => {
  loadHistory();
});
</script>
