<script setup lang="ts">
const phases = [
  { phase: 'Phase 1 · Connect', detail: 'AdapterFactory.GetAdapter(connectionID)', desc: 'Establish MySQL connection to the business database.' },
  { phase: 'Phase 2 · Register', detail: 'GetOrCreateDatasource() → INSERT INTO rc_datasources', desc: 'Register in Lake-Base catalog with status=active.' },
  { phase: 'Phase 3 · Sync Schema', detail: 'dispatch → SchemaSyncWorker', desc: 'Worker reads INFORMATION_SCHEMA, UPSERTs rc_tables/rc_columns/rc_relations.' },
  { phase: 'Phase 4 · Load', detail: 'GetTables/Columns/RelationsByDatasource()', desc: 'Read synced schema to build prompt for the ReAct agent.' },
  { phase: 'Phase 5 · Explore', detail: 'ReActExploreWorker (×N)', desc: 'LLM agent: execute_sql on business DB, set_rich_context writes to rc_business_context. Budget: ComputeChunkBudget(), ghost budget 3×.' },
  { phase: 'Phase 6 · Embed', detail: 'dispatch → EmbeddingWorker', desc: 'Batch embed via Doubao (1536d), UPSERT rc_embeddings + HNSW index.' },
  { phase: 'Phase 7 · Complete', detail: 'SendEvent("complete") · rc_change_log', desc: 'Close SSE stream, write audit entry.' },
]
</script>

<template>
  <div>
    <h3 class="text-base font-bold text-gray-900 mb-1">Context Generator</h3>
    <p class="text-sm text-gray-500 mb-4">OnboardingCoordinator — 7-phase orchestration, dispatches Worker Agents</p>

    <div class="space-y-2.5 mb-4">
      <div v-for="p in phases" :key="p.phase" class="flex items-start gap-2">
        <span class="text-sm font-mono font-semibold flex-shrink-0 w-28" :class="p.phase === 'Phase 5 · Explore' ? 'text-purple-500' : 'text-gray-600'">{{ p.phase }}</span>
        <div>
          <div class="text-sm text-gray-800">{{ p.detail }}</div>
          <div class="text-xs text-gray-400">{{ p.desc }}</div>
        </div>
      </div>
    </div>

    <!-- ReAct internals -->
    <div class="rounded-lg bg-purple-50/60 border border-purple-100 p-3">
      <div class="text-sm font-semibold text-purple-600 mb-2">Phase 5 · ReActExploreWorker internals</div>
      <div class="grid grid-cols-2 gap-2 mb-2">
        <div class="bg-white rounded border p-2">
          <div class="text-sm font-semibold text-teal-700">🔧 execute_sql</div>
          <div class="text-xs text-gray-500">SELECT / SHOW / DESCRIBE on business DB (read-only safety enforced)</div>
        </div>
        <div class="bg-white rounded border p-2">
          <div class="text-sm font-semibold text-purple-700">✏️ set_rich_context</div>
          <div class="text-xs text-gray-500">JSON payload → LakebaseRCWriter → rc_business_context (11 types)</div>
        </div>
      </div>
      <div class="text-xs text-purple-500 flex items-center gap-1">
        <span>💭 Thought</span><span>→</span><span>⚡ Action</span><span>→</span><span>👁️ Observation</span><span>→ repeat (5–15 iters)</span>
      </div>
    </div>
  </div>
</template>
