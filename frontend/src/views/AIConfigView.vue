<template>
  <div class="page-grid">
    <section class="panel">
      <div class="panel-header">
        <div>
          <h3>模型配置</h3>
          <p>配置你自己的模型接口，面试出题、评分和复盘会优先使用当前账号下的配置。</p>
        </div>
      </div>

      <div v-if="state.error" class="message-banner message-error">
        {{ state.error }}
      </div>

      <div v-if="testMessage" class="message-banner" :class="{ 'message-error': !testOK }">
        {{ testMessage }}
      </div>

      <form class="config-grid" @submit.prevent="handleSave">
        <article class="config-card">
          <h4>基础信息</h4>
          <div class="field-list">
            <label class="field">
              <span>提供商</span>
              <select v-model="form.provider">
                <option value="deepseek">DeepSeek</option>
                <option value="openai">OpenAI</option>
              </select>
            </label>
            <label class="field">
              <span>是否启用</span>
              <select v-model="enabledValue">
                <option value="true">启用</option>
                <option value="false">停用</option>
              </select>
            </label>
          </div>
        </article>

        <article class="config-card">
          <h4>接口参数</h4>
          <div class="field-list">
            <label class="field">
              <span>API Key</span>
              <input
                v-model="form.apiKey"
                type="password"
                :placeholder="state.aiConfig?.apiKeyMasked || '输入新的 API Key'"
              />
            </label>
            <label class="field">
              <span>模型名称</span>
              <input v-model="form.model" type="text" />
            </label>
            <label class="field">
              <span>Base URL</span>
              <input v-model="form.baseUrl" type="text" />
            </label>
          </div>
        </article>
      </form>

      <div class="button-row">
        <button class="ghost-button" :disabled="state.loading.aiConfigTest" @click="handleTest">
          {{ state.loading.aiConfigTest ? "测试中..." : "测试连接" }}
        </button>
        <button class="primary-button" :disabled="state.loading.aiConfig" @click="handleSave">
          {{ state.loading.aiConfig ? "保存中..." : "保存配置" }}
        </button>
      </div>
    </section>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref, watch } from "vue";
import { useInterviewState } from "../state/interview";

const { state, loadAIConfig, saveAIConfig, testAIConfig } = useInterviewState();

const form = reactive({
  provider: "deepseek",
  apiKey: "",
  model: "deepseek-v4-flash",
  baseUrl: "https://api.deepseek.com",
  isEnabled: true
});

const testMessage = ref("");
const testOK = ref(false);

const enabledValue = computed({
  get() {
    return form.isEnabled ? "true" : "false";
  },
  set(value) {
    form.isEnabled = value === "true";
  }
});

watch(
  () => form.provider,
  (provider) => {
    if (provider === "deepseek") {
      if (!form.model || form.model === "gpt-5.5") {
        form.model = "deepseek-v4-flash";
      }
      if (!form.baseUrl || form.baseUrl.includes("openai.com")) {
        form.baseUrl = "https://api.deepseek.com";
      }
      return;
    }

    if (!form.model || form.model === "deepseek-v4-flash") {
      form.model = "gpt-5.5";
    }
    if (!form.baseUrl || form.baseUrl.includes("deepseek.com")) {
      form.baseUrl = "https://api.openai.com/v1/responses";
    }
  }
);

function syncForm(config) {
  if (!config) {
    return;
  }

  form.provider = config.provider || "deepseek";
  form.apiKey = "";
  form.model = config.model || form.model;
  form.baseUrl = config.baseUrl || form.baseUrl;
  form.isEnabled = Boolean(config.isEnabled);
}

async function handleLoad() {
  const config = await loadAIConfig();
  syncForm(config);
}

async function handleTest() {
  const result = await testAIConfig({
    provider: form.provider,
    apiKey: form.apiKey,
    model: form.model,
    baseUrl: form.baseUrl
  });

  testMessage.value = result.message;
  testOK.value = Boolean(result.ok);
}

async function handleSave() {
  const config = await saveAIConfig({
    provider: form.provider,
    apiKey: form.apiKey,
    model: form.model,
    baseUrl: form.baseUrl,
    isEnabled: form.isEnabled
  });

  syncForm(config);
  testMessage.value = "配置已保存";
  testOK.value = true;
}

onMounted(() => {
  handleLoad();
});
</script>
