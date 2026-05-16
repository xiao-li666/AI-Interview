<template>
  <div class="admin-grid">
    <section class="admin-card">
      <div class="admin-card-head">
        <div>
          <h3>AI 配置查看</h3>
          <p>查看全局模型配置使用情况，列表中只展示脱敏后的 Key。</p>
        </div>
      </div>

      <div class="toolbar">
        <input v-model="filters.keyword" type="text" placeholder="搜索用户、Provider 或模型名" />
        <select v-model="filters.status">
          <option value="">全部状态</option>
          <option value="active">启用中</option>
          <option value="disabled">已停用</option>
        </select>
        <button type="button" class="admin-button ghost" @click="handleSearch">查询</button>
      </div>

      <div v-if="state.loading.aiConfigs" class="admin-empty">正在加载配置列表...</div>
      <div v-else class="table-list">
        <div v-for="item in state.aiConfigs.items" :key="item.id" class="table-row">
          <span>{{ item.userEmail }}</span>
          <span>{{ item.provider }}</span>
          <span>{{ item.model }}</span>
          <span>{{ item.apiKeyMasked }}</span>
          <span>{{ item.isEnabled ? "启用" : "停用" }}</span>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { onMounted, reactive } from "vue";
import { useAdminState } from "../state/admin";

const { state, loadAIConfigs } = useAdminState();
const filters = reactive({
  keyword: "",
  status: ""
});

async function handleSearch() {
  await loadAIConfigs({ page: 1, keyword: filters.keyword, status: filters.status });
}

onMounted(() => {
  handleSearch();
});
</script>
