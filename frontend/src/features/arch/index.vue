<script setup lang="ts">
import { ref, shallowRef, type Component } from 'vue'
import { NModal } from 'naive-ui'
import UILayer from './components/UILayer.vue'
import OnboardingLane from './components/OnboardingLane.vue'
import InferenceLane from './components/InferenceLane.vue'
import SelfMaintenanceLane from './components/SelfMaintenanceLane.vue'
import StorageLayer from './components/StorageLayer.vue'

// --- Modal lazy imports ---
const OnbSchemaSyncModal  = () => import('./components/modals/OnbSchemaSyncModal.vue')
const OnbContextGenModal  = () => import('./components/modals/OnbContextGenModal.vue')
const OnbEmbeddingModal   = () => import('./components/modals/OnbEmbeddingModal.vue')
const InfLinkingModal     = () => import('./components/modals/InfLinkingModal.vue')
const InfSQLGenModal      = () => import('./components/modals/InfSQLGenModal.vue')
const InfVerifyModal      = () => import('./components/modals/InfVerifyModal.vue')
const SMDetectorModal     = () => import('./components/modals/SMDetectorModal.vue')
const SMCoordinatorModal   = () => import('./components/modals/SMCoordinatorModal.vue')
const SMRefresherModal    = () => import('./components/modals/SMRefresherModal.vue')
const SMLoggerModal       = () => import('./components/modals/SMLoggerModal.vue')
const StorePillModal      = () => import('./components/modals/StorePillModal.vue')

const modalMap: Record<string, () => Promise<Component>> = {
  'onb-sync':     OnbSchemaSyncModal,
  'onb-ctx':      OnbContextGenModal,
  'onb-emb':      OnbEmbeddingModal,
  'inf-link':     InfLinkingModal,
  'inf-sql':      InfSQLGenModal,
  'inf-verify':   InfVerifyModal,
  'sm-detect':    SMDetectorModal,
  'sm-coord':     SMCoordinatorModal,
  'sm-refresh':   SMRefresherModal,
  'sm-log':       SMLoggerModal,
  'store-schema': StorePillModal,
  'store-ctx':    StorePillModal,
  'store-emb':    StorePillModal,
  'store-log':    StorePillModal,
  'store-hnsw':   StorePillModal,
}

// Storage pill metadata for modal title
const storePills: Record<string, { label: string; tag: string; desc: string }> = {
  'store-schema': { label: 'Schema Metadata', tag: 'rc_tables / rc_columns / rc_relations',
    desc: 'Physical schema from INFORMATION_SCHEMA. Tables, columns (type, nullable, PK, FK), and foreign key relationships.' },
  'store-ctx': { label: 'Rich Context', tag: 'rc_business_context · 11 types',
    desc: 'LLM-generated semantic context: description, example, constraint, synonym, value_mapping, business_rule, calculation, semantic, enum_meaning, join_hint, data_quality.' },
  'store-emb': { label: 'Vector Embeddings', tag: 'rc_embeddings · VECTOR(2048) · HNSW',
    desc: 'Doubao embedding 1536d. COSINE HNSW index for sub-ms similarity search. Soft-delete (is_deleted) and staleness tracking (is_stale) for evolution.' },
  'store-log': { label: 'Change Log', tag: 'rc_change_log',
    desc: 'Audit trail: old_value → new_value, change_type, entity_type, reason, changed_by. Enables full traceability for self-maintenance.' },
  'store-hnsw': { label: 'HNSW Index', tag: 'Native MariaDB 12',
    desc: 'Native VECTOR column type with HNSW (Hierarchical Navigable Small World). Cosine distance. Built into MariaDB 12 — no external vector database needed.' },
}

const modalTitle = ref('')
const modalComponent = shallowRef<Component | null>(null)
const storePillProps = ref<{ label: string; tag: string; desc: string } | null>(null)

function openModal(key: string) {
  const loader = modalMap[key]
  if (!loader) return

  if (key.startsWith('store-')) {
    storePillProps.value = storePills[key] || null
  } else {
    storePillProps.value = null
  }

  loader().then(mod => {
    modalComponent.value = (mod as any).default
  })
}
function closeModal() {
  modalComponent.value = null
  storePillProps.value = null
}
</script>

<template>
  <div class="min-h-screen bg-white">
    <!-- Header -->
    <div class="bg-white border-b border-gray-200/80 px-8 py-4 sticky top-14 z-40">
      <div class="max-w-7xl mx-auto flex items-center gap-3">
        <div class="w-8 h-8 rounded-lg bg-primary-50 flex items-center justify-center">
          <div class="i-lucide-boxes text-primary-600 text-base" />
        </div>
        <div>
          <h2 class="text-lg font-bold text-gray-900 m-0">System Architecture</h2>
          <p class="text-xs text-gray-400">Click any component to see internal workflow</p>
        </div>
      </div>
    </div>

    <div class="max-w-7xl mx-auto px-8 py-6 space-y-5">
      <!-- Layer 1 -->
      <UILayer />
      <div class="flex justify-center gap-6">
        <div v-for="i in 3" :key="i" class="flex flex-col items-center gap-0.5">
          <div class="w-px h-4 bg-gray-300" /><div class="i-lucide-chevron-down text-gray-400 text-xs" />
        </div>
      </div>

      <!-- Layer 2 -->
      <div class="grid grid-cols-12 gap-4">
        <div class="col-span-3"><OnboardingLane @click="openModal" /></div>
        <div class="col-span-6"><InferenceLane @click="openModal" /></div>
        <div class="col-span-3"><SelfMaintenanceLane @click="openModal" /></div>
      </div>

      <!-- Layer 3 -->
      <StorageLayer @click="openModal" />

      <!-- Legend -->
      <div class="flex items-center justify-center gap-6 text-xs text-gray-400 pt-2">
        <div class="flex items-center gap-1.5"><div class="w-3 h-3 rounded-full bg-emerald-500" /><span>Onboarding</span></div>
        <div class="flex items-center gap-1.5"><div class="w-3 h-3 rounded-full bg-blue-500" /><span>Inference</span></div>
        <div class="flex items-center gap-1.5"><div class="w-3 h-3 rounded-full bg-amber-500" /><span>Self-Maintenance</span></div>
      </div>
    </div>

    <!-- Modal -->
    <NModal
      :show="modalComponent !== null"
      :on-update:show="(v: boolean) => { if (!v) closeModal() }"
      preset="card"
      style="max-width:640px"
      title=""
      :bordered="false"
    >
      <template #header>
        <div class="w-full pr-6">{{ modalTitle }}</div>
      </template>
      <component :is="modalComponent" v-bind="storePillProps || {}" />
    </NModal>
  </div>
</template>
