<template>
  <div class="app-shell">
    <aside class="sidebar">
      <div class="brand">
        <span class="brand-mark">
          <img :src="brandLogo" alt="智能模拟面试" class="brand-mark-image" />
        </span>
        <div>
          <p class="brand-title">智能模拟面试</p>
        </div>
      </div>

      <nav class="nav-tree">
        <div v-for="group in navTree" :key="group.label" class="nav-group">
          <button
            v-if="group.children?.length"
            type="button"
            class="nav-group-link nav-group-button"
            :class="{ 'is-active': isExpanded(group.label), 'is-current': isGroupCurrent(group) }"
            @click="toggleGroup(group.label)"
          >
            <span>{{ group.label }}</span>
            <span class="nav-group-arrow" :class="{ 'is-expanded': isExpanded(group.label) }">›</span>
          </button>

          <RouterLink
            v-else
            :to="group.to"
            class="nav-group-link"
            active-class="is-active"
            @click.prevent="handleNavClick(group)"
          >
            <span>{{ group.label }}</span>
          </RouterLink>

          <div v-if="group.children?.length && isExpanded(group.label)" class="nav-children">
            <RouterLink
              v-for="item in group.children"
              :key="item.to"
              :to="item.to"
              class="nav-child-link"
              active-class="is-active"
              @click.prevent="handleNavClick(item)"
            >
              {{ item.label }}
            </RouterLink>
          </div>
        </div>
      </nav>
    </aside>

    <main class="content">
      <header class="topbar topbar-layout">
        <div class="tabbar">
          <div
            v-for="item in visibleTabs"
            :key="item.path"
            class="tab-item"
            :class="{ 'is-active': route.path === item.path }"
          >
            <button type="button" class="tab-link" @click="router.push(item.path)">
              {{ item.label }}
            </button>
            <button
              v-if="item.path !== '/'"
              type="button"
              class="tab-close"
              @click="handleCloseTab(item.path)"
            >
              ×
            </button>
          </div>
        </div>

        <div class="user-panel" ref="userPanelRef">
          <div
            class="user-profile-zone"
            @mouseenter="handleProfileEnter"
            @mouseleave="handleProfileLeave"
          >
            <button type="button" class="user-profile-trigger" @click="goToSettings">
              <span class="user-avatar">
                <img
                  v-if="state.user?.avatarUrl"
                  :src="state.user.avatarUrl"
                  :alt="displayName"
                  class="user-avatar-image"
                />
                <span v-else class="user-avatar-fallback">{{ avatarText }}</span>
              </span>

              <span class="user-summary">
                <strong>{{ displayName }}</strong>
                <span>账号设置</span>
              </span>
            </button>

            <div class="user-hover-card" :class="{ 'is-open': userHoverOpen }">
              <div class="user-hover-head">
                <span class="user-avatar user-avatar-large">
                  <img
                    v-if="state.user?.avatarUrl"
                    :src="state.user.avatarUrl"
                    :alt="displayName"
                    class="user-avatar-image"
                  />
                  <span v-else class="user-avatar-fallback">{{ avatarText }}</span>
                </span>

                <div class="user-hover-copy">
                  <strong>{{ displayName }}</strong>
                  <span>{{ state.user?.email || "未绑定账号" }}</span>
                </div>
              </div>

              <div class="user-hover-body">
                <div>
                  <span class="user-hover-label">昵称</span>
                  <p>{{ state.user?.nickname || "未设置昵称" }}</p>
                </div>
                <div>
                  <span class="user-hover-label">账号</span>
                  <p>{{ state.user?.email || "-" }}</p>
                </div>
              </div>
            </div>
          </div>

          <div class="user-menu-zone">
            <button
              type="button"
              class="user-menu-trigger"
              :class="{ 'is-open': userMenuOpen }"
              @click.stop="toggleUserMenu"
            >
              ▾
            </button>

            <div v-if="userMenuOpen" class="user-menu-dropdown">
              <button type="button" class="user-menu-item" @click="handleLogout">退出登录</button>
            </div>
          </div>
        </div>
      </header>

      <section class="page-body">
        <slot />
      </section>
    </main>
  </div>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { RouterLink, useRoute, useRouter } from "vue-router";
