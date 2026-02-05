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
  <div class="landing-page min-h-screen bg-gray-50">
    <!-- Hero Section -->
    <div class="relative overflow-hidden bg-white border-b border-gray-100">
      <!-- Background decoration -->
      <div class="absolute inset-0 overflow-hidden pointer-events-none">
        <div class="absolute -top-40 -right-40 w-96 h-96 rounded-full bg-blue-50 blur-3xl opacity-50" />
        <div class="absolute -bottom-40 -left-40 w-96 h-96 rounded-full bg-purple-50 blur-3xl opacity-50" />
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
              class="h-[280px] rounded-xl bg-gray-50 border-2 border-dashed border-gray-300 flex flex-col items-center justify-center cursor-pointer hover:border-primary-400 hover:bg-white transition-all duration-300 group"
              @click="showAddDialog = true"
            >
              <div class="w-16 h-16 rounded-2xl bg-white flex items-center justify-center mb-4 border border-gray-200 shadow-sm group-hover:scale-110 transition-transform duration-300">
                <div class="i-carbon-add text-3xl text-gray-400 group-hover:text-primary-600 transition-colors" />
              </div>
              <p class="text-gray-500 font-bold group-hover:text-primary-600 transition-colors">Add New Database</p>
            </div>
          </div>
        </div>

        <!-- Quick Links -->
        <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
          <RouterLink 
            to="/demo"
            class="p-8 rounded-xl bg-white border border-gray-200 hover:border-primary-200 hover:shadow-lg transition-all duration-300 group block"
          >
            <div class="flex items-center gap-3 mb-3">
              <div class="p-2 rounded-lg bg-blue-50 text-blue-600">
                <div class="i-carbon-play text-2xl" />
              </div>
              <h3 class="text-lg font-bold text-gray-900 group-hover:text-primary-600 transition-colors">Live Demo</h3>
            </div>
            <p class="text-sm text-gray-500 font-medium">Explore LUCID capabilities with interactive demos</p>
          </RouterLink>

          <a 
            href="#"
            class="p-8 rounded-xl bg-white border border-gray-200 hover:border-cyan-200 hover:shadow-lg transition-all duration-300 group"
          >
            <div class="flex items-center gap-3 mb-3">
              <div class="p-2 rounded-lg bg-cyan-50 text-cyan-600">
                <div class="i-carbon-document text-2xl" />
              </div>
              <h3 class="text-lg font-bold text-gray-900 group-hover:text-cyan-600 transition-colors">Documentation</h3>
            </div>
            <p class="text-sm text-gray-500 font-medium">Read detailed guides and API references</p>
          </a>

          <a 
            href="https://github.com/lucid-sql/lucid"
            target="_blank"
            class="p-8 rounded-xl bg-white border border-gray-200 hover:border-purple-200 hover:shadow-lg transition-all duration-300 group"
          >
            <div class="flex items-center gap-3 mb-3">
              <div class="p-2 rounded-lg bg-purple-50 text-purple-600">
                <div class="i-carbon-logo-github text-2xl" />
              </div>
              <h3 class="text-lg font-bold text-gray-900 group-hover:text-purple-600 transition-colors">GitHub</h3>
            </div>
            <p class="text-sm text-gray-500 font-medium">View source code and contribute</p>
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
