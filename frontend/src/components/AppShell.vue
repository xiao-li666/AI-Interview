<template>
  <div class="app-shell">
    <aside class="sidebar">
      <div class="brand">
        <span class="brand-mark">AI</span>
        <div>
          <p class="brand-title">模拟面试</p>
          <p class="brand-subtitle">Go API + Vue 3 前端</p>
        </div>
      </div>

      <nav class="nav-list">
        <RouterLink
          v-for="item in shellNavItems"
          :key="item.to"
          :to="item.to"
          class="nav-link"
          active-class="is-active"
        >
          {{ item.label }}
        </RouterLink>
      </nav>

      <div class="sidebar-card">
        <p class="sidebar-card-title">当前能力</p>
        <p class="sidebar-card-text">
          当前前端已经接通实时 Go 接口，并将面试计划、会话、题目、回答和报告落到 MySQL。
        </p>
      </div>
    </aside>

    <main class="content">
      <header class="topbar">
        <div>
          <p class="eyebrow">AI 模拟面试训练器</p>
          <h1>真实面试流程</h1>
        </div>
        <RouterLink to="/setup" class="primary-button">开始新面试</RouterLink>
      </header>

      <section class="page-body">
        <slot />
      </section>
    </main>
  </div>
</template>

<script setup>
import { computed } from "vue";
import { RouterLink } from "vue-router";
import { navItems } from "../data/mock";

const shellNavItems = computed(() => [
  ...navItems,
  { label: "模型配置", to: "/ai-config" }
]);
</script>
