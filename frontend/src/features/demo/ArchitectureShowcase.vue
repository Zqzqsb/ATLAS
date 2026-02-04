<script setup lang="ts">
import { NCard, NTag, NGrid, NGridItem } from 'naive-ui'
</script>

<template>
  <div class="architecture-showcase">
    <!-- Intro -->
    <div class="text-center mb-8">
      <h2 class="text-2xl font-bold text-gray-800 dark:text-gray-100 mb-2">
        湖仓一体架构
      </h2>
      <p class="text-gray-500 max-w-2xl mx-auto">
        基于 MariaDB 的湖仓一体存储，实现 Schema、向量索引、Rich Context 的原生统一管理
      </p>
    </div>

    <!-- Architecture comparison -->
    <NGrid :x-gap="24" :cols="2" class="mb-8">
      <!-- Traditional approach -->
      <NGridItem>
        <NCard title="传统方案" class="h-full">
          <template #header-extra>
            <NTag type="error" size="small">复杂</NTag>
          </template>

          <div class="relative py-8">
            <!-- Components -->
            <div class="flex justify-center gap-8 mb-8">
              <div class="text-center">
                <div class="w-16 h-16 rounded-lg bg-blue-100 dark:bg-blue-900/30 flex items-center justify-center mb-2">
                  <div class="i-logos-mysql text-2xl" />
                </div>
                <span class="text-sm text-gray-600 dark:text-gray-400">MySQL</span>
              </div>
              <div class="text-center">
                <div class="w-16 h-16 rounded-lg bg-purple-100 dark:bg-purple-900/30 flex items-center justify-center mb-2">
                  <div class="i-simple-icons-milvus text-2xl text-purple-500" />
                </div>
                <span class="text-sm text-gray-600 dark:text-gray-400">Milvus</span>
              </div>
              <div class="text-center">
                <div class="w-16 h-16 rounded-lg bg-green-100 dark:bg-green-900/30 flex items-center justify-center mb-2">
                  <div class="i-logos-redis text-2xl" />
                </div>
                <span class="text-sm text-gray-600 dark:text-gray-400">Redis</span>
              </div>
            </div>

            <!-- Sync arrows -->
            <div class="flex justify-center items-center gap-4 mb-8">
              <div class="h-px bg-gray-300 dark:bg-gray-600 w-24" />
              <div class="px-3 py-1 rounded bg-yellow-100 dark:bg-yellow-900/30 text-yellow-700 dark:text-yellow-400 text-sm">
                Sync Agent
              </div>
              <div class="h-px bg-gray-300 dark:bg-gray-600 w-24" />
            </div>

            <!-- Problems -->
            <div class="space-y-2 text-sm">
              <div class="flex items-center gap-2 text-red-500">
                <div class="i-carbon-close-filled" />
                <span>数据同步延迟，一致性难保证</span>
              </div>
              <div class="flex items-center gap-2 text-red-500">
                <div class="i-carbon-close-filled" />
                <span>多组件运维复杂，故障点多</span>
              </div>
              <div class="flex items-center gap-2 text-red-500">
                <div class="i-carbon-close-filled" />
                <span>Schema 变更需要手动同步</span>
              </div>
              <div class="flex items-center gap-2 text-red-500">
                <div class="i-carbon-close-filled" />
                <span>难以实现事务一致性</span>
              </div>
            </div>
          </div>
        </NCard>
      </NGridItem>

      <!-- Lakebase approach -->
      <NGridItem>
        <NCard title="湖仓一体方案" class="h-full border-blue-500">
          <template #header-extra>
            <NTag type="success" size="small">简洁</NTag>
          </template>

          <div class="relative py-8">
            <!-- Single component -->
            <div class="flex justify-center mb-8">
              <div class="text-center">
                <div class="w-24 h-24 rounded-xl bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center mb-2 shadow-lg">
                  <div class="text-white">
                    <div class="i-logos-mariadb-icon text-3xl" />
                  </div>
                </div>
                <span class="text-sm font-medium text-gray-700 dark:text-gray-300">MariaDB</span>
                <p class="text-xs text-gray-500">湖仓一体</p>
              </div>
            </div>

            <!-- Internal components -->
            <div class="flex justify-center gap-4 mb-8">
              <div class="px-3 py-2 rounded-lg bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 text-sm">
                <div class="i-carbon-data-table text-blue-500 mb-1" />
                <span class="text-gray-600 dark:text-gray-400">Schema</span>
              </div>
              <div class="px-3 py-2 rounded-lg bg-purple-50 dark:bg-purple-900/20 border border-purple-200 dark:border-purple-800 text-sm">
                <div class="i-carbon-machine-learning text-purple-500 mb-1" />
                <span class="text-gray-600 dark:text-gray-400">Vector</span>
              </div>
              <div class="px-3 py-2 rounded-lg bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 text-sm">
                <div class="i-carbon-document text-green-500 mb-1" />
                <span class="text-gray-600 dark:text-gray-400">Context</span>
              </div>
            </div>

            <!-- Benefits -->
            <div class="space-y-2 text-sm">
              <div class="flex items-center gap-2 text-green-600 dark:text-green-400">
                <div class="i-carbon-checkmark-filled" />
                <span>原生数据一致性，无需同步</span>
              </div>
              <div class="flex items-center gap-2 text-green-600 dark:text-green-400">
                <div class="i-carbon-checkmark-filled" />
                <span>单一组件，运维简单</span>
              </div>
              <div class="flex items-center gap-2 text-green-600 dark:text-green-400">
                <div class="i-carbon-checkmark-filled" />
                <span>Schema 变更自动感知</span>
              </div>
              <div class="flex items-center gap-2 text-green-600 dark:text-green-400">
                <div class="i-carbon-checkmark-filled" />
                <span>继承企业级高可用能力</span>
              </div>
            </div>
          </div>
        </NCard>
      </NGridItem>
    </NGrid>

    <!-- Core features -->
    <h3 class="text-xl font-semibold text-gray-800 dark:text-gray-100 mb-4">核心技术特性</h3>
    <NGrid :x-gap="16" :y-gap="16" :cols="3">
      <NGridItem>
        <NCard size="small">
          <div class="flex items-start gap-3">
            <div class="w-10 h-10 rounded-lg bg-blue-100 dark:bg-blue-900/30 flex items-center justify-center flex-shrink-0">
              <div class="i-carbon-connect text-blue-500 text-xl" />
            </div>
            <div>
              <h4 class="font-medium text-gray-800 dark:text-gray-100 mb-1">多阶段 Grounding</h4>
              <p class="text-sm text-gray-500">
                表级链接 + 列级链接 + 语义召回，精准识别相关 Schema 元素
              </p>
            </div>
          </div>
        </NCard>
      </NGridItem>

      <NGridItem>
        <NCard size="small">
          <div class="flex items-start gap-3">
            <div class="w-10 h-10 rounded-lg bg-purple-100 dark:bg-purple-900/30 flex items-center justify-center flex-shrink-0">
              <div class="i-carbon-flow text-purple-500 text-xl" />
            </div>
            <div>
              <h4 class="font-medium text-gray-800 dark:text-gray-100 mb-1">ReAct 推理框架</h4>
              <p class="text-sm text-gray-500">
                思考-行动-观察循环，透明可解释的 SQL 生成过程
              </p>
            </div>
          </div>
        </NCard>
      </NGridItem>

      <NGridItem>
        <NCard size="small">
          <div class="flex items-start gap-3">
            <div class="w-10 h-10 rounded-lg bg-green-100 dark:bg-green-900/30 flex items-center justify-center flex-shrink-0">
              <div class="i-carbon-recycle text-green-500 text-xl" />
            </div>
            <div>
              <h4 class="font-medium text-gray-800 dark:text-gray-100 mb-1">Rich Context 自维持</h4>
              <p class="text-sm text-gray-500">
                错误反馈自动修正，Schema 变更自动适应，持续优化
              </p>
            </div>
          </div>
        </NCard>
      </NGridItem>
    </NGrid>
  </div>
</template>
