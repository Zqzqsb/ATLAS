<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import { NButton, NSpin, useMessage } from 'naive-ui'
import { useDatabaseStore } from '@/stores/database'
import DatabaseCard from './DatabaseCard.vue'
import SpiderDatasetCard from './SpiderDatasetCard.vue'
import AddDatabaseDialog from './AddDatabaseDialog.vue'
import type { DatabaseConfig } from '@/types'

const databaseStore = useDatabaseStore()
const message = useMessage()

const showAddDialog = ref(false)

// Spider database name patterns
const SPIDER_PATTERNS = ['spider_tvshow', 'spider_flight', 'spider_wta']

// Check if database belongs to Spider dataset
function isSpiderDatabase(name: string): boolean {
  return name.toLowerCase().startsWith('spider_')
}

// Separate Spider databases from other databases
const spiderDatabases = computed(() => 
  databaseStore.databases.filter(db => isSpiderDatabase(db.name))
)

const otherDatabases = computed(() => 
  databaseStore.databases.filter(db => !isSpiderDatabase(db.name))
)

// Whether to show Spider Dataset card
const showSpiderCard = computed(() => spiderDatabases.value.length > 0)

onMounted(async () => {
  await databaseStore.fetchDatabases()
})

async function handleTestConnection(id: string) {
  const result = await databaseStore.testConnection(id)
  if (result.success) {
    message.success('Connection successful')
  } else {
    message.error(result.message || 'Connection failed')
  }
}

async function handleAddDatabase(config: DatabaseConfig) {
  const result = await databaseStore.addDatabase(config)
  showAddDialog.value = false
  if (result.success) {
    message.success('Connection added, schema synced')
  } else {
    message.error(result.error || 'Failed to add connection')
  }
}
</script>

<template>
  <div class="landing-page min-h-screen bg-slate-50">
    <!-- Hero Section -->
    <div class="relative">
      <div class="max-w-6xl mx-auto px-6 py-12">
        <!-- Header -->
        <div class="text-center mb-10">
          <div class="flex items-center justify-center gap-2.5 mb-4">
            <div class="w-14 h-14 rounded-xl bg-primary-600 flex items-center justify-center">
              <span class="text-white font-semibold text-2xl">L</span>
            </div>
          </div>
          <h1 class="text-3xl font-semibold text-gray-900 tracking-tight mb-3">
            LUCID
          </h1>
          <p class="text-base text-gray-500 max-w-2xl mx-auto mb-3">
            Lakebase-Unified Context-aware Intelligence for Data
          </p>
          <div class="inline-flex items-center gap-2.5 px-4 py-2 rounded-lg bg-white border border-gray-200 text-sm text-gray-500">
            <span class="font-medium">Agent Self-Maintaining</span>
            <span class="w-1 h-1 rounded-full bg-gray-300" />
            <span class="font-medium">Vector Grounding</span>
            <span class="w-1 h-1 rounded-full bg-gray-300" />
            <span class="font-medium">ReAct Reasoning</span>
          </div>
        </div>

        <!-- Database Collection -->
        <div class="mb-10">
          <div class="flex items-center justify-between mb-5">
            <h2 class="text-lg font-semibold text-gray-900 flex items-center gap-2">
              <div class="w-1 h-5 rounded-full bg-primary-500" />
              Databases
            </h2>
          </div>

          <!-- Loading -->
          <div v-if="databaseStore.loading" class="flex justify-center py-20">
            <NSpin size="large" />
          </div>

          <!-- Empty state -->
          <div 
            v-else-if="databaseStore.databases.length === 0"
            class="py-16 text-center rounded-lg bg-white border border-gray-200"
          >
            <div class="w-16 h-16 rounded-lg bg-gray-100 flex items-center justify-center mx-auto mb-4">
              <div class="i-lucide-database text-3xl text-gray-400" />
            </div>
            <p class="text-lg text-gray-700 font-medium mb-1">No databases connected</p>
            <p class="text-gray-400 mb-6 text-sm">Connect your first database to get started</p>
            <button 
              class="px-5 py-2.5 rounded-lg bg-primary-600 text-white font-medium text-sm hover:bg-primary-700 transition-colors"
              @click="showAddDialog = true"
            >
              Add Connection
            </button>
          </div>

          <!-- Database grid -->
          <div 
            v-else
            class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4"
          >
            <!-- Spider Dataset Card (merged display) -->
            <SpiderDatasetCard 
              v-if="showSpiderCard"
              :databases="spiderDatabases"
            />
            
            <!-- Other database cards -->
            <DatabaseCard
              v-for="db in otherDatabases"
              :key="db.id"
              :database="db"
              @test="handleTestConnection"
            />

            <!-- Add new card -->
            <div
              class="database-add-card rounded-lg bg-white border-2 border-dashed border-gray-200 flex flex-col items-center justify-center cursor-pointer hover:border-primary-400 hover:bg-primary-50/50 transition-colors group"
              @click="showAddDialog = true"
            >
              <div class="w-12 h-12 rounded-lg bg-gray-100 flex items-center justify-center mb-3 group-hover:bg-primary-100 transition-colors">
                <div class="i-lucide-plus text-xl text-gray-400 group-hover:text-primary-600 transition-colors" />
              </div>
              <p class="text-gray-600 font-medium text-sm group-hover:text-primary-600 transition-colors">Add New Database</p>
              <p class="text-xs text-gray-400 mt-0.5">MySQL, MariaDB, PostgreSQL</p>
            </div>
          </div>
        </div>

        <!-- Quick Links -->
        <div class="flex items-center justify-center gap-4">
          <RouterLink 
            to="/demo"
            class="group flex items-center gap-3 px-5 py-3 rounded-lg bg-white border border-gray-200 hover:border-gray-300 transition-colors"
          >
            <div class="w-9 h-9 rounded-lg bg-primary-50 text-primary-600 flex items-center justify-center">
              <div class="i-lucide-play text-lg" />
            </div>
            <div>
              <span class="font-medium text-sm text-gray-800 block">Live Demo</span>
              <span class="text-xs text-gray-400">Interactive playground</span>
            </div>
          </RouterLink>

          <a 
            href="https://github.com/zqzqsb/lucid"
            target="_blank"
            class="group flex items-center gap-3 px-5 py-3 rounded-lg bg-white border border-gray-200 hover:border-gray-300 transition-colors"
          >
            <div class="w-9 h-9 rounded-lg bg-gray-100 text-gray-700 flex items-center justify-center">
              <div class="i-lucide-github text-lg" />
            </div>
            <div>
              <span class="font-medium text-sm text-gray-800 block">GitHub</span>
              <span class="text-xs text-gray-400">View source code</span>
            </div>
          </a>
        </div>
      </div>
    </div>

    <!-- Add Database Dialog -->
    <AddDatabaseDialog
      v-model:show="showAddDialog"
      @submit="handleAddDatabase"
    />
  </div>
</template>

<style scoped>
.database-add-card {
  min-height: 220px;
}
</style>
