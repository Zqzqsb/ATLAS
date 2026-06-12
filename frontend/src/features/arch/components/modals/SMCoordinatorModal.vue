<template>
  <div>
    <h3 class="text-base font-bold text-gray-900 mb-1">Coordinator</h3>
    <p class="text-sm text-gray-500 mb-3">AgentService orchestrates the self-maintenance loop. Receives delta from DDLDetector, dispatches ContextMaintainer for LLM refresh, and ChangeLogger for audit trail.</p>
    <div class="text-sm text-gray-500 font-semibold mb-1">Loop</div>
    <p class="text-sm text-gray-500">Detect → Expire → Refresh Context (LLM) → Re-embed → Log → Verify (next detection pass confirms no delta → equilibrium).</p>
  </div>
</template>
