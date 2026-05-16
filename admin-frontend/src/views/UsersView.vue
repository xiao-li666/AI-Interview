<template>
  <div class="admin-grid two-column">
    <section class="admin-card">
      <div class="admin-card-head">
        <div>
          <h3>用户管理</h3>
          <p>支持按关键词和状态筛选，查看详情、禁用账号和重置密码。</p>
        </div>
      </div>

      <div class="toolbar">
        <input v-model="filters.keyword" type="text" placeholder="搜索邮箱或昵称" />
        <select v-model="filters.status">
          <option value="">全部状态</option>
          <option value="active">正常</option>
          <option value="disabled">已禁用</option>
        </select>
        <button type="button" class="admin-button ghost" @click="handleSearch">查询</button>
      </div>

      <div v-if="state.loading.users" class="admin-empty">正在加载用户列表...</div>
      <div v-else class="table-list">
        <button
          v-for="item in state.users.items"
          :key="item.id"
          type="button"
          class="table-row button-row"
          :class="{ 'is-selected': state.selectedUser?.user?.id === item.id }"
          @click="handleSelect(item.id)"
        >
          <span>{{ item.nickname }}</span>
          <span>{{ item.email }}</span>
          <span>{{ item.status }}</span>
          <span>{{ formatTime(item.createdAt) }}</span>
        </button>
      </div>
    </section>

    <section class="admin-card">
      <div class="admin-card-head">
        <div>
          <h3>用户详情</h3>
          <p>查看当前用户的基础信息和资源统计。</p>
        </div>
      </div>

      <div v-if="state.loading.userDetail" class="admin-empty">正在加载详情...</div>
      <div v-else-if="!state.selectedUser" class="admin-empty">请选择左侧用户。</div>
      <div v-else class="detail-stack">
        <div class="detail-grid">
          <article class="detail-card">
            <span>昵称</span>
            <strong>{{ state.selectedUser.user.nickname }}</strong>
          </article>
          <article class="detail-card">
            <span>邮箱</span>
            <strong>{{ state.selectedUser.user.email }}</strong>
          </article>
          <article class="detail-card">
            <span>状态</span>
            <strong>{{ state.selectedUser.user.status }}</strong>
          </article>
          <article class="detail-card">
            <span>创建时间</span>
            <strong>{{ formatTime(state.selectedUser.user.createdAt) }}</strong>
          </article>
          <article class="detail-card">
            <span>面试会话</span>
            <strong>{{ state.selectedUser.sessionCount }}</strong>
          </article>
          <article class="detail-card">
            <span>简历数量</span>
            <strong>{{ state.selectedUser.resumeCount }}</strong>
          </article>
          <article class="detail-card">
            <span>复盘报告</span>
            <strong>{{ state.selectedUser.reportCount }}</strong>
          </article>
          <article class="detail-card">
            <span>AI 配置</span>
            <strong>{{ state.selectedUser.aiConfigCount }}</strong>
          </article>
        </div>

        <div class="toolbar">
          <button
            type="button"
            class="admin-button primary"
            :disabled="state.loading.userAction"
            @click="toggleStatus"
          >
            {{ state.selectedUser.user.status === "active" ? "禁用用户" : "启用用户" }}
          </button>
        </div>

        <div class="detail-card">
          <h4>重置密码</h4>
          <div class="toolbar compact">
            <input v-model="newPassword" type="text" placeholder="输入新的登录密码" />
            <button
              type="button"
              class="admin-button ghost"
              :disabled="state.loading.userAction"
              @click="handleResetPassword"
            >
              重置
            </button>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from "vue";
import { useAdminState } from "../state/admin";

const { state, loadUsers, loadUserDetail, updateUserStatus, resetUserPassword } = useAdminState();
const newPassword = ref("");
const filters = reactive({
  keyword: "",
  status: ""
});

function formatTime(value) {
  return value ? new Date(value).toLocaleString("zh-CN") : "-";
}

async function handleSearch() {
  await loadUsers({ page: 1, keyword: filters.keyword, status: filters.status });
}

async function handleSelect(userId) {
  await loadUserDetail(userId);
}

async function toggleStatus() {
  const nextStatus = state.selectedUser.user.status === "active" ? "disabled" : "active";
  await updateUserStatus(state.selectedUser.user.id, nextStatus);
}

async function handleResetPassword() {
  if (!newPassword.value.trim()) {
    return;
  }
  await resetUserPassword(state.selectedUser.user.id, newPassword.value.trim());
  window.alert("密码已重置");
  newPassword.value = "";
}

onMounted(() => {
  handleSearch();
});
</script>
