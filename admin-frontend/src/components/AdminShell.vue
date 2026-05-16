<template>
  <div class="admin-shell">
    <aside class="admin-sidebar">
      <div class="admin-brand">
        <span class="admin-brand-mark">AI</span>
        <div>
          <strong>管理员后台</strong>
          <p>系统级运营与全局查看</p>
        </div>
      </div>

      <nav class="admin-nav">
        <RouterLink to="/" class="admin-nav-link" active-class="is-active">后台概览</RouterLink>
        <RouterLink to="/users" class="admin-nav-link" active-class="is-active">用户管理</RouterLink>
        <RouterLink to="/sessions" class="admin-nav-link" active-class="is-active">面试记录</RouterLink>
        <RouterLink to="/resumes" class="admin-nav-link" active-class="is-active">简历查看</RouterLink>
        <RouterLink to="/ai-configs" class="admin-nav-link" active-class="is-active">AI 配置</RouterLink>
      </nav>
    </aside>

    <main class="admin-main">
      <header class="admin-topbar">
        <div>
          <p class="admin-topbar-label">管理员工作台</p>
          <h1>{{ route.meta.title || "管理员后台" }}</h1>
        </div>

        <div class="admin-topbar-user">
          <div>
            <strong>{{ state.admin?.nickname || state.admin?.email || "管理员" }}</strong>
            <span>{{ state.admin?.email || "" }}</span>
          </div>
          <button type="button" class="admin-button ghost" @click="handleLogout">退出登录</button>
        </div>
      </header>

      <section class="admin-page">
        <slot />
      </section>
    </main>
  </div>
</template>

<script setup>
import { RouterLink, useRoute, useRouter } from "vue-router";
import { useAdminState } from "../state/admin";

const route = useRoute();
const router = useRouter();
const { state, logout } = useAdminState();

async function handleLogout() {
  await logout();
  router.replace("/login");
}
</script>
