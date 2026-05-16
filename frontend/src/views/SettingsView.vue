<template>
  <div class="page-grid">
    <section class="panel settings-panel">
      <div class="panel-header">
        <div>
          <h3>个人设置</h3>
          <p>这里可以维护你的头像、昵称和基础账号信息。</p>
        </div>
      </div>

      <div v-if="state.error" class="message-banner message-error">
        {{ state.error }}
      </div>

      <div v-else-if="successMessage" class="message-banner message-success">
        {{ successMessage }}
      </div>

      <div class="settings-layout">
        <aside class="settings-preview-card">
          <button
            type="button"
            class="settings-avatar-button"
            :disabled="state.loading.profile"
            @click="openFilePicker"
          >
            <div class="settings-avatar">
              <img
                v-if="form.avatarUrl"
                :src="form.avatarUrl"
                :alt="previewName"
                class="settings-avatar-image"
              />
              <span v-else>{{ avatarText }}</span>
            </div>
            <span class="settings-avatar-tip">点击更换头像</span>
          </button>

          <input
            ref="fileInputRef"
            class="upload-input"
            type="file"
            accept="image/png,image/jpeg,image/webp,image/gif"
            @change="handleAvatarChange"
          />

          <div class="settings-preview-copy">
            <strong>{{ previewName }}</strong>
            <span>{{ state.user?.email || "未绑定邮箱" }}</span>
          </div>

          <p class="helper-text">
            支持 PNG、JPG、WEBP、GIF 图片。上传后会直接作为你的头像保存并显示在右上角用户区。
          </p>

          <button
            v-if="form.avatarUrl"
            type="button"
            class="ghost-button settings-clear-button"
            :disabled="state.loading.profile"
            @click="clearAvatar"
          >
            移除头像
          </button>
        </aside>

        <form class="field-list" @submit.prevent="handleSubmit">
          <label class="field">
            <span>昵称</span>
            <input v-model="form.nickname" type="text" maxlength="24" placeholder="请输入昵称" />
          </label>

          <label class="field">
            <span>账号邮箱</span>
            <input :value="state.user?.email || ''" type="email" readonly />
          </label>

          <div class="button-row">
            <button class="primary-button" :disabled="state.loading.profile">
              {{ state.loading.profile ? "保存中..." : "保存设置" }}
            </button>
            <button
              type="button"
              class="ghost-button"
              :disabled="state.loading.profile"
              @click="resetForm"
            >
              恢复当前资料
            </button>
          </div>
        </form>
      </div>
    </section>
  </div>
</template>

<script setup>
import { computed, reactive, ref, watch } from "vue";
import { useInterviewState } from "../state/interview";

const { state, updateCurrentUser, setError } = useInterviewState();

const successMessage = ref("");
const fileInputRef = ref(null);

const form = reactive({
  nickname: "",
  avatarUrl: ""
});

const previewName = computed(() => form.nickname.trim() || state.user?.nickname || "当前用户");
const avatarText = computed(() => {
  const source = String(previewName.value || "").trim();
  return source ? source.slice(0, 2).toUpperCase() : "AI";
});

function syncForm() {
  form.nickname = state.user?.nickname || "";
  form.avatarUrl = state.user?.avatarUrl || "";
}

function resetForm() {
  successMessage.value = "";
  setError("");
  syncForm();
}

function openFilePicker() {
  fileInputRef.value?.click();
}

function clearAvatar() {
  form.avatarUrl = "";
  successMessage.value = "";
  setError("");
}

function readFileAsDataURL(file) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => resolve(String(reader.result || ""));
    reader.onerror = () => reject(new Error("头像读取失败，请重试"));
    reader.readAsDataURL(file);
  });
}

async function handleAvatarChange(event) {
  const [file] = event.target.files || [];
  event.target.value = "";

  if (!file) {
    return;
  }

  if (file.size > 2 * 1024 * 1024) {
    setError("头像图片不能超过 2MB");
    return;
  }

  successMessage.value = "";
  setError("");

  try {
    form.avatarUrl = await readFileAsDataURL(file);
  } catch (error) {
    setError(error.message);
  }
}

async function handleSubmit() {
  successMessage.value = "";
  setError("");

  try {
    await updateCurrentUser({
      nickname: form.nickname,
      avatarUrl: form.avatarUrl
    });
    successMessage.value = "个人资料已更新";
  } catch {
    // error already handled by store
  }
}

watch(
  () => state.user,
  () => {
    syncForm();
  },
  { immediate: true }
);
</script>
