<template>
  <div class="admin-grid two-column">
    <section class="admin-card">
      <div class="admin-card-head">
        <div>
          <h3>面试记录</h3>
          <p>查看全局会话列表、题目、答案和复盘报告。</p>
        </div>
      </div>

      <div class="toolbar">
        <input v-model="filters.keyword" type="text" placeholder="搜索用户、岗位、公司或会话名" />
        <select v-model="filters.status">
          <option value="">全部状态</option>
          <option value="draft">草稿</option>
          <option value="in_progress">进行中</option>
          <option value="completed">已完成</option>
        </select>
        <button type="button" class="admin-button ghost" @click="handleSearch">查询</button>
      </div>

      <div v-if="state.loading.sessions" class="admin-empty">正在加载面试记录...</div>
      <div v-else class="table-list">
        <button
          v-for="item in state.sessions.items"
          :key="item.sessionId"
          type="button"
          class="table-row button-row"
          :class="{ 'is-selected': state.selectedSession?.session?.id === item.sessionId }"
          @click="handleSelect(item.sessionId)"
        >
          <span>{{ item.sessionName }}</span>
          <span>{{ item.userEmail }}</span>
          <span>{{ item.jobTitle }}</span>
          <span>{{ item.status }}</span>
        </button>
      </div>
    </section>

    <section class="admin-card">
      <div class="admin-card-head">
        <div>
          <h3>会话详情</h3>
          <p>展示题目列表、作答历史与复盘结论。</p>
        </div>
      </div>

      <div v-if="state.loading.sessionDetail" class="admin-empty">正在加载详情...</div>
      <div v-else-if="!state.selectedSession" class="admin-empty">请选择左侧会话。</div>
      <div v-else class="detail-stack">
        <div class="detail-grid">
          <article class="detail-card">
            <span>所属用户</span>
            <strong>{{ state.selectedSession.user.email }}</strong>
          </article>
          <article class="detail-card">
            <span>岗位</span>
            <strong>{{ state.selectedSession.session.jobTitle }}</strong>
          </article>
          <article class="detail-card">
            <span>轮次</span>
            <strong>{{ state.selectedSession.session.roundType }}</strong>
          </article>
          <article class="detail-card">
            <span>状态</span>
            <strong>{{ state.selectedSession.session.status }}</strong>
          </article>
        </div>

        <div class="detail-card">
          <h4>题目列表</h4>
          <ul class="plain-list">
            <li v-for="item in state.selectedSession.questions" :key="item.id">
              {{ item.questionNo }}. {{ item.prompt }}
            </li>
          </ul>
        </div>

        <div class="detail-card">
          <h4>答案历史</h4>
          <ul class="plain-list" v-if="state.sessionAnswers.length">
            <li v-for="item in state.sessionAnswers" :key="item.answer.id">
              <strong>Q{{ item.questionNo }}：</strong>{{ item.answer.answerText || "未作答" }}
            </li>
          </ul>
          <p v-else>暂无答案。</p>
        </div>

        <div class="detail-card" v-if="state.sessionReport">
          <h4>复盘报告</h4>
          <p>综合得分：{{ Number(state.sessionReport.overallScore || 0).toFixed(1) }}</p>
          <p>{{ state.sessionReport.performanceSummary }}</p>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { onMounted, reactive } from "vue";
import { useAdminState } from "../state/admin";

const { state, loadSessions, loadSessionDetail } = useAdminState();
const filters = reactive({
  keyword: "",
  status: ""
});

async function handleSearch() {
  await loadSessions({ page: 1, keyword: filters.keyword, status: filters.status });
}

async function handleSelect(sessionId) {
  await loadSessionDetail(sessionId);
}

onMounted(() => {
  handleSearch();
});
</script>
