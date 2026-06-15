<script setup lang="ts">
import { provide } from 'vue'
import { SNOW_LAYERS, type ArchNode } from '../../model/architecture'
import { SOURCES } from '../../model/sources'
import ArchLayer from '../../../arch/components/overview/ArchLayer.vue'
import { SOURCE_CATALOG_KEY } from '../../../arch/components/overview/source-catalog'

defineEmits<{ select: [node: ArchNode, ev: MouseEvent] }>()

provide(SOURCE_CATALOG_KEY, SOURCES)
</script>

<template>
  <div class="max-w-5xl mx-auto px-6 py-8">
    <div class="text-center mb-6">
      <h1 class="text-2xl font-extrabold text-gray-900 tracking-tight">Snowflake 全景架构</h1>
      <p class="text-sm text-gray-400 mt-1">Cortex Analyst + Semantic View + VQR + Cortex Search · 点击带「展开」的模块查看内部 dataflow</p>
      <div class="flex flex-wrap items-center justify-center gap-1.5 mt-3">
        <span class="px-2.5 py-1 rounded-full text-[11px] font-semibold bg-amber-50 text-amber-700 border border-amber-200">Semantic View · DDL 原生对象</span>
        <span class="px-2.5 py-1 rounded-full text-[11px] font-semibold bg-violet-50 text-violet-700 border border-violet-200">Verified Query Repository</span>
        <span class="px-2.5 py-1 rounded-full text-[11px] font-semibold bg-blue-50 text-blue-700 border border-blue-200">Cortex Search · 混合检索</span>
        <span class="px-2.5 py-1 rounded-full text-[11px] font-semibold bg-emerald-50 text-emerald-700 border border-emerald-200">API 不执行 · 调用方运行</span>
      </div>
    </div>

    <template v-for="(layer, idx) in SNOW_LAYERS" :key="layer.id">
      <ArchLayer :layer="layer" @select="(n, ev) => $emit('select', n, ev)" />
      <div v-if="idx < SNOW_LAYERS.length - 1" class="flex justify-center py-1.5">
        <div class="flex flex-col items-center text-gray-300">
          <div class="w-px h-3 bg-gray-300" />
          <div class="i-lucide-chevron-down text-sm -my-0.5" />
        </div>
      </div>
    </template>

    <p class="text-center text-[11px] text-gray-400 mt-6 leading-relaxed">
      所有事实点都附 <code class="font-mono">[Sn]</code> 角标，悬停查看出处、点击在新标签页打开官方文档。
      证据目录与 <code class="font-mono">WiseCat/.claude/skills/research/results/snowflake_cortex_analyst_semantic_views.yaml</code> 一致。
    </p>
  </div>
</template>
