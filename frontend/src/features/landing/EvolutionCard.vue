<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import type { Database } from '@/types'

const props = defineProps<{
  database: Database
}>()

const router = useRouter()

const isConnected = computed(() => props.database.status === 'connected')

function handleEnter() {
  if (isConnected.value) {
    router.push(`/workspace/${props.database.id}`)
  }
}

const stages = [
  { step: 1, label: 'Add Table', desc: 'New product_reviews table' },
  { step: 2, label: 'Add Column', desc: 'delivery_status to orders' },
  { step: 3, label: 'Modify Column', desc: 'Expand price precision' },
  { step: 4, label: 'Rename Column', desc: 'qty → quantity' },
  { step: 5, label: 'Drop Column', desc: 'Remove deprecated field' },
]
</script>

<template>
  <div
    class="evo-card group relative overflow-hidden rounded-2xl cursor-pointer bg-white/80 backdrop-blur-sm border border-white/60 shadow-lg shadow-gray-200/40 hover:shadow-xl hover:shadow-teal-200/40 hover:-translate-y-1 hover:bg-white/95 transition-all duration-300"
    :class="{ 'opacity-60 grayscale': !isConnected }"
    @click="handleEnter"
  >
    <!-- Top accent bar -->
    <div class="h-1.5 w-full bg-gradient-to-r from-teal-400 via-cyan-500 to-blue-500" />

    <div class="p-5 flex flex-col h-full">
      <!-- Header -->
      <div class="flex items-center gap-3 mb-4">
        <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-teal-500 to-cyan-600 flex items-center justify-center shadow-lg shadow-teal-500/30">
          <div class="i-atlas-refresh-cw text-2xl text-white" />
        </div>
        <div class="flex-1 min-w-0">
          <h3 class="font-bold text-base text-gray-800 leading-tight group-hover:text-teal-600 transition-colors">
            Evolution Demo
          </h3>
          <p class="text-xs text-gray-500 mt-0.5">Agent Self-Maintenance Pipeline</p>
        </div>
        <div
          class="w-3 h-3 rounded-full flex-shrink-0 ring-4"
          :class="isConnected
            ? 'bg-green-500 ring-green-500/20 animate-pulse'
            : 'bg-yellow-500 ring-yellow-500/20'"
        />
      </div>

      <!-- DDL Evolution stages -->
      <div class="flex-1 mb-4">
        <div class="flex items-center gap-1 mb-3">
          <span class="px-2.5 py-1 text-[10px] font-bold rounded-lg bg-teal-50 text-teal-700 border border-teal-200 uppercase tracking-wide">
            5-Stage DDL Evolution
          </span>
          <span class="text-xs text-gray-400 ml-1">
            {{ database.tableCount }} tables · {{ database.contextCount }} ctx
          </span>
        </div>

        <div class="space-y-1">
          <div
            v-for="s in stages"
            :key="s.step"
            class="flex items-center gap-2.5 py-1.5 px-2 rounded-lg hover:bg-teal-50/60 transition-colors"
          >
            <!-- Step number -->
            <div class="w-5 h-5 rounded-md bg-gradient-to-br from-teal-100 to-cyan-100 text-teal-700 flex items-center justify-center text-[10px] font-extrabold flex-shrink-0">
              {{ s.step }}
            </div>
            <!-- Label + desc -->
            <div class="flex-1 min-w-0 flex items-center gap-2">
              <span class="text-xs font-bold text-gray-700">{{ s.label }}</span>
              <span class="text-[10px] text-gray-400 truncate">{{ s.desc }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Footer -->
      <div class="flex items-center justify-between pt-3 border-t border-gray-100">
        <span class="px-2.5 py-1 text-xs font-semibold rounded-lg bg-gray-100 text-gray-600">
          lakebase
        </span>
        <div class="flex items-center gap-1.5 text-xs font-semibold text-teal-600 opacity-0 group-hover:opacity-100 transition-all duration-300">
          Open <div class="i-atlas-arrow-right text-sm" />
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.evo-card {
  min-height: 280px;
}
</style>
