<script setup lang="ts">
import type { PromptSection } from '../../../model/modules'
import SectionHeading from './SectionHeading.vue'

defineProps<{ data: PromptSection }>()
</script>

<template>
  <section>
    <SectionHeading icon="i-lucide-square-terminal" :title="data.title" :subtitle="data.subtitle" accent="amber" />

    <div class="grid grid-cols-1 lg:grid-cols-5 gap-4">
      <!-- Left: prompt building blocks -->
      <div class="lg:col-span-3 rounded-2xl border border-gray-200 bg-white p-4">
        <div class="flex items-center gap-2 mb-3">
          <div class="i-lucide-cpu text-amber-500" />
          <span class="text-xs font-bold text-gray-700">{{ data.engine }}</span>
          <div class="flex items-center gap-1 ml-auto">
            <code
              v-for="t in data.tools"
              :key="t"
              class="px-1.5 py-0.5 rounded bg-amber-50 text-amber-700 border border-amber-200 font-mono text-[10px]"
            >{{ t }}</code>
          </div>
        </div>
        <div class="space-y-2">
          <div
            v-for="(b, i) in data.blocks"
            :key="b.label"
            class="flex items-start gap-2.5 rounded-lg border border-gray-100 bg-gray-50/60 px-3 py-2"
          >
            <div class="w-6 h-6 rounded-md bg-white border border-gray-200 flex-center flex-shrink-0">
              <div :class="[b.icon, 'text-amber-500 text-xs']" />
            </div>
            <div class="flex-1 min-w-0">
              <div class="flex items-center gap-1.5">
                <span class="text-[10px] font-mono text-gray-400">{{ String(i + 1).padStart(2, '0') }}</span>
                <span class="text-xs font-bold text-gray-800">{{ b.label }}</span>
              </div>
              <div class="text-xs text-gray-500 leading-snug">{{ b.desc }}</div>
            </div>
          </div>
        </div>
      </div>

      <!-- Right: rules / constraints -->
      <div class="lg:col-span-2 rounded-2xl border border-amber-200 bg-amber-50/40 p-4">
        <div class="flex items-center gap-2 mb-3">
          <div class="i-lucide-shield-alert text-amber-600" />
          <span class="text-xs font-bold text-amber-700">关键约束 / 技巧</span>
        </div>
        <ul class="space-y-2">
          <li v-for="(r, i) in data.rules" :key="i" class="flex items-start gap-2 text-xs text-gray-700 leading-relaxed">
            <span class="w-4 h-4 rounded-full bg-amber-100 text-amber-700 flex-center text-[9px] font-bold flex-shrink-0 mt-0.5">{{ i + 1 }}</span>
            <span>{{ r }}</span>
          </li>
        </ul>
      </div>
    </div>
  </section>
</template>
