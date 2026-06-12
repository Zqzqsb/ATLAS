<script setup lang="ts">
const steps = [
  'Query all rc_tables/rc_columns where is_expired=true',
  'Build targeted prompt with old context + new schema changes',
  'LLM generates updated descriptions, synonyms, business rules',
  'UPDATE rc_tables/rc_columns with new context + bump version',
  'Mark is_stale=true on corresponding rc_embeddings rows',
  'EmbeddingWorker re-embeds the updated entities',
]
</script>

<template>
  <div>
    <h3 class="text-base font-bold text-gray-900 mb-1">Context Refresher</h3>
    <p class="text-sm text-gray-500 mb-4">ContextMaintainer — LLM incremental refresh. Only re-generates for tables/columns marked is_expired=true. Avoids full re-onboarding.</p>
    <div class="space-y-1.5">
      <div v-for="s in steps" :key="s" class="text-sm text-gray-600">{{ s }}</div>
    </div>
  </div>
</template>
