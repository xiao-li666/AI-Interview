<template>
  <div class="admin-auth-page">
    <section class="admin-auth-card">
      <div class="admin-auth-head">
        <span class="admin-brand-mark large">AI</span>
        <h1>管理员登录</h1>
        <p>使用独立管理员账号登录，查看全局用户、面试和模型配置数据。</p>
      </div>

      <div v-if="state.error" class="admin-banner error">{{ state.error }}</div>

      <form class="admin-form" @submit.prevent="handleSubmit">
        <label>
          <span>邮箱</span>
          <input v-model="form.email" type="email" autocomplete="username" />
        </label>
        <label>
          <span>密码</span>
          <input v-model="form.password" type="password" autocomplete="current-password" />
        </label>
        <button class="admin-button primary full" :disabled="state.loading.login">
          {{ state.loading.login ? "登录中..." : "登录后台" }}
        </button>
      </form>
    </section>
  </div>
</template>

<script setup>
import { reactive } from "vue";
import { useRouter } from "vue-router";
import { useAdminState } from "../state/admin";

const router = useRouter();
const { state, login, setError } = useAdminState();

const form = reactive({
  email: "",
  password: ""
});

async function handleSubmit() {
  setError("");
  try {
    await login(form);
    router.replace("/");
  } catch {
    // handled by store
  }
}
</script>
