<template>
  <div class="page-grid two-column-layout">
    <section class="panel">
      <div class="panel-header">
        <h3>面试配置</h3>
        <p>这里会同时创建面试计划和真实面试会话，创建成功后直接进入答题页。</p>
      </div>

      <div v-if="state.error" class="message-banner message-error">
        {{ state.error }}
      </div>

      <form class="config-grid" @submit.prevent="handleSubmit">
        <article class="config-card">
          <h4>岗位信息</h4>
          <div class="field-list">
            <label class="field">
              <span>用户 ID</span>
              <input v-model.number="form.userId" type="number" min="1" />
            </label>
            <label class="field">
              <span>岗位名称</span>
              <input v-model="form.jobTitle" type="text" />
            </label>
            <label class="field">
              <span>岗位类别</span>
              <input v-model="form.jobCategory" type="text" />
            </label>
            <label class="field">
              <span>岗位级别</span>
              <input v-model="form.levelCode" type="text" />
            </label>
          </div>
        </article>

        <article class="config-card">
          <h4>面试设置</h4>
          <div class="field-list">
            <label class="field">
              <span>会话名称</span>
              <input v-model="form.sessionName" type="text" />
            </label>
            <label class="field">
              <span>面试轮次</span>
              <select v-model="form.roundType">
                <option v-for="item in roundTypeOptions" :key="item.value" :value="item.value">
                  {{ item.label }}
                </option>
              </select>
            </label>
            <label class="field">
              <span>题目类型</span>
              <select v-model="form.interviewType">
                <option v-for="item in roundTypeOptions" :key="item.value" :value="item.value">
                  {{ item.label }}
                </option>
              </select>
            </label>
            <label class="field">
              <span>面试模式</span>
              <select v-model="form.interviewMode">
                <option
                  v-for="item in interviewModeOptions"
                  :key="item.value"
                  :value="item.value"
                >
                  {{ item.label }}
                </option>
              </select>
            </label>
          </div>
        </article>

        <article class="config-card">
          <h4>难度与风格</h4>
          <div class="field-list">
            <label class="field">
              <span>难度等级</span>
              <select v-model="form.difficultyLevel">
                <option v-for="item in difficultyOptions" :key="item.value" :value="item.value">
                  {{ item.label }}
                </option>
              </select>
            </label>
            <label class="field">
              <span>面试官风格</span>
              <select v-model="form.interviewerStyle">
                <option
                  v-for="item in interviewerStyleOptions"
                  :key="item.value"
                  :value="item.value"
                >
                  {{ item.label }}
                </option>
              </select>
            </label>
            <label class="field">
              <span>题目数量</span>
              <input v-model.number="form.questionCount" type="number" min="1" max="5" />
            </label>
          </div>
        </article>
      </form>
    </section>

    <section class="panel slim-panel">
      <div class="panel-header">
        <h3>配置预览</h3>
      </div>

      <div class="summary-list">
        <div class="summary-row">
          <span>岗位名称</span>
          <strong>{{ form.jobTitle }}</strong>
        </div>
        <div class="summary-row">
          <span>面试轮次</span>
          <strong>{{ toRoundTypeLabel(form.roundType) }}</strong>
        </div>
        <div class="summary-row">
          <span>难度等级</span>
          <strong>{{ difficultyText }}</strong>
        </div>
        <div class="summary-row">
          <span>题目数量</span>
          <strong>{{ form.questionCount }}</strong>
        </div>
      </div>

      <button class="primary-button button-block" :disabled="state.loading.setup" @click="handleSubmit">
        {{ state.loading.setup ? "创建中..." : "创建并开始面试" }}
      </button>
    </section>
  </div>
</template>

<script setup>
import { computed, reactive, watch } from "vue";
import { useRouter } from "vue-router";
import {
  difficultyOptions,
  interviewModeOptions,
  interviewerStyleOptions,
  roundTypeOptions,
  toRoundTypeLabel
} from "../data/mock";
import { useInterviewState } from "../state/interview";

const router = useRouter();
const { state, updateConfig, createInterviewFlow, setError } = useInterviewState();

const form = reactive({
  ...state.config
});

const difficultyText = computed(
  () => difficultyOptions.find((item) => item.value === form.difficultyLevel)?.label || "--"
);

watch(
  form,
  (value) => {
    updateConfig(value);
  },
  { deep: true }
);

async function handleSubmit() {
  setError("");

  try {
    const detail = await createInterviewFlow();
    router.push({
      path: "/interview",
      query: {
        sessionId: String(detail.session.id)
      }
    });
  } catch {
    // state.error already set in the store
  }
}
</script>
