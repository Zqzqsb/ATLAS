<script setup lang="ts">
import { ref, onMounted } from 'vue'
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
  <div class="min-h-screen bg-gray-50">
    <!-- Header -->
    <header class="bg-white/80 backdrop-blur-md border-b border-gray-200 sticky top-0 z-50 h-16 transition-all">
      <div class="max-w-7xl mx-auto px-6 h-full flex-between">
        <div class="flex items-center gap-4">
          <div class="w-8 h-8 rounded-lg bg-primary-600 shadow-sm flex items-center justify-center">
            <span class="text-white font-serif font-bold text-lg">L</span>
          </div>
          <div class="flex flex-col">
            <h1 class="text-xl font-bold text-gray-900 leading-none">LUCID</h1>
            <span class="text-xs text-gray-500 font-bold uppercase tracking-wider mt-0.5">Live Demo Environment</span>
          </div>
        </div>
        <div class="flex items-center gap-6 text-sm font-medium text-gray-600">
          <div class="flex items-center gap-2 px-3 py-1 bg-gray-100 rounded-full border border-gray-200">
            <span class="i-carbon-cube text-primary-600 text-lg" />
            <span class="font-bold text-gray-700">MariaDB 12 Vector</span>
          </div>
          <div class="flex items-center gap-2 px-3 py-1 bg-gray-100 rounded-full border border-gray-200">
            <span class="i-carbon-model-alt text-purple-600 text-lg" />
            <span class="font-bold text-gray-700">OpenAI GPT-4o</span>
          </div>
        </div>
      </div>
    </header>

    <!-- Main Content -->
    <main class="max-w-7xl mx-auto px-6 py-8">
      <!-- Scenario Tabs -->
      <ScenarioTabs 
        v-model="store.currentScenario" 
        class="mb-8"
      />

      <!-- Content Grid -->
      <div class="grid grid-cols-12 gap-8">
        <!-- Left: Main Demo Area -->
        <div class="col-span-8">
          <!-- Context Comparison (Core) -->
          <ContextComparison 
            v-if="store.currentScenario === 'comparison'"
          />

          <!-- Self-Maintain Demo -->
          <SelfMaintainDemo 
            v-else-if="store.currentScenario === 'selfmaintain'"
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
