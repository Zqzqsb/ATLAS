<script setup lang="ts">
import { WREN_LAYERS, type ArchNode } from '../../model/wren'
import ArchLayer from '../../../arch/components/overview/ArchLayer.vue'

const emit = defineEmits<{ select: [node: ArchNode, ev: MouseEvent] }>()
</script>

<template>
  <div class="max-w-5xl mx-auto px-6 py-8">
    <!-- Title -->
    <div class="text-center mb-6">
      <h1 class="text-2xl font-extrabold text-gray-900 tracking-tight">WrenAI 全景架构</h1>
      <p class="text-sm text-gray-400 mt-1">开放上下文层 for agents · 点击带「展开」的模块查看内部 dataflow</p>
      <div class="flex flex-wrap items-center justify-center gap-1.5 mt-3">
        <span class="px-2.5 py-1 rounded-full text-[11px] font-semibold bg-slate-100 text-slate-600 border border-slate-200">BYO Agent · 不内置 NL2SQL LLM</span>
        <span class="px-2.5 py-1 rounded-full text-[11px] font-semibold bg-emerald-50 text-emerald-700 border border-emerald-200">人工建模语义契约 MDL</span>
        <span class="px-2.5 py-1 rounded-full text-[11px] font-semibold bg-amber-50 text-amber-700 border border-amber-200">正确性即原语</span>
        <span class="px-2.5 py-1 rounded-full text-[11px] font-semibold bg-blue-50 text-blue-700 border border-blue-200">LanceDB 记忆 / few-shot</span>
      </div>
    </div>

    <!-- Layered stack with connectors -->
    <template v-for="(layer, idx) in WREN_LAYERS" :key="layer.id">
      <ArchLayer :layer="layer" @select="(n, ev) => emit('select', n, ev)" />
      <div v-if="idx < WREN_LAYERS.length - 1" class="flex justify-center py-1.5">
        <div class="flex flex-col items-center text-gray-300">
          <div class="w-px h-3 bg-gray-300" />
          <div class="i-lucide-chevron-down text-sm -my-0.5" />
        </div>
      </div>
    </template>

    <!-- footnote -->
    <p class="text-center text-[11px] text-gray-400 mt-6 leading-relaxed">
      基于本地 WrenAI checkout（2026-05 合并后新布局 · <code class="font-mono">core/ + sdk/</code>）与 docs.getwren.ai。
      Legacy v1（Haystack 管道 + Qdrant + Wren UI）已迁至 <code class="font-mono">legacy/v1</code> 分支。
    </p>
  </div>
</template>
