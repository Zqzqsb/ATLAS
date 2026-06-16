<script setup lang="ts">
import { OKF_LAYERS, type ArchNode } from '../../model/okf'
import ArchLayer from '../../../arch/components/overview/ArchLayer.vue'

const emit = defineEmits<{ select: [node: ArchNode, ev: MouseEvent] }>()
</script>

<template>
  <div class="max-w-5xl mx-auto px-6 py-8">
    <div class="text-center mb-6">
      <h1 class="text-2xl font-extrabold text-gray-900 tracking-tight">OKF · Open Knowledge Format</h1>
      <p class="text-sm text-gray-400 mt-1">
        GoogleCloudPlatform/knowledge-catalog · vendor-neutral 的元数据目录格式
      </p>
      <p class="text-[12px] text-gray-500 mt-1.5 leading-snug max-w-3xl mx-auto">
        把"知识"落成 <code class="font-mono text-rose-600 font-semibold">markdown 文件 + YAML frontmatter</code> 的目录树。
        任何 agent 都能写、任何 UI 都能读、git 化 / 可分发的元数据目录。
      </p>
      <div class="flex flex-wrap items-center justify-center gap-1.5 mt-3">
        <span class="px-2.5 py-1 rounded-full text-[11px] font-semibold bg-slate-100 text-slate-600 border border-slate-200">
          vendor-neutral
        </span>
        <span class="px-2.5 py-1 rounded-full text-[11px] font-semibold bg-violet-50 text-violet-700 border border-violet-200">
          Google ADK + Gemini
        </span>
        <span class="px-2.5 py-1 rounded-full text-[11px] font-semibold bg-emerald-50 text-emerald-700 border border-emerald-200">
          Git 化 / 文件即资产
        </span>
        <span class="px-2.5 py-1 rounded-full text-[11px] font-semibold bg-amber-50 text-amber-700 border border-amber-200">
          extension-first
        </span>
        <span class="px-2.5 py-1 rounded-full text-[11px] font-semibold bg-blue-50 text-blue-700 border border-blue-200">
          human + agent 友好
        </span>
      </div>
    </div>

    <template v-for="(layer, idx) in OKF_LAYERS" :key="layer.id">
      <ArchLayer :layer="layer" @select="(n, ev) => emit('select', n, ev)" />
      <div v-if="idx < OKF_LAYERS.length - 1" class="flex justify-center py-1.5">
        <div class="flex flex-col items-center text-gray-300">
          <div class="w-px h-3 bg-gray-300" />
          <div class="i-lucide-chevron-down text-sm -my-0.5" />
        </div>
      </div>
    </template>

    <p class="text-center text-[11px] text-gray-400 mt-6 leading-relaxed">
      基于 GoogleCloudPlatform/knowledge-catalog (Apache-2.0) 本地 checkout ·
      <code class="font-mono">okf/SPEC.md</code> 规范本体 ·
      <code class="font-mono">okf/src/enrichment_agent/</code> ADK agent 实现 ·
      <code class="font-mono">bundles/</code> 三个真实 bundle (ga4 / stackoverflow / crypto_bitcoin)
    </p>
  </div>
</template>
