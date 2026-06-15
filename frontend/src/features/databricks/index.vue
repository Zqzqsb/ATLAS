<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import DbxOverview from './components/overview/DbxOverview.vue'
import DbxModuleDetail from './components/module/DbxModuleDetail.vue'
import { getDbxFlow } from './model/flows'
import type { ArchNode } from './model/architecture'

type Level = 'overview' | 'module'

const level = ref<Level>('overview')
const activeFlowId = ref<string | null>(null)
const origin = ref('50% 40%')

const stageRef = ref<HTMLElement>()
const activeFlow = computed(() => getDbxFlow(activeFlowId.value))

function drillInto(node: ArchNode, ev: MouseEvent) {
  if (!node.flow) return
  const stage = stageRef.value
  if (stage) {
    const rect = stage.getBoundingClientRect()
    const x = ((ev.clientX - rect.left) / rect.width) * 100
    const y = ((ev.clientY - rect.top) / rect.height) * 100
    origin.value = `${x.toFixed(1)}% ${y.toFixed(1)}%`
  }
  activeFlowId.value = node.flow
  level.value = 'module'
}

function back() {
  level.value = 'overview'
}

function onKey(e: KeyboardEvent) {
  if (e.key === 'Escape' && level.value === 'module') back()
}

onMounted(() => window.addEventListener('keydown', onKey))
onUnmounted(() => window.removeEventListener('keydown', onKey))
</script>

<template>
  <div class="dbx-page bg-gradient-to-br from-slate-100 via-rose-50/40 to-orange-50/30">
    <div class="bg-white/80 backdrop-blur border-b border-gray-200/80 px-8 py-3 flex-shrink-0">
      <div class="max-w-5xl mx-auto flex items-center gap-2 text-sm">
        <button
          class="flex items-center gap-1.5 font-semibold transition-colors"
          :class="level === 'overview' ? 'text-gray-900' : 'text-gray-400 hover:text-gray-700'"
          @click="back"
        >
          <div class="i-lucide-mountain text-base" />
          Databricks 全景架构
        </button>
        <template v-if="level === 'module' && activeFlow">
          <div class="i-lucide-chevron-right text-gray-300" />
          <span class="font-semibold text-gray-900">{{ activeFlow.label }}</span>
        </template>
      </div>
    </div>

    <div ref="stageRef" class="relative flex-1 overflow-hidden">
      <div
        class="absolute inset-0 overflow-y-auto transition-all duration-[420ms] ease-out"
        :style="{ transformOrigin: origin }"
        :class="level === 'module'
          ? 'opacity-0 scale-[1.12] pointer-events-none'
          : 'opacity-100 scale-100'"
      >
        <DbxOverview @select="drillInto" />
      </div>

      <div
        class="absolute inset-0 overflow-y-auto transition-all duration-[420ms] ease-out"
        :style="{ transformOrigin: origin }"
        :class="level === 'module'
          ? 'opacity-100 scale-100'
          : 'opacity-0 scale-[0.92] pointer-events-none'"
      >
        <DbxModuleDetail v-if="activeFlow" :flow="activeFlow" @back="back" />
      </div>
    </div>
  </div>
</template>

<style scoped>
.dbx-page {
  height: calc(100vh - 56px);
  height: calc(100dvh - 56px);
  display: flex;
  flex-direction: column;
}
</style>
