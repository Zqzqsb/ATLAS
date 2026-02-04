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
  <div class="landing-page min-h-screen bg-gradient-to-br from-gray-900 via-slate-900 to-gray-950">
    <!-- Hero Section -->
    <div class="relative overflow-hidden">
      <!-- Background decoration with animated gradient -->
      <div class="absolute inset-0 overflow-hidden">
        <div class="absolute -top-40 -right-40 w-96 h-96 rounded-full bg-blue-600/20 blur-3xl animate-pulse" />
        <div class="absolute -bottom-40 -left-40 w-96 h-96 rounded-full bg-purple-600/20 blur-3xl animate-pulse" style="animation-delay: 1s;" />
        <div class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-96 h-96 rounded-full bg-cyan-600/10 blur-3xl" />
      </div>

      <div class="relative max-w-7xl mx-auto px-6 py-16">
        <!-- Header -->
        <div class="text-center mb-16">
          <div class="flex items-center justify-center gap-3 mb-6">
            <div class="w-20 h-20 rounded-2xl bg-gradient-to-br from-blue-500 via-cyan-500 to-purple-600 flex items-center justify-center shadow-2xl shadow-blue-500/50 animate-pulse">
              <span class="text-white font-bold text-3xl">LC</span>
            </div>
          </div>
          <h1 class="text-5xl font-bold bg-gradient-to-r from-blue-400 via-cyan-400 to-purple-400 bg-clip-text text-transparent mb-4">
            LUCID
          </h1>
          <p class="text-xl text-gray-300 max-w-3xl mx-auto mb-2">
            Lakebase-Unified Context-aware Intelligence for Data
          </p>
          <p class="text-sm text-gray-500">
            Agent Self-Maintaining · Vector Grounding · ReAct Reasoning
          </p>
        </div>

        <!-- Database Collection -->
        <div class="mb-12">
          <div class="flex items-center justify-between mb-8">
            <h2 class="text-2xl font-bold text-white flex items-center gap-3">
              <div class="w-1 h-8 bg-gradient-to-b from-blue-500 to-cyan-500 rounded-full" />
              我的数据库
            </h2>
            <NButton 
              type="primary" 
              size="large"
              class="shadow-lg shadow-blue-500/30"
              @click="showAddDialog = true"
            >
              <template #icon>
                <div class="i-carbon-add" />
              </template>
              添加连接
            </NButton>
          </div>

          <!-- Loading -->
          <div v-if="databaseStore.loading" class="flex justify-center py-24">
            <NSpin size="large" />
          </div>

          <!-- Empty state -->
          <div 
            v-else-if="databaseStore.databases.length === 0"
            class="py-24 text-center"
          >
            <div class="w-24 h-24 rounded-2xl bg-white/5 backdrop-blur-sm flex items-center justify-center mx-auto mb-6 border border-white/10">
              <div class="i-carbon-data-base text-5xl text-white/40" />
            </div>
            <p class="text-xl text-white/60 mb-6">暂无数据库连接</p>
            <NButton 
              type="primary" 
              size="large"
              @click="showAddDialog = true"
            >
              添加第一个连接
            </NButton>
          </div>

          <!-- Database grid - Steam library style -->
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

            <!-- Add new card - Steam style -->
            <div
              class="h-[320px] rounded-xl bg-gradient-to-br from-gray-800/50 to-gray-900/50 border-2 border-dashed border-white/20 backdrop-blur-sm flex items-center justify-center cursor-pointer hover:border-blue-400 hover:bg-gradient-to-br hover:from-blue-900/20 hover:to-cyan-900/20 transition-all duration-300 group"
              @click="showAddDialog = true"
            >
              <div class="text-center">
                <div class="w-16 h-16 rounded-2xl bg-white/10 backdrop-blur-md flex items-center justify-center mx-auto mb-4 border border-white/20 group-hover:bg-white/15 group-hover:scale-110 transition-all duration-300">
                  <div class="i-carbon-add text-3xl text-white/60 group-hover:text-white/80" />
                </div>
                <p class="text-white/60 group-hover:text-white/80 transition-colors">添加新数据库</p>
              </div>
            </div>
          </div>
        </div>

        <!-- Quick Links -->
        <div class="grid grid-cols-1 md:grid-cols-3 gap-4 mt-12">
          <RouterLink 
            to="/demo"
            class="p-6 rounded-xl bg-white/5 backdrop-blur-sm border border-white/10 hover:bg-white/10 hover:border-white/20 transition-all duration-300 group block"
          >
            <div class="flex items-center gap-3 mb-2">
              <div class="i-carbon-play text-2xl text-blue-400" />
              <h3 class="text-lg font-semibold text-white">查看演示</h3>
            </div>
            <p class="text-sm text-gray-400 group-hover:text-gray-300">体验 LUCID 核心功能演示</p>
          </RouterLink>

          <a 
            href="#"
            class="p-6 rounded-xl bg-white/5 backdrop-blur-sm border border-white/10 hover:bg-white/10 hover:border-white/20 transition-all duration-300 group"
          >
            <div class="flex items-center gap-3 mb-2">
              <div class="i-carbon-document text-2xl text-cyan-400" />
              <h3 class="text-lg font-semibold text-white">文档</h3>
            </div>
            <p class="text-sm text-gray-400 group-hover:text-gray-300">查看详细使用文档</p>
          </a>

          <a 
            href="https://github.com/lucid-sql/lucid"
            target="_blank"
            class="p-6 rounded-xl bg-white/5 backdrop-blur-sm border border-white/10 hover:bg-white/10 hover:border-white/20 transition-all duration-300 group"
          >
            <div class="flex items-center gap-3 mb-2">
              <div class="i-carbon-logo-github text-2xl text-purple-400" />
              <h3 class="text-lg font-semibold text-white">GitHub</h3>
            </div>
            <p class="text-sm text-gray-400 group-hover:text-gray-300">访问开源仓库</p>
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