import brandLogo from "../assets/brand-logo.svg";
import { useInterviewState } from "../state/interview";

const route = useRoute();
const router = useRouter();
const { state, openTab, closeTab, logout } = useInterviewState();

const navTree = [
  { label: "主页", to: "/" },
  {
    label: "面试",
    children: [
      { label: "面试配置", to: "/setup" },
      { label: "模拟面试", to: "/interview" }
    ]
  },
  {
    label: "复盘",
    children: [
      { label: "历史记录", to: "/history" },
      { label: "复盘报告", to: "/report" }
    ]
  },
  { label: "模型配置", to: "/ai-config" }
];

const routeLabelMap = {
  "/": "主页",
  "/setup": "面试配置",
  "/interview": "模拟面试",
  "/history": "历史记录",
  "/report": "复盘报告",
  "/ai-config": "模型配置",
  "/settings": "个人设置"
};

const expandedGroups = ref([]);
const userHoverOpen = ref(false);
const userMenuOpen = ref(false);
const userPanelRef = ref(null);

const visibleTabs = computed(() => {
  const uniquePaths = state.openTabs.filter((path, index, items) => items.indexOf(path) === index);
  return uniquePaths.map((path) => ({
    path,
    label: routeLabelMap[path] || path
  }));
});

const displayName = computed(() => state.user?.nickname || state.user?.email || "当前用户");

const avatarText = computed(() => {
  const source = String(displayName.value || "").trim();
  if (!source) {
    return "AI";
  }
  return source.slice(0, 2).toUpperCase();
});

function isExpanded(label) {
  return expandedGroups.value.includes(label);
}

function isGroupCurrent(group) {
  if (!group.children?.length) {
    return route.path === group.to;
  }

  return group.children.some((item) => item.to === route.path);
}

function toggleGroup(label) {
  if (isExpanded(label)) {
    expandedGroups.value = expandedGroups.value.filter((item) => item !== label);
    return;
  }

  expandedGroups.value = [...expandedGroups.value, label];
}

function hasActiveInterviewSession() {
  return (
    !!state.sessionId &&
    !!state.sessionDetail?.session &&
    state.sessionDetail.session.id === state.sessionId &&
    state.sessionDetail.session.status !== "completed"
  );
}

function handleNavClick(item) {
  if (item.to === "/interview" && !hasActiveInterviewSession()) {
    window.alert("请先进行面试配置");
    openTab("/setup");
    router.push("/setup");
    return;
  }

  if (item.to === "/report" && !state.selectedReportSessionId) {
    window.alert("请先选择复盘记录");
    openTab("/history");
    router.push("/history");
    return;
  }

  openTab(item.to);
  router.push(item.to);
}

function handleCloseTab(path) {
  const currentTabs = visibleTabs.value;
  const currentIndex = currentTabs.findIndex((item) => item.path === path);
  const fallbackPath =
    currentTabs[currentIndex - 1]?.path || currentTabs[currentIndex + 1]?.path || "/";

  closeTab(path);

  if (route.path === path) {
    router.push(fallbackPath);
  }
}

function goToSettings() {
  userHoverOpen.value = false;
  if (document.activeElement instanceof HTMLElement) {
    document.activeElement.blur();
  }
  openTab("/settings");
  router.push("/settings");
}

function handleProfileEnter() {
  userHoverOpen.value = true;
}

function handleProfileLeave() {
  userHoverOpen.value = false;
}

function toggleUserMenu() {
  userMenuOpen.value = !userMenuOpen.value;
}

function handleOutsidePointer(event) {
  if (!userPanelRef.value?.contains(event.target)) {
    userMenuOpen.value = false;
  }
}

async function handleLogout() {
  userMenuOpen.value = false;
  await logout();
  router.replace("/login");
}

watch(
  () => route.path,
  (path) => {
    userHoverOpen.value = false;
    userMenuOpen.value = false;
    openTab(path);
  },
  { immediate: true }
);

onMounted(() => {
  document.addEventListener("mousedown", handleOutsidePointer);
});

onBeforeUnmount(() => {
  document.removeEventListener("mousedown", handleOutsidePointer);
});
</script>
