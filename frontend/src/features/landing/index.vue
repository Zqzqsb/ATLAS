<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import { NButton, NSpin, useMessage } from 'naive-ui'
import { useDatabaseStore } from '@/stores/database'
import atlasLogo from '@/assets/atlas-logo.svg'
import DatabaseCard from './DatabaseCard.vue'
import SpiderDatasetCard from './SpiderDatasetCard.vue'
import TpchEnterpriseCard from './TpchEnterpriseCard.vue'
import EvolutionCard from './EvolutionCard.vue'
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

// Special database identifiers
const SPECIAL_IDS = new Set(['tpch_enterprise', 'lucid_evolution'])

// Separate Spider databases from other databases
const spiderDatabases = computed(() => 
  databaseStore.databases.filter(db => isSpiderDatabase(db.name))
)

const tpchDatabase = computed(() =>
  databaseStore.databases.find(db => db.name === 'tpch_enterprise') || null
)

const evolutionDatabase = computed(() =>
  databaseStore.databases.find(db => db.name === 'lucid_evolution') || null
)

const otherDatabases = computed(() => 
  databaseStore.databases.filter(db => !isSpiderDatabase(db.name) && !SPECIAL_IDS.has(db.name))
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
  <div class="landing-page min-h-screen bg-gradient-to-br from-slate-100 via-slate-50 to-blue-50">
    <!-- Hero Section -->
    <div class="relative overflow-hidden">
      <!-- Background decoration - Steam style with layered gradients -->
      <div class="absolute inset-0 overflow-hidden pointer-events-none">
        <div class="absolute -top-40 -right-40 w-[600px] h-[600px] rounded-full bg-gradient-to-br from-blue-200/40 to-indigo-300/30 blur-3xl" />
        <div class="absolute -bottom-40 -left-40 w-[600px] h-[600px] rounded-full bg-gradient-to-br from-violet-200/40 to-purple-300/30 blur-3xl" />
        <div class="absolute top-1/3 right-1/4 w-[400px] h-[400px] rounded-full bg-gradient-to-br from-cyan-100/30 to-teal-200/20 blur-3xl" />
      </div>

      <div class="relative max-w-7xl mx-auto px-6 py-16">
        <!-- Header -->
        <div class="text-center mb-14">
          <div class="flex items-center justify-center gap-3 mb-6">
<img :src="atlasLogo" alt="ATLAS" class="w-20 h-20 rounded-2xl shadow-2xl shadow-primary-500/30 ring-4 ring-white/50" />
          </div>
          <h1 class="text-5xl font-extrabold text-gray-900 tracking-tight mb-4 drop-shadow-sm">
            ATLAS
          </h1>
          <p class="text-xl text-gray-600 font-medium max-w-3xl mx-auto mb-4">
            Adaptive Text-to-SQL with Lifecycle-Aware Self-Maintaining Context
          </p>
          <div class="inline-flex items-center gap-3 px-5 py-2.5 rounded-full bg-white/60 backdrop-blur-sm border border-white/80 shadow-lg shadow-gray-200/50">
            <span class="text-sm font-semibold text-gray-600">Agent Self-Maintaining</span>
            <span class="w-1.5 h-1.5 rounded-full bg-gradient-to-r from-primary-400 to-blue-500" />
            <span class="text-sm font-semibold text-gray-600">Vector Grounding</span>
            <span class="w-1.5 h-1.5 rounded-full bg-gradient-to-r from-blue-400 to-indigo-500" />
            <span class="text-sm font-semibold text-gray-600">ReAct Reasoning</span>
          </div>
        </div>

        <!-- Database Collection -->
        <div class="mb-14">
          <div class="flex items-center justify-between mb-8">
            <h2 class="text-2xl font-bold text-gray-900 flex items-center gap-3">
              <div class="w-1.5 h-8 rounded-full bg-gradient-to-b from-primary-500 to-blue-600" />
              Databases
            </h2>
          </div>

          <!-- Loading -->
          <div v-if="databaseStore.loading" class="flex justify-center py-24">
            <NSpin size="large" />
          </div>

          <!-- Empty state -->
          <div 
            v-else-if="databaseStore.databases.length === 0"
            class="py-20 text-center rounded-2xl bg-white/40 backdrop-blur-sm border border-white/60 shadow-xl shadow-gray-200/30"
          >
            <div class="w-24 h-24 rounded-2xl bg-gradient-to-br from-gray-100 to-slate-200 flex items-center justify-center mx-auto mb-6 shadow-lg">
              <div class="i-lucide-database text-5xl text-gray-400" />
            </div>
            <p class="text-xl text-gray-700 font-bold mb-2">No databases connected</p>
            <p class="text-gray-500 mb-8">Connect your first database to get started</p>
            <button 
              class="px-6 py-3 rounded-xl bg-gradient-to-r from-primary-500 to-blue-600 text-white font-bold shadow-lg shadow-primary-500/30 hover:shadow-xl hover:-translate-y-0.5 transition-all"
              @click="showAddDialog = true"
            >
              Add Connection
            </button>
          </div>

          <!-- Database grid -->
          <div 
            v-else
            class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-5"
          >
            <!-- Spider Dataset Card (merged display) -->
            <SpiderDatasetCard 
              v-if="showSpiderCard"
              :databases="spiderDatabases"
            />

            <!-- TPC-H Enterprise Card -->
            <TpchEnterpriseCard
              v-if="tpchDatabase"
              :database="tpchDatabase"
            />

            <!-- Evolution Demo Card -->
            <EvolutionCard
              v-if="evolutionDatabase"
              :database="evolutionDatabase"
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
              class="database-add-card rounded-2xl bg-gradient-to-br from-white/60 to-slate-100/60 backdrop-blur-sm border-2 border-dashed border-gray-300/80 flex flex-col items-center justify-center cursor-pointer hover:border-primary-400 hover:from-primary-50/80 hover:to-blue-50/80 hover:shadow-xl hover:-translate-y-1 transition-all duration-300 group"
              @click="showAddDialog = true"
            >
              <div class="w-16 h-16 rounded-xl bg-white flex items-center justify-center mb-4 shadow-lg group-hover:shadow-xl group-hover:scale-105 transition-all duration-300">
                <div class="i-lucide-plus text-3xl text-gray-400 group-hover:text-primary-600 transition-colors" />
              </div>
              <p class="text-gray-700 font-bold group-hover:text-primary-600 transition-colors">Add New Database</p>
              <p class="text-sm text-gray-500 mt-1">MySQL, MariaDB, PostgreSQL</p>
            </div>
          </div>
        </div>

        <!-- Quick Links -->
        <div class="flex items-center justify-center gap-5">
          <RouterLink 
            to="/features"
            class="group flex items-center gap-4 px-6 py-4 rounded-2xl bg-white/70 backdrop-blur-sm border border-white/80 shadow-lg shadow-gray-200/40 hover:shadow-xl hover:bg-white/90 hover:-translate-y-1 transition-all duration-300"
          >
            <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-primary-500 to-violet-600 text-white flex items-center justify-center shadow-lg shadow-primary-500/30 group-hover:scale-105 transition-transform">
              <div class="i-lucide-sparkles text-xl" />
            </div>
            <div>
              <span class="font-bold text-gray-800 group-hover:text-primary-600 transition-colors block">Feature Showcase</span>
              <span class="text-sm text-gray-500">Explore 4 innovations</span>
            </div>
          </RouterLink>

          <a 
            href="https://github.com/zqzqsb/lucid"
            target="_blank"
            class="group flex items-center gap-4 px-6 py-4 rounded-2xl bg-white/70 backdrop-blur-sm border border-white/80 shadow-lg shadow-gray-200/40 hover:shadow-xl hover:bg-white/90 hover:-translate-y-1 transition-all duration-300"
          >
            <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-gray-700 to-gray-900 text-white flex items-center justify-center shadow-lg shadow-gray-500/30 group-hover:scale-105 transition-transform">
              <div class="i-lucide-github text-2xl" />
            </div>
            <div>
              <span class="font-bold text-gray-800 group-hover:text-gray-900 transition-colors block">GitHub</span>
              <span class="text-sm text-gray-500">View source code</span>
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
  min-height: 260px;
}
</style>
