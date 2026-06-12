<script setup lang="ts">
const signals = [
  'TABLE: entity_name from rc_embeddings',
  'COLUMN: source_table + source_column metadata',
  'CONTEXT: business_rules + domain knowledge from rc_business_context',
  'SQL_TEMPLATE: historical query patterns',
]
</script>

<template>
  <div>
    <h3 class="text-base font-bold text-gray-900 mb-1">Adaptive Schema Linking</h3>
    <p class="text-sm text-gray-500 mb-4">Two-stage grounding — Stage 1: HNSW vector coarse retrieval. Stage 2: LLM LinkingAgent fine selection. Strategy auto-detected by table count threshold (30).</p>

    <div class="grid grid-cols-2 gap-3 mb-4">
      <div class="rounded-lg bg-blue-50/50 border border-blue-100 p-3">
        <div class="text-sm font-semibold text-blue-600 mb-1">≤ 30 tables — SmallScale</div>
        <p class="text-xs text-gray-500">Full schema directly to LinkingAgent.LinkDirect() — single LLM call selects relevant tables + columns + query-specific hints.</p>
      </div>
      <div class="rounded-lg bg-purple-50/50 border border-purple-100 p-3">
        <div class="text-sm font-semibold text-purple-600 mb-1">&gt; 30 tables — LargeScale</div>
        <p class="text-xs text-gray-500">4-way parallel HNSW search → merge candidate tables → LinkingAgent.LinkDirect() selects from candidates. ReAct mode available.</p>
      </div>
    </div>

    <!-- Decision diamond -->
    <div class="flex items-center justify-center mb-4">
      <div class="w-10 h-10 bg-blue-500 rounded-sm rotate-45 flex items-center justify-center">
        <span class="text-white text-xs font-bold -rotate-45">?</span>
      </div>
    </div>

    <div class="grid grid-cols-2 gap-3 mb-4">
      <div class="rounded border border-blue-200 bg-blue-50/30 p-2.5 text-center">
        <div class="text-sm font-semibold text-blue-600">Direct Path</div>
        <div class="text-xs text-blue-400">LinkAgent.LinkDirect()</div>
      </div>
      <div class="rounded border border-purple-200 bg-purple-50/30 p-2.5 text-center">
        <div class="text-sm font-semibold text-purple-600">Two-Stage Path</div>
        <div class="text-xs text-purple-400">4× HNSW → Merge → LLM Rerank</div>
      </div>
    </div>

    <div class="text-sm font-semibold text-gray-600 mb-1">4 Signal Types (parallel HNSW queries)</div>
    <div v-for="s in signals" :key="s" class="text-xs text-gray-500">• {{ s }}</div>
  </div>
</template>
