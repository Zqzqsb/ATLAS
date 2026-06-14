<script setup lang="ts">
import { KTX_LAYERS, type ArchNode } from '../../model/ktx'
import ArchLayer from '../../../arch/components/overview/ArchLayer.vue'

const emit = defineEmits<{ select: [node: ArchNode, ev: MouseEvent] }>()
</script>

<template>
  <div class="max-w-5xl mx-auto px-6 py-8">
    <div class="text-center mb-6">
      <h1 class="text-2xl font-extrabold text-gray-900 tracking-tight">ktx 全景架构</h1>
      <p class="text-sm text-gray-400 mt-1">
        The context layer for data agents · BYO Agent · 点击带「展开」的模块查看内部 dataflow
      </p>
      <div class="flex flex-wrap items-center justify-center gap-1.5 mt-3">
        <span class="px-2.5 py-1 rounded-full text-[11px] font-semibold bg-slate-100 text-slate-600 border border-slate-200">
          BYO Agent · MCP-first
        </span>
        <span class="px-2.5 py-1 rounded-full text-[11px] font-semibold bg-emerald-50 text-emerald-700 border border-emerald-200">
          Wiki + 语义层双载体
        </span>
        <span class="px-2.5 py-1 rounded-full text-[11px] font-semibold bg-amber-50 text-amber-700 border border-amber-200">
          Git 化 + 凭据分离
        </span>
        <span class="px-2.5 py-1 rounded-full text-[11px] font-semibold bg-blue-50 text-blue-700 border border-blue-200">
          FTS5 + Embeddings 混合检索
        </span>
        <span class="px-2.5 py-1 rounded-full text-[11px] font-semibold bg-violet-50 text-violet-700 border border-violet-200">
          read-only by design
        </span>
      </div>
    </div>

    <template v-for="(layer, idx) in KTX_LAYERS" :key="layer.id">
      <ArchLayer :layer="layer" @select="(n, ev) => emit('select', n, ev)" />
      <div v-if="idx < KTX_LAYERS.length - 1" class="flex justify-center py-1.5">
        <div class="flex flex-col items-center text-gray-300">
          <div class="w-px h-3 bg-gray-300" />
          <div class="i-lucide-chevron-down text-sm -my-0.5" />
        </div>
      </div>
    </template>

    <p class="text-center text-[11px] text-gray-400 mt-6 leading-relaxed">
      基于 Kaelio/ktx (Apache-2.0) 本地 checkout · pnpm + uv workspace ·
      <code class="font-mono">packages/cli</code> (TS) +
      <code class="font-mono">python/ktx-sl</code> · <code class="font-mono">python/ktx-daemon</code> (Python) ·
      <code class="font-mono">.ktx/db.sqlite</code> 索引 · git-backed 项目目录
    </p>
  </div>
</template>
