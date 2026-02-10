<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import { useDemoStore } from '@/stores/demo'
import ScenarioTabs from '@/components/demo/ScenarioTabs.vue'
import ContextComparison from '@/components/demo/ContextComparison.vue'
import SelfMaintainDemo from '@/components/demo/SelfMaintainDemo.vue'
import BenchmarkDemo from '@/components/demo/BenchmarkDemo.vue'
import LakebaseArchitecture from '@/components/demo/LakebaseArchitecture.vue'

const store = useDemoStore()

onMounted(() => {
  store.loadComparisonCases()
  store.loadMaintenanceLogs()
})
</script>

<template>
  <div class="min-h-screen bg-slate-50">
    <!-- Page Title -->
    <div class="bg-white border-b border-gray-200 px-6 py-4">
      <div class="max-w-7xl mx-auto flex items-center justify-between">
        <div class="flex items-center gap-3">
          <RouterLink 
            to="/"
            class="group w-9 h-9 rounded-lg bg-gray-100 flex items-center justify-center hover:bg-primary-50 transition-colors"
          >
            <div class="i-lucide-arrow-left text-lg text-gray-500 group-hover:text-primary-600 transition-colors" />
          </RouterLink>
          <div>
            <h1 class="text-lg font-semibold text-gray-900">Live Demo</h1>
            <p class="text-xs text-gray-400 mt-0.5">Interactive demonstration of LUCID capabilities</p>
          </div>
        </div>
        <div class="flex items-center gap-2">
          <div class="flex items-center gap-2 px-3 py-1.5 bg-gray-50 rounded-lg border border-gray-200 text-sm">
            <span class="i-lucide-box text-primary-600" />
            <span class="font-medium text-gray-600">MariaDB 12</span>
          </div>
          <div class="flex items-center gap-2 px-3 py-1.5 bg-gray-50 rounded-lg border border-gray-200 text-sm">
            <span class="i-lucide-brain text-purple-600" />
            <span class="font-medium text-gray-600">GPT-4o</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Main Content -->
    <main class="max-w-7xl mx-auto px-6 py-6">
      <!-- Scenario Tabs -->
      <ScenarioTabs 
        v-model="store.currentScenario" 
        class="mb-6"
      />

      <!-- Self-Maintain Demo (full width — has its own layout) -->
      <SelfMaintainDemo 
        v-if="store.currentScenario === 'selfmaintain'"
      />

      <!-- Other demos: split layout with architecture sidebar -->
      <div v-else class="grid grid-cols-12 gap-6">
        <!-- Left: Main Demo Area -->
        <div class="col-span-8">
          <!-- Context Comparison (Core) -->
          <ContextComparison 
            v-if="store.currentScenario === 'comparison'"
          />

          <!-- Benchmark Demo -->
          <BenchmarkDemo 
            v-else
          />
        </div>

        <!-- Right: Architecture & Info -->
        <div class="col-span-4">
          <LakebaseArchitecture />
        </div>
      </div>
    </main>
  </div>
</template>
