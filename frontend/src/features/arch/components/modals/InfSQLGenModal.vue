<script setup lang="ts">
const tools = [
  { name: 'execute_sql', desc: 'Run SELECT / SHOW / DESCRIBE on the business database. Read-only safety enforced.' },
  { name: 'verify_sql', desc: 'Semantic verification — do query results match user intent? DryRun with EXPLAIN before execution.' },
]
</script>

<template>
  <div>
    <h3 class="text-base font-bold text-gray-900 mb-1">SQL Generator</h3>
    <p class="text-sm text-gray-500 mb-4">ReAct agent loop. Ghost budget: tell LLM 5 iterations, allow up to 15 actual. GroundedContext injected as prompt.</p>

    <div class="grid grid-cols-2 gap-2 mb-3">
      <div v-for="t in tools" :key="t.name" class="bg-gray-50 border rounded p-2.5">
        <div class="text-sm font-semibold text-blue-600">{{ t.name }}</div>
        <div class="text-xs text-gray-500">{{ t.desc }}</div>
      </div>
    </div>

    <div class="text-sm text-blue-500 flex items-center gap-1">
      <span>💭 Thought</span><span>→</span><span>⚡ Action</span><span>→</span><span>👁️ Observation</span><span>→ repeat (ghost budget 3×)</span>
    </div>
  </div>
</template>
