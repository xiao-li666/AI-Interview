<template>
  <AdminShell v-if="showShell">
    <RouterView />
  </AdminShell>
  <RouterView v-else />
</template>

<script setup>
import { computed } from "vue";
import { RouterView, useRoute } from "vue-router";
import AdminShell from "./components/AdminShell.vue";

const route = useRoute();
const showShell = computed(() => !route.meta.public);
</script>
