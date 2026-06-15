<script setup lang="ts">
import { provide } from 'vue'
import { DBX_LAYERS, type ArchNode } from '../../model/architecture'
import { SOURCES } from '../../model/sources'
import ArchLayer from '../../../arch/components/overview/ArchLayer.vue'
import { SOURCE_CATALOG_KEY } from '../../../arch/components/overview/source-catalog'

defineEmits<{ select: [node: ArchNode, ev: MouseEvent] }>()

provide(SOURCE_CATALOG_KEY, SOURCES)
</script>

<template>
  <div class="max-w-5xl mx-auto px-6 py-8">
    <!-- Title -->
    <div class="text-center mb-6">
      <h1 class="text-2xl font-extrabold text-gray-900 tracking-tight">Databricks 全景架构</h1>
      <p class="text-sm text-gray-400 mt-1">Unity Catalog Metric Views + Agent Metadata + Genie · 点击带「展开」的模块查看内部 dataflow</p>
      <div class="flex flex-wrap items-center justify-center gap-1.5 mt-3">
        <span class="px-2.5 py-1 rounded-full text-[11px] font-semibold bg-amber-50 text-amber-700 border border-amber-200">关系 + 指标模型</span>
        <span class="px-2.5 py-1 rounded-full text-[11px] font-semibold bg-indigo-50 text-indigo-700 border border-indigo-200">Metric View = UC 原生对象</span>
        <span class="px-2.5 py-1 rounded-full text-[11px] font-semibold bg-emerald-50 text-emerald-700 border border-emerald-200">行列策略运行时强制</span>
        <span class="px-2.5 py-1 rounded-full text-[11px] font-semibold bg-slate-50 text-slate-600 border border-slate-200">Genie · 关键词召回（非向量）</span>
      </div>
    </div>

    <!-- Layered stack with connectors -->
    <template v-for="(layer, idx) in DBX_LAYERS" :key="layer.id">
      <ArchLayer :layer="layer" @select="(n, ev) => $emit('select', n, ev)" />
      <div v-if="idx < DBX_LAYERS.length - 1" class="flex justify-center py-1.5">
        <div class="flex flex-col items-center text-gray-300">
          <div class="w-px h-3 bg-gray-300" />
          <div class="i-lucide-chevron-down text-sm -my-0.5" />
        </div>
      </div>
    </template>

    <!-- footnote -->
    <p class="text-center text-[11px] text-gray-400 mt-6 leading-relaxed">
      所有事实点都附 <code class="font-mono">[Sn]</code> 角标，悬停查看出处、点击在新标签页打开官方文档。
      证据目录与 <code class="font-mono">WiseCat/.claude/skills/research/results/databricks_uc_metric_views.yaml</code> 一致。
    </p>
  </div>
</template>
