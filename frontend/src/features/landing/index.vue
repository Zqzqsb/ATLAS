<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { NButton, NEmpty, NSpin, useMessage } from 'naive-ui'
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
  <div class="landing-page min-h-screen bg-gradient-to-br from-gray-50 to-gray-100 dark:from-gray-900 dark:to-gray-950">
    <!-- Hero Section -->
    <div class="relative overflow-hidden">
      <!-- Background decoration -->
      <div class="absolute inset-0 overflow-hidden">
        <div class="absolute -top-40 -right-40 w-80 h-80 rounded-full bg-blue-500/10 blur-3xl" />
        <div class="absolute -bottom-40 -left-40 w-80 h-80 rounded-full bg-purple-500/10 blur-3xl" />
      </div>

      <div class="relative max-w-7xl mx-auto px-6 py-16">
        <!-- Header -->
        <div class="text-center mb-12">
          <div class="flex items-center justify-center gap-3 mb-4">
            <div class="w-16 h-16 rounded-2xl bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center shadow-lg">
              <span class="text-white font-bold text-2xl">LC</span>
            </div>
          </div>
          <h1 class="text-4xl font-bold text-gray-800 dark:text-gray-100 mb-3">
            LUCID
          </h1>
          <p class="text-lg text-gray-600 dark:text-gray-400 max-w-2xl mx-auto">
            Lakebase-Unified Context-aware Intelligence for Data
            <br>
            <span class="text-sm">Agent Self-Maintaining · Vector Grounding · ReAct Reasoning</span>
          </p>
        </div>

        <!-- Database Collection -->
        <div class="mb-8">
          <div class="flex items-center justify-between mb-6">
            <h2 class="text-xl font-semibold text-gray-800 dark:text-gray-100">
              我的数据库
            </h2>
            <NButton type="primary" @click="showAddDialog = true">
              <template #icon>
                <div class="i-carbon-add" />
              </template>
              添加连接
            </NButton>
          </div>

          <!-- Loading -->
          <div v-if="databaseStore.loading" class="flex justify-center py-16">
            <NSpin size="large" />
          </div>

          <!-- Empty state -->
          <NEmpty 
            v-else-if="databaseStore.databases.length === 0"
            description="暂无数据库连接"
            class="py-16"
          >
            <template #extra>
              <NButton type="primary" @click="showAddDialog = true">
                添加第一个连接
              </NButton>
            </template>
          </NEmpty>

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
              class="min-h-[240px] rounded-lg border-2 border-dashed border-gray-300 dark:border-gray-700 flex items-center justify-center cursor-pointer hover:border-blue-400 hover:bg-blue-50/50 dark:hover:bg-blue-900/20 transition-colors"
              @click="showAddDialog = true"
            >
              <div class="text-center">
                <div class="w-12 h-12 rounded-full bg-gray-100 dark:bg-gray-800 flex items-center justify-center mx-auto mb-3">
                  <div class="i-carbon-add text-2xl text-gray-400" />
                </div>
                <p class="text-gray-500">添加新数据库</p>
              </div>
            </div>
          </div>
        </div>

        <!-- Quick Links -->
        <div class="flex justify-center gap-4 mt-12">
          <RouterLink to="/demo">
            <NButton quaternary size="large">
              <template #icon>
                <div class="i-carbon-demo" />
              </template>
              查看演示
            </NButton>
          </RouterLink>
          <NButton quaternary size="large">
            <template #icon>
              <div class="i-carbon-document" />
            </template>
            文档
          </NButton>
          <NButton quaternary size="large">
            <template #icon>
              <div class="i-carbon-logo-github" />
            </template>
            GitHub
          </NButton>
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
