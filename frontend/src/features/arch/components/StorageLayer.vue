<script setup lang="ts">
defineEmits<{ click: [key: string] }>()

const pills = [
  { key: 'store-schema', label: 'Schema Metadata', tag: 'rc_tables / rc_columns / rc_relations',
    desc: 'Physical schema from INFORMATION_SCHEMA. Tables, columns (type, nullable, PK, FK), and foreign key relationships.' },
  { key: 'store-ctx', label: 'Rich Context', tag: 'rc_business_context · 11 types',
    desc: 'LLM-generated semantic context: description, example, constraint, synonym, value_mapping, business_rule, calculation, semantic, enum_meaning, join_hint, data_quality.' },
  { key: 'store-emb', label: 'Vector Embeddings', tag: 'rc_embeddings · VECTOR(2048) · HNSW', highlight: true,
    desc: 'Doubao embedding 1536d. COSINE HNSW index for sub-ms similarity search. Soft-delete (is_deleted) and staleness tracking (is_stale) for evolution.' },
  { key: 'store-log', label: 'Change Log', tag: 'rc_change_log',
    desc: 'Audit trail: old_value → new_value, change_type, entity_type, reason, changed_by. Enables full traceability for self-maintenance.' },
  { key: 'store-hnsw', label: 'HNSW Index', tag: 'Native MariaDB 12',
    desc: 'Native VECTOR column type with HNSW (Hierarchical Navigable Small World). Cosine distance. Built into MariaDB 12 — no external vector database needed.' },
]
</script>

<template>
  <div class="rounded-xl bg-slate-100/80 border border-slate-200 px-6 py-5">
    <div class="text-sm font-bold text-slate-600 mb-4 text-center">Unified Storage — MariaDB 12 (VECTOR + HNSW)</div>
    <div class="grid grid-cols-5 gap-3">
      <button
        v-for="pill in pills" :key="pill.key"
        class="rounded-lg bg-white border p-3 text-left transition-all hover:shadow-md cursor-pointer"
        :class="pill.highlight ? 'border-cyan-300/80 shadow-md shadow-cyan-100/30' : 'border-gray-200'"
        @click="$emit('click', pill.key)"
      >
        <div class="text-xs font-semibold text-gray-800 mb-0.5">{{ pill.label }}</div>
        <div class="text-2xs text-gray-400">{{ pill.tag }}</div>
      </button>
    </div>
  </div>
</template>

<style scoped>.text-2xs{font-size:.7rem}</style>
