<script setup lang="ts">
import { ACCENTS } from '../../../model/architecture'
import type { StrategySection } from '../../../model/modules'
import SectionHeading from './SectionHeading.vue'

defineProps<{ data: StrategySection }>()
</script>

<template>
  <section>
    <SectionHeading icon="i-lucide-split" :title="data.title" :subtitle="data.subtitle" accent="violet" />

    <!-- decision pill -->
    <div class="flex justify-center mb-4">
      <div class="inline-flex items-center gap-2 px-3.5 py-1.5 rounded-full bg-violet-50 border border-violet-200 text-xs font-semibold text-violet-700">
        <div class="i-lucide-git-commit-horizontal" />
        {{ data.decision }}
      </div>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <div
        v-for="opt in data.options"
        :key="opt.id"
        class="rounded-2xl border bg-white p-4"
        :class="ACCENTS[opt.accent].surface"
      >
        <div class="flex items-center gap-2.5 mb-3">
          <div class="w-9 h-9 rounded-xl flex-center text-white bg-gradient-to-br" :class="ACCENTS[opt.accent].gradient">
            <div :class="[opt.icon, 'text-base']" />
          </div>
          <div>
            <div class="text-sm font-bold text-gray-900 leading-tight">{{ opt.label }}</div>
            <div class="text-xs font-semibold" :class="ACCENTS[opt.accent].text">{{ opt.when }}</div>
          </div>
        </div>
        <ul class="space-y-1.5">
          <li v-for="(p, i) in opt.points" :key="i" class="flex items-start gap-2 text-xs text-gray-600 leading-relaxed">
            <div class="i-lucide-check mt-0.5 flex-shrink-0" :class="ACCENTS[opt.accent].text" />
            <span>{{ p }}</span>
          </li>
        </ul>
      </div>
    </div>
  </section>
</template>
