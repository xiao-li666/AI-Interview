<template>
  <div class="auth-page">
    <section class="auth-card">
      <div class="auth-card-head">
        <img :src="brandLogo" alt="智能模拟面试" class="auth-logo" />
        <h1>创建账号</h1>
        <p>注册后你的面试记录、简历和模型配置都会按账号隔离保存。</p>
      </div>

      <div v-if="state.error" class="message-banner message-error">
        {{ state.error }}
      </div>

      <form class="field-list" @submit.prevent="handleSubmit">
        <label class="field">
          <span>昵称</span>
          <input v-model="form.nickname" type="text" autocomplete="nickname" />
        </label>
        <label class="field">
          <span>邮箱</span>
          <input v-model="form.email" type="email" autocomplete="username" />
        </label>
        <label class="field">
          <span>密码</span>
          <input v-model="form.password" type="password" autocomplete="new-password" />
        </label>
        <button class="primary-button auth-button" :disabled="state.loading.register">
          {{ state.loading.register ? "注册中..." : "注册并登录" }}
        </button>
      </form>

      <p class="auth-switch">
        已有账号？
        <RouterLink to="/login">去登录</RouterLink>
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
const { state, register, setError } = useInterviewState();

const form = reactive({
  nickname: "",
  email: "",
  password: ""
});

async function handleSubmit() {
  setError("");
  try {
    await register(form);
    router.replace("/");
  } catch {
    // store already set error
  }
}
</script>
