<script setup lang="ts">
import { WREN_LAYERS, type ArchNode } from '../../model/wren'
import ArchLayer from '../../../arch/components/overview/ArchLayer.vue'

const emit = defineEmits<{ select: [node: ArchNode, ev: MouseEvent] }>()
</script>

<template>
  <div class="max-w-5xl mx-auto px-6 py-8">
    <!-- Title -->
    <div class="text-center mb-5">
      <h1 class="text-2xl font-extrabold text-gray-900 tracking-tight">WrenAI 全景架构</h1>
      <p class="text-sm text-gray-400 mt-1">开放上下文层 for agents · 点击带「展开」的模块查看内部 dataflow</p>
    </div>

    <!-- Contrast banner: agentic (ATLAS) vs semantic-layer + primitives (WrenAI) -->
    <div class="grid grid-cols-1 md:grid-cols-2 gap-3 mb-7">
      <div class="rounded-xl border border-emerald-200/70 bg-emerald-50/40 px-4 py-3">
        <div class="flex items-center gap-1.5 mb-1">
          <div class="i-lucide-bot text-emerald-600 text-sm" />
          <span class="text-sm font-bold text-emerald-800">ATLAS · Agentic</span>
        </div>
        <p class="text-[11px] text-emerald-700/90 leading-relaxed">
          Coordinator / Worker ReAct 内核<b>内置</b>，Agent 自动探查数据、生成 Rich Context、端到端跑 NL2SQL 与自维护。
        </p>
      </div>
      <div class="rounded-xl border border-violet-200/70 bg-violet-50/40 px-4 py-3">
        <div class="flex items-center gap-1.5 mb-1">
          <div class="i-lucide-box text-violet-600 text-sm" />
          <span class="text-sm font-bold text-violet-800">WrenAI · Semantic Layer + 原语</span>
        </div>
        <p class="text-[11px] text-violet-700/90 leading-relaxed">
          Agent <b>外置</b>（BYO）。平台提供人工建模的语义契约（MDL）与一组正确性<b>原语</b>（memory / dry-plan / dry-run），由 Agent 自行编排。
        </p>
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
