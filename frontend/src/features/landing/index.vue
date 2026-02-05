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
          <div class="flex items-center justify-center gap-3 text-sm font-bold uppercase tracking-wider">
            <span class="px-4 py-1.5 rounded-full bg-gradient-to-r from-blue-50 to-indigo-50 text-blue-600 border border-blue-200 hover:shadow-md hover:from-blue-100 hover:to-indigo-100 transition-all cursor-default">
              Agent Self-Maintaining
            </span>
            <span class="px-4 py-1.5 rounded-full bg-gradient-to-r from-emerald-50 to-teal-50 text-emerald-600 border border-emerald-200 hover:shadow-md hover:from-emerald-100 hover:to-teal-100 transition-all cursor-default">
              Vector Grounding
            </span>
            <span class="px-4 py-1.5 rounded-full bg-gradient-to-r from-violet-50 to-purple-50 text-violet-600 border border-violet-200 hover:shadow-md hover:from-violet-100 hover:to-purple-100 transition-all cursor-default">
              ReAct Reasoning
            </span>
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
              class="h-[280px] rounded-2xl bg-gradient-to-br from-slate-50 to-gray-100 border-2 border-dashed border-gray-300 flex flex-col items-center justify-center cursor-pointer hover:border-primary-400 hover:from-primary-50 hover:to-blue-50 hover:shadow-lg hover:shadow-primary-100/50 transition-all duration-300 group"
              @click="showAddDialog = true"
            >
              <div class="w-16 h-16 rounded-2xl bg-white flex items-center justify-center mb-4 border border-gray-200 shadow-sm group-hover:scale-110 group-hover:bg-gradient-to-br group-hover:from-primary-500 group-hover:to-blue-600 group-hover:border-transparent group-hover:shadow-lg group-hover:shadow-primary-500/30 transition-all duration-300">
                <div class="i-carbon-add text-3xl text-gray-400 group-hover:text-white transition-colors" />
              </div>
              <p class="text-gray-500 font-bold group-hover:text-primary-600 transition-colors text-lg">Add New Database</p>
              <p class="text-gray-400 text-sm mt-1 group-hover:text-primary-500 transition-colors">Click to connect</p>
            </div>
          </div>
        </div>

        <!-- Quick Links -->
        <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
          <RouterLink 
            to="/demo"
            class="quick-link-card p-6 rounded-2xl bg-gradient-to-br from-blue-50 to-indigo-50 border border-blue-100 hover:border-blue-300 hover:shadow-xl hover:shadow-blue-100/50 hover:-translate-y-1 transition-all duration-300 group block"
          >
            <div class="flex items-center gap-4 mb-4">
              <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-blue-500 to-indigo-600 text-white flex items-center justify-center shadow-lg shadow-blue-500/30 group-hover:scale-110 transition-transform duration-300">
                <div class="i-carbon-play text-2xl" />
              </div>
              <div>
                <h3 class="text-lg font-bold text-gray-900 group-hover:text-blue-600 transition-colors">Live Demo</h3>
                <p class="text-xs text-blue-500 font-medium">Interactive preview</p>
              </div>
            </div>
            <p class="text-sm text-gray-600 font-medium leading-relaxed">Explore LUCID capabilities with interactive demonstrations</p>
          </RouterLink>

          <a 
            href="#"
            class="quick-link-card p-6 rounded-2xl bg-gradient-to-br from-emerald-50 to-teal-50 border border-emerald-100 hover:border-emerald-300 hover:shadow-xl hover:shadow-emerald-100/50 hover:-translate-y-1 transition-all duration-300 group block"
          >
            <div class="flex items-center gap-4 mb-4">
              <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-emerald-500 to-teal-600 text-white flex items-center justify-center shadow-lg shadow-emerald-500/30 group-hover:scale-110 transition-transform duration-300">
                <div class="i-carbon-document text-2xl" />
              </div>
              <div>
                <h3 class="text-lg font-bold text-gray-900 group-hover:text-emerald-600 transition-colors">Documentation</h3>
                <p class="text-xs text-emerald-500 font-medium">API & guides</p>
              </div>
            </div>
            <p class="text-sm text-gray-600 font-medium leading-relaxed">Read detailed guides and comprehensive API references</p>
          </a>

          <a 
            href="https://github.com/lucid-sql/lucid"
            target="_blank"
            class="quick-link-card p-6 rounded-2xl bg-gradient-to-br from-violet-50 to-purple-50 border border-violet-100 hover:border-violet-300 hover:shadow-xl hover:shadow-violet-100/50 hover:-translate-y-1 transition-all duration-300 group block"
          >
            <div class="flex items-center gap-4 mb-4">
              <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-violet-500 to-purple-600 text-white flex items-center justify-center shadow-lg shadow-violet-500/30 group-hover:scale-110 transition-transform duration-300">
                <div class="i-carbon-logo-github text-2xl" />
              </div>
              <div>
                <h3 class="text-lg font-bold text-gray-900 group-hover:text-violet-600 transition-colors">GitHub</h3>
                <p class="text-xs text-violet-500 font-medium">Open source</p>
              </div>
            </div>
            <p class="text-sm text-gray-600 font-medium leading-relaxed">View source code and contribute to the project</p>
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
