<template>
  <div class="auth-page">
    <section class="auth-card">
      <div class="auth-card-head">
        <img :src="brandLogo" alt="智能模拟面试" class="auth-logo" />
        <h1>登录智能模拟面试</h1>
        <p>登录后查看你的面试记录、AI 配置和复盘内容。</p>
      </div>

      <div v-if="state.error" class="message-banner message-error">
        {{ state.error }}
      </div>

      <form class="field-list" @submit.prevent="handleSubmit">
        <label class="field">
          <span>邮箱</span>
          <input v-model="form.email" type="email" autocomplete="username" />
        </label>
        <label class="field">
          <span>密码</span>
          <input v-model="form.password" type="password" autocomplete="current-password" />
        </label>
        <button class="primary-button auth-button" :disabled="state.loading.login">
          {{ state.loading.login ? "登录中..." : "登录" }}
        </button>
      </form>

      <p class="auth-switch">
        没有账号？
        <RouterLink to="/register">去注册</RouterLink>
      </p>
    </section>
  </div>
</template>

<script setup>
import { reactive } from "vue";
import { RouterLink, useRouter } from "vue-router";
import brandLogo from "../assets/brand-logo.svg";
import { useInterviewState } from "../state/interview";

const router = useRouter();
const { state, login, setError } = useInterviewState();

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
    // store already set error
  }
}
</script>
