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
  <div class="min-h-screen bg-gradient-to-br from-slate-100 via-gray-50 to-blue-50/50">
    <!-- Page Title - modernized -->
    <div class="bg-gradient-to-r from-white via-white to-slate-50/80 border-b border-gray-200/80 px-6 py-5 backdrop-blur-sm">
      <div class="max-w-7xl mx-auto flex items-center justify-between">
        <div class="flex items-center gap-4">
          <RouterLink 
            to="/"
            class="group w-11 h-11 rounded-xl bg-gradient-to-br from-gray-100 to-slate-200 flex items-center justify-center shadow-md hover:shadow-lg hover:from-primary-50 hover:to-blue-100 hover:-translate-y-0.5 transition-all duration-200"
          >
            <div class="i-carbon-arrow-left text-xl text-gray-600 group-hover:text-primary-600 transition-colors" />
          </RouterLink>
          <div>
            <h1 class="text-2xl font-bold text-gray-900">Live Demo</h1>
            <p class="text-sm text-gray-500 mt-0.5">Interactive demonstration of LUCID capabilities</p>
          </div>
        </div>
        <div class="flex items-center gap-3">
          <div class="flex items-center gap-2.5 px-4 py-2 bg-gradient-to-r from-blue-50 to-indigo-50 rounded-xl border border-blue-100 shadow-sm">
            <span class="i-carbon-cube text-lg text-primary-600" />
            <span class="font-bold text-gray-700">MariaDB 12</span>
          </div>
          <div class="flex items-center gap-2.5 px-4 py-2 bg-gradient-to-r from-purple-50 to-violet-50 rounded-xl border border-purple-100 shadow-sm">
            <span class="i-carbon-model-alt text-lg text-purple-600" />
            <span class="font-bold text-gray-700">GPT-4o</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Main Content -->
    <main class="max-w-7xl mx-auto px-6 py-8">
      <!-- Scenario Tabs -->
      <ScenarioTabs 
        v-model="store.currentScenario" 
        class="mb-8"
      />

      <!-- Self-Maintain Demo (full width — has its own layout) -->
      <SelfMaintainDemo 
        v-if="store.currentScenario === 'selfmaintain'"
      />

      <!-- Other demos: split layout with architecture sidebar -->
      <div v-else class="grid grid-cols-12 gap-8">
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
