<script setup lang="ts">
import { computed } from 'vue'
import { COMM_LAYERS, type ArchNode, type ArchLayer } from '../../model/comm'
import { ACCENTS } from '../../../arch/model/architecture'

const emit = defineEmits<{ select: [node: ArchNode, ev: MouseEvent] }>()

function nodeAccentClasses(node: ArchNode) {
  const a = ACCENTS[node.accent]
  return `${a.surface} ${node.flow ? a.hover + ' cursor-pointer' : ''}`
}

function layerAccentClasses(layer: ArchLayer) {
  return ACCENTS[layer.accent]
}

function gridColsClass(cols: number) {
  return cols === 4
    ? 'grid-cols-1 sm:grid-cols-2 lg:grid-cols-4'
    : cols === 3
    ? 'grid-cols-1 sm:grid-cols-3'
    : 'grid-cols-1 sm:grid-cols-2'
}

function onNodeClick(node: ArchNode, ev: MouseEvent) {
  if (!node.flow) return
  emit('select', node, ev)
}
</script>

<template>
  <div class="max-w-7xl mx-auto px-6 py-8">
    <!-- title -->
    <div class="text-center mb-6">
      <h1 class="text-2xl font-extrabold text-gray-900 tracking-tight">Context Layer · 通用构建框架</h1>
      <p class="text-sm text-gray-500 mt-1.5 leading-snug max-w-3xl mx-auto">
        不针对单一产品，而是把"NL2SQL / 数据 Agent"系统拆成 6 个一般性环节。
        每个环节里给出我们的 <code class="text-rose-600 font-semibold">Common Sense</code>，并展示各家在这一环节的不同取舍。
      </p>
      <div class="flex flex-wrap items-center justify-center gap-1.5 mt-3">
        <span class="px-2.5 py-1 rounded-full text-[11px] font-semibold bg-violet-50 text-violet-700 border border-violet-200">
          Agentic
        </span>
        <span class="px-2.5 py-1 rounded-full text-[11px] font-semibold bg-amber-50 text-amber-700 border border-amber-200">
          Semantic Layer
        </span>
        <span class="px-2.5 py-1 rounded-full text-[11px] font-semibold bg-blue-50 text-blue-700 border border-blue-200">
          Managed Cloud
        </span>
        <span class="px-2.5 py-1 rounded-full text-[11px] font-semibold bg-emerald-50 text-emerald-700 border border-emerald-200">
          Open Context
        </span>
      </div>
    </div>

    <!-- 6-stage horizontal pipeline header -->
    <div class="mb-4 px-4 py-2.5 rounded-xl bg-gradient-to-r from-slate-50 via-emerald-50/40 to-indigo-50/40 border border-slate-200">
      <div class="flex items-center justify-between text-[11px] font-semibold text-gray-600">
        <span>① 入口</span>
        <div class="i-lucide-arrow-right text-gray-300" />
        <span>② 上下文</span>
        <div class="i-lucide-arrow-right text-gray-300" />
        <span>③ 推理</span>
        <div class="i-lucide-arrow-right text-gray-300" />
        <span>④ 校验</span>
        <div class="i-lucide-arrow-right text-gray-300" />
        <span>⑤ 反馈</span>
        <div class="i-lucide-arrow-right text-gray-300" />
        <span>⑥ 记忆</span>
      </div>
    </div>

    <!-- Layered stack -->
    <template v-for="(layer, idx) in COMM_LAYERS" :key="layer.id">
      <div
        class="relative rounded-2xl border px-5 py-4 overflow-hidden shadow-sm"
        :class="layerAccentClasses(layer).surface"
      >
        <!-- accent left rail -->
        <div class="absolute left-0 top-0 bottom-0 w-1.5 bg-gradient-to-b" :class="layerAccentClasses(layer).gradient" />
        <!-- layer header -->
        <div class="flex items-baseline gap-2 mb-3 pl-1.5">
          <div
            class="w-8 h-8 rounded-lg flex-center flex-shrink-0 bg-gradient-to-br text-white shadow-sm"
            :class="layerAccentClasses(layer).gradient"
          >
            <div :class="[layer.icon, 'text-base text-white']" />
          </div>
          <div class="flex-1 min-w-0">
            <div class="text-[15px] font-extrabold text-gray-900 leading-tight">{{ layer.title }}</div>
            <div v-if="layer.subtitle" class="text-[11.5px] text-gray-500 leading-snug">{{ layer.subtitle }}</div>
          </div>
          <span class="text-[10px] font-bold tracking-wider px-2 py-0.5 rounded" :class="layerAccentClasses(layer).chip">
            {{ layer.nodes.length }} 子环节 · 点击展开
          </span>
        </div>

        <!-- nodes -->
        <div class="grid gap-2 pl-1.5" :class="gridColsClass(layer.cols)">
          <button
            v-for="node in layer.nodes"
            :key="node.id"
            type="button"
            class="relative rounded-xl border bg-white/70 backdrop-blur-sm pl-4 pr-3 py-2.5 text-left transition-all hover:shadow-md overflow-hidden"
            :class="nodeAccentClasses(node)"
            @click="(ev) => onNodeClick(node, ev)"
          >
            <div class="absolute left-0 top-0 bottom-0 w-1" :class="ACCENTS[node.accent].dot" />
            <div class="flex items-center gap-1.5 mb-0.5">
              <div
                class="w-5 h-5 rounded-md flex-center flex-shrink-0"
                :class="ACCENTS[node.accent].iconBg"
              >
                <div :class="[node.icon, ACCENTS[node.accent].iconText, 'text-[12px]']" />
              </div>
              <span class="text-[12.5px] font-bold text-gray-800 leading-tight">{{ node.label }}</span>
              <div v-if="node.flow" class="i-lucide-chevron-right text-xs ml-auto" :class="ACCENTS[node.accent].text" />
            </div>
            <div v-if="node.sublabel" class="text-[10.5px] text-gray-500 leading-snug pl-6">{{ node.sublabel }}</div>
          </button>
        </div>
      </div>

      <!-- connector arrow -->
      <div v-if="idx < COMM_LAYERS.length - 1" class="flex justify-center py-2">
        <div class="flex flex-col items-center text-gray-300">
          <div class="w-px h-4 bg-gray-300" />
          <div class="i-lucide-chevron-down text-base -my-0.5" />
        </div>
      </div>
    </template>

    <p class="text-center text-[11px] text-gray-400 mt-6 leading-relaxed">
      点击任一子环节进入"该环节的设计取舍 + 各家做法对照 + Common Sense"。
      6 个环节是一条完整的问答时间线（入口 → 上下文 → 推理 → 校验 → 反馈 → 记忆）。
    </p>
  </div>
</template>
