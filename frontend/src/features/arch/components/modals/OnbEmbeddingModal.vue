<script setup lang="ts">
const steps = [
  'Collect rc_tables texts (table_name + description)',
  'Collect rc_columns texts (table.column + type + description + samples + synonyms)',
  'Collect rc_terms texts (term + definition + synonyms)',
  'Batch embed: 100 texts/batch via Doubao Embedding API (1536 dimensions)',
  'UPSERT INTO rc_embeddings — VECTOR(2048) + HNSW COSINE index auto-built',
]
</script>

<template>
  <div>
    <h3 class="text-base font-bold text-gray-900 mb-1">Embedding</h3>
    <p class="text-sm text-gray-500 mb-4">EmbeddingWorker — GenerateAndSaveEmbeddings()</p>
    <div class="space-y-2">
      <div v-for="s in steps" :key="s" class="flex items-start gap-2 text-sm text-gray-600">
        <span class="text-emerald-500 flex-shrink-0">→</span>
        <span>{{ s }}</span>
      </div>
    </div>
  </div>
</template>
