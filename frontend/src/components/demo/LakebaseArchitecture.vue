<script setup lang="ts">
import { ref } from 'vue'

const showArchitecture = ref(true)

const storageItems = [
  { label: 'Schema Metadata', type: 'RELATIONAL', icon: 'i-lucide-table-2', color: 'blue' },
  { label: 'Vector Index', type: 'VECTOR', icon: 'i-lucide-shapes', color: 'purple' },
  { label: 'Rich Context', type: 'JSON', icon: 'i-lucide-file-text', color: 'green' },
  { label: 'SQL Templates', type: 'SQL', icon: 'i-lucide-code-2', color: 'orange' }
]
</script>

<template>
  <div class="space-y-4">
    <!-- Storage Status -->
    <div class="card p-4">
      <div class="flex-between mb-3">
        <h3 class="font-medium flex items-center gap-2">
          <span class="i-lucide-box text-blue-500" />
          Lake-Base Unified Storage
        </h3>
        <button 
          @click="showArchitecture = !showArchitecture"
          class="text-sm text-blue-500 hover:underline"
        >
          {{ showArchitecture ? 'Collapse' : 'Expand' }} Comparison
        </button>
      </div>

      <div class="space-y-2">
        <div 
          v-for="item in storageItems" 
          :key="item.label"
          class="flex-between p-2 bg-gray-50 rounded"
        >
          <div class="flex items-center gap-2">
            <span :class="item.icon" class="text-gray-500" />
            <span class="text-sm">{{ item.label }}</span>
          </div>
          <span 
            class="text-xs px-1.5 py-0.5 rounded"
            :class="{
              'bg-blue-100 text-blue-600': item.color === 'blue',
              'bg-purple-100 text-purple-600': item.color === 'purple',
              'bg-green-100 text-green-600': item.color === 'green',
              'bg-orange-100 text-orange-600': item.color === 'orange'
            }"
          >
            {{ item.type }}
          </span>
        </div>
      </div>

      <div class="mt-3 text-xs text-gray-500 text-center">
        All stored in MariaDB &middot; Single-DB Lake
      </div>
    </div>

    <!-- Architecture Comparison -->
    <div v-if="showArchitecture" class="card p-4">
      <h3 class="font-medium mb-3 text-sm">Architecture Comparison</h3>
      
      <div class="grid grid-cols-2 gap-3">
        <!-- Traditional -->
        <div class="p-3 bg-red-50 rounded-lg border border-red-100">
          <div class="text-xs font-medium text-red-700 mb-2">Traditional Approach</div>
          <div class="flex gap-2 justify-center mb-2">
            <div class="px-2 py-1 bg-white rounded text-xs shadow-sm">MySQL</div>
            <div class="px-2 py-1 bg-white rounded text-xs shadow-sm">Milvus</div>
          </div>
          <div class="text-center text-xs text-red-600">
            <div class="i-lucide-x-filled inline-block mr-1" />
            Data sync latency
          </div>
        </div>

        <!-- Lakebase -->
        <div class="p-3 bg-green-50 rounded-lg border border-green-100">
          <div class="text-xs font-medium text-green-700 mb-2">Lake-Base Approach</div>
          <div class="flex justify-center mb-2">
            <div class="px-3 py-1 bg-white rounded text-xs shadow-sm">MariaDB</div>
          </div>
          <div class="text-center text-xs text-green-600">
            <div class="i-lucide-check-filled inline-block mr-1" />
            Native consistency
          </div>
        </div>
      </div>
    </div>

    <!-- Innovation Points -->
    <div class="card p-4">
      <h3 class="font-medium mb-3 text-sm flex items-center gap-2">
        <span class="i-lucide-lightbulb text-yellow-500" />
        Key Innovations
      </h3>
      <ul class="space-y-2 text-sm">
        <li class="flex items-start gap-2">
          <span class="i-lucide-check text-green-500 mt-0.5" />
          <span>Lake-base multi-modal unified storage</span>
        </li>
        <li class="flex items-start gap-2">
          <span class="i-lucide-check text-green-500 mt-0.5" />
          <span>In-database vector search for Schema Linking</span>
        </li>
        <li class="flex items-start gap-2">
          <span class="i-lucide-check text-green-500 mt-0.5" />
          <span>Agent-driven self-maintenance</span>
        </li>
        <li class="flex items-start gap-2">
          <span class="i-lucide-check text-green-500 mt-0.5" />
          <span>End-to-end, zero external dependencies</span>
        </li>
      </ul>
    </div>
  </div>
</template>
