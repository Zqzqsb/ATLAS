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
            <div class="w-20 h-20 rounded-2xl bg-gradient-to-br from-primary-500 via-blue-500 to-indigo-600 flex items-center justify-center shadow-2xl shadow-primary-500/30 ring-4 ring-white/50">
              <span class="text-white font-serif font-bold text-4xl drop-shadow-lg">L</span>
            </div>
          </div>
          <h1 class="text-5xl font-extrabold text-gray-900 tracking-tight mb-4 drop-shadow-sm">
            LUCID
          </h1>
          <p class="text-xl text-gray-600 font-medium max-w-3xl mx-auto mb-4">
            Lakebase-Unified Context-aware Intelligence for Data
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
            <button 
              class="flex items-center gap-2.5 px-5 py-2.5 rounded-xl bg-gradient-to-r from-primary-500 to-blue-600 text-white font-bold shadow-lg shadow-primary-500/30 hover:shadow-xl hover:shadow-primary-500/40 hover:-translate-y-0.5 transition-all duration-200"
              @click="showAddDialog = true"
            >
              <div class="i-carbon-add text-lg" />
              Connect Database
            </button>
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
              <div class="i-carbon-data-base text-5xl text-gray-400" />
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
            <DatabaseCard
              v-for="db in databaseStore.databases"
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
                <div class="i-carbon-add text-3xl text-gray-400 group-hover:text-primary-600 transition-colors" />
              </div>
              <p class="text-gray-700 font-bold group-hover:text-primary-600 transition-colors">Add New Database</p>
              <p class="text-sm text-gray-500 mt-1">MySQL, MariaDB, PostgreSQL</p>
            </div>
          </div>
        </div>

        <!-- Quick Links - Steam style cards -->
        <div class="flex items-center justify-center gap-5">
          <RouterLink 
            to="/demo"
            class="group flex items-center gap-4 px-6 py-4 rounded-2xl bg-white/70 backdrop-blur-sm border border-white/80 shadow-lg shadow-gray-200/40 hover:shadow-xl hover:bg-white/90 hover:-translate-y-1 transition-all duration-300"
          >
            <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-primary-500 to-blue-600 text-white flex items-center justify-center shadow-lg shadow-primary-500/30 group-hover:scale-105 transition-transform">
              <div class="i-carbon-play-filled text-xl" />
            </div>
            <div>
              <span class="font-bold text-gray-800 group-hover:text-primary-600 transition-colors block">Live Demo</span>
              <span class="text-sm text-gray-500">Interactive playground</span>
            </div>
          </RouterLink>

          <a 
            href="https://github.com/zqzqsb/lucid"
            target="_blank"
            class="group flex items-center gap-4 px-6 py-4 rounded-2xl bg-white/70 backdrop-blur-sm border border-white/80 shadow-lg shadow-gray-200/40 hover:shadow-xl hover:bg-white/90 hover:-translate-y-1 transition-all duration-300"
          >
            <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-gray-700 to-gray-900 text-white flex items-center justify-center shadow-lg shadow-gray-500/30 group-hover:scale-105 transition-transform">
              <div class="i-carbon-logo-github text-2xl" />
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
  height: 260px; /* Match DatabaseCard height */
}
</style>
