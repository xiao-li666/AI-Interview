<template>
  <div class="admin-grid">
    <div v-if="state.error" class="admin-banner error">{{ state.error }}</div>

    <section class="admin-card">
      <div class="admin-card-head">
        <div>
          <h3>系统概览</h3>
          <p>查看管理员、用户、面试、简历、报告与 AI 配置的全局数量。</p>
        </div>
        <button type="button" class="admin-button ghost" @click="handleRefresh">刷新</button>
      </div>

      <div v-if="state.loading.overview" class="admin-empty">正在加载概览数据...</div>
      <div v-else class="metric-grid">
        <article class="metric-card" v-for="item in metrics" :key="item.label">
          <span>{{ item.label }}</span>
          <strong>{{ item.value }}</strong>
        </article>
      </div>
    </section>
  </div>
</template>

<script setup>
import { computed, onMounted } from "vue";
import { useAdminState } from "../state/admin";

const { state, loadOverview } = useAdminState();

const metrics = computed(() => {
  const overview = state.overview || {};
  return [
    { label: "管理员账号", value: overview.adminCount || 0 },
    { label: "普通用户", value: overview.userCount || 0 },
    { label: "面试会话", value: overview.sessionCount || 0 },
    { label: "简历数量", value: overview.resumeCount || 0 },
    { label: "复盘报告", value: overview.reportCount || 0 },
    { label: "AI 配置", value: overview.aiConfigCount || 0 }
  ];
});

async function handleRefresh() {
  await loadOverview();
}

onMounted(() => {
  handleRefresh();
});
</script>
