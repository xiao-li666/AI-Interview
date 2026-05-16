<template>
  <div class="admin-grid two-column">
    <section class="admin-card">
      <div class="admin-card-head">
        <div>
          <h3>简历查看</h3>
          <p>按用户或简历名搜索，查看当前存储的简历文本内容。</p>
        </div>
      </div>

      <div class="toolbar">
        <input v-model="keyword" type="text" placeholder="搜索用户或简历名" />
        <button type="button" class="admin-button ghost" @click="handleSearch">查询</button>
      </div>

      <div v-if="state.loading.resumes" class="admin-empty">正在加载简历...</div>
      <div v-else class="table-list">
        <button
          v-for="item in state.resumes.items"
          :key="item.id"
          type="button"
          class="table-row button-row"
          :class="{ 'is-selected': state.selectedResume?.id === item.id }"
          @click="handleSelect(item.id)"
        >
          <span>{{ item.name }}</span>
          <span>{{ item.userEmail }}</span>
          <span>{{ item.sourceType }}</span>
          <span>{{ formatTime(item.updatedAt) }}</span>
        </button>
      </div>
    </section>

    <section class="admin-card">
      <div class="admin-card-head">
        <div>
          <h3>简历详情</h3>
          <p>直接查看当前系统中保存的简历正文。</p>
        </div>
      </div>

      <div v-if="state.loading.resumeDetail" class="admin-empty">正在加载详情...</div>
      <div v-else-if="!state.selectedResume" class="admin-empty">请选择左侧简历。</div>
      <div v-else class="detail-stack">
        <div class="detail-grid">
          <article class="detail-card">
            <span>所属用户</span>
            <strong>{{ state.selectedResume.user.email }}</strong>
          </article>
          <article class="detail-card">
            <span>简历名称</span>
            <strong>{{ state.selectedResume.name }}</strong>
          </article>
          <article class="detail-card">
            <span>来源类型</span>
            <strong>{{ state.selectedResume.sourceType }}</strong>
          </article>
          <article class="detail-card">
            <span>更新时间</span>
            <strong>{{ formatTime(state.selectedResume.updatedAt) }}</strong>
          </article>
        </div>

        <div class="detail-card">
          <h4>简历正文</h4>
          <pre class="detail-pre">{{ state.selectedResume.rawText || "暂无正文内容" }}</pre>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { onMounted, ref } from "vue";
import { useAdminState } from "../state/admin";

const { state, loadResumes, loadResumeDetail } = useAdminState();
const keyword = ref("");

function formatTime(value) {
  return value ? new Date(value).toLocaleString("zh-CN") : "-";
}

async function handleSearch() {
  await loadResumes({ page: 1, keyword: keyword.value });
}

async function handleSelect(resumeId) {
  await loadResumeDetail(resumeId);
}

onMounted(() => {
  handleSearch();
});
</script>
