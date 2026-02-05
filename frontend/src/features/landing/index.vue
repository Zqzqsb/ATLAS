<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import { NButton, NSpin, useMessage } from 'naive-ui'
import { useDatabaseStore } from '@/stores/database'
import DatabaseCard from './DatabaseCard.vue'
import AddDatabaseDialog from './AddDatabaseDialog.vue'
import type { DatabaseConfig } from '@/types'

const databaseStore = useDatabaseStore()
const message = useMessage()

const showAddDialog = ref(false)

onMounted(async () => {
  await databaseStore.fetchDatabases()
})

async function handleTestConnection(id: string) {
  const result = await databaseStore.testConnection(id)
  if (result.success) {
    message.success('连接成功')
  } else {
    message.error(result.message || '连接失败')
  }
}

async function handleAddDatabase(config: DatabaseConfig) {
  const db = await databaseStore.addDatabase(config)
  if (db) {
    message.success('添加成功')
  } else {
    message.error('添加失败')
  }
}
</script>

<template>
  <div class="landing-page min-h-screen bg-gradient-to-b from-slate-50 via-white to-slate-50">
    <!-- Hero Section -->
    <div class="relative overflow-hidden">
      <!-- Background decoration -->
      <div class="absolute inset-0 overflow-hidden pointer-events-none">
        <div class="absolute -top-40 -right-40 w-[500px] h-[500px] rounded-full bg-gradient-to-br from-blue-100/60 to-indigo-100/60 blur-3xl" />
        <div class="absolute -bottom-40 -left-40 w-[500px] h-[500px] rounded-full bg-gradient-to-br from-violet-100/60 to-purple-100/60 blur-3xl" />
        <div class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[800px] h-[800px] rounded-full bg-gradient-radial from-primary-50/30 to-transparent" />
        <!-- Grid pattern overlay -->
        <div class="absolute inset-0 bg-[linear-gradient(rgba(0,0,0,0.02)_1px,transparent_1px),linear-gradient(90deg,rgba(0,0,0,0.02)_1px,transparent_1px)] bg-[size:40px_40px]" />
      </div>

      <div class="relative max-w-7xl mx-auto px-6 py-20">
        <!-- Header -->
        <div class="text-center mb-16">
          <div class="flex items-center justify-center gap-3 mb-6">
            <div class="w-20 h-20 rounded-2xl bg-gradient-to-br from-primary-500 to-blue-600 flex items-center justify-center shadow-xl shadow-primary-500/20">
              <span class="text-white font-serif font-bold text-4xl">L</span>
            </div>
          </div>
          <h1 class="text-5xl font-extrabold text-gray-900 tracking-tight mb-4">
            LUCID
          </h1>
          <p class="text-xl text-gray-600 font-medium max-w-3xl mx-auto mb-3">
            Lakebase-Unified Context-aware Intelligence for Data
          </p>
          <div class="flex items-center justify-center gap-2 text-sm font-semibold text-gray-500 uppercase tracking-wider">
            <span>Agent Self-Maintaining</span>
            <span class="w-1 h-1 rounded-full bg-gray-300" />
            <span>Vector Grounding</span>
            <span class="w-1 h-1 rounded-full bg-gray-300" />
            <span>ReAct Reasoning</span>
          </div>
        </div>

        <!-- Database Collection -->
        <div class="mb-16">
          <div class="flex items-center justify-between mb-8">
            <h2 class="text-2xl font-bold text-gray-900 flex items-center gap-3">
              <div class="w-1.5 h-8 bg-primary-600 rounded-full" />
              Databases
            </h2>
            <NButton 
              type="primary" 
              size="large"
              class="shadow-sm font-bold"
              @click="showAddDialog = true"
            >
              <template #icon>
                <div class="i-carbon-add" />
              </template>
              Connect Database
            </NButton>
          </div>

          <!-- Loading -->
          <div v-if="databaseStore.loading" class="flex justify-center py-24">
            <NSpin size="large" />
          </div>

          <!-- Empty state -->
          <div 
            v-else-if="databaseStore.databases.length === 0"
            class="py-24 text-center bg-gray-50 rounded-2xl border border-dashed border-gray-300"
          >
            <div class="w-24 h-24 rounded-2xl bg-white flex items-center justify-center mx-auto mb-6 border border-gray-200 shadow-sm">
              <div class="i-carbon-data-base text-5xl text-gray-300" />
            </div>
            <p class="text-xl text-gray-600 font-bold mb-2">No databases connected</p>
            <p class="text-gray-500 mb-8">Connect your first database to get started</p>
            <NButton 
              type="primary" 
              size="large"
              @click="showAddDialog = true"
            >
              Add Connection
            </NButton>
          </div>

          <!-- Database grid -->
          <div 
            v-else
            class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6"
          >
            <DatabaseCard
              v-for="db in databaseStore.databases"
              :key="db.id"
              :database="db"
              @test="handleTestConnection"
            />

            <!-- Add new card -->
            <div
              class="database-add-card rounded-xl bg-gradient-to-br from-gray-50 to-slate-100 border-2 border-dashed border-gray-300 flex flex-col items-center justify-center cursor-pointer hover:border-primary-400 hover:from-primary-50 hover:to-blue-50 transition-all group"
              @click="showAddDialog = true"
            >
              <div class="w-14 h-14 rounded-xl bg-white flex items-center justify-center mb-4 border border-gray-200 shadow-sm group-hover:border-primary-300 group-hover:shadow-md transition-all">
                <div class="i-carbon-add text-2xl text-gray-400 group-hover:text-primary-600 transition-colors" />
              </div>
              <p class="text-gray-600 font-semibold group-hover:text-primary-600 transition-colors">Add New Database</p>
              <p class="text-xs text-gray-400 mt-1">Connect to MySQL, MariaDB, PostgreSQL</p>
            </div>
          </div>
        </div>

        <!-- Quick Links -->
        <div class="flex items-center justify-center gap-4">
          <RouterLink 
            to="/demo"
            class="flex items-center gap-3 px-5 py-3 rounded-lg bg-white border border-gray-200 hover:border-primary-300 hover:shadow-md transition-all group"
          >
            <div class="w-10 h-10 rounded-lg bg-primary-50 text-primary-600 flex items-center justify-center group-hover:bg-primary-100 transition-colors">
              <div class="i-carbon-play text-xl" />
            </div>
            <span class="font-semibold text-gray-700 group-hover:text-primary-600 transition-colors">Live Demo</span>
          </RouterLink>

          <a 
            href="https://github.com/zqzqsb/lucid"
            target="_blank"
            class="flex items-center gap-3 px-5 py-3 rounded-lg bg-white border border-gray-200 hover:border-gray-400 hover:shadow-md transition-all group"
          >
            <div class="w-10 h-10 rounded-lg bg-gray-100 text-gray-700 flex items-center justify-center group-hover:bg-gray-200 transition-colors">
              <div class="i-carbon-logo-github text-xl" />
            </div>
            <span class="font-semibold text-gray-700 group-hover:text-gray-900 transition-colors">GitHub</span>
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
  height: 246px; /* Match DatabaseCard height (240px + 6px for top bar) */
}
</style>
