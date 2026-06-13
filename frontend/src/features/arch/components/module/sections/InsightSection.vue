<script setup lang="ts">
import type { InsightSection } from '../../../model/modules'
import SectionHeading from './SectionHeading.vue'

defineProps<{ data: InsightSection }>()
</script>

<template>
  <section>
    <SectionHeading icon="i-lucide-lightbulb" :title="data.title" :subtitle="data.subtitle" accent="emerald" />

    <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
      <div
        v-for="(item, i) in data.items"
        :key="i"
        class="rounded-xl border border-gray-200 bg-gradient-to-br from-white to-gray-50/60 p-4 hover:border-emerald-200 hover:shadow-sm transition-all"
      >
        <div class="flex items-center gap-2 mb-1.5">
          <div class="w-7 h-7 rounded-lg bg-emerald-50 flex-center">
            <div :class="[item.icon, 'text-emerald-600 text-sm']" />
          </div>
          <span class="text-sm font-bold text-gray-900">{{ item.title }}</span>
        </div>
        <p class="text-xs text-gray-600 leading-relaxed">{{ item.body }}</p>
      </div>
    </div>
  </section>
</template>
