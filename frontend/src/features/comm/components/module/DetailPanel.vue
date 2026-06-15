<script setup lang="ts">
/**
 * DetailPanel — collapsible "summary + bullets" block, styled like PeekPanel
 * (rounded amber-accented card, chevron toggles, smooth expand).
 */
import { ref } from 'vue'
import type { VendorDetail, AccentKey } from '../../model/comm'
import { ACCENTS } from '../../model/comm'
import InlineCode from './InlineCode.vue'

const props = defineProps<{
  detail: VendorDetail
  /** primary accent for the trigger ring + bullet bullets */
  accent?: AccentKey
  /** start expanded? defaults to true so dense detail is visible by default */
  defaultOpen?: boolean
}>()

const open = ref(props.defaultOpen ?? true)
function toggle() {
  open.value = !open.value
}
</script>

<template>
  <div class="rounded-xl border bg-white overflow-hidden" :class="ACCENTS[accent ?? 'amber'].surface">
    <!-- ─── trigger ─── -->
    <button
      type="button"
      class="w-full flex items-start gap-2.5 px-3 py-2.5 text-left transition-colors"
      :class="open ? 'border-b' : ''"
      @click="toggle"
    >
      <div
        class="w-5 h-5 rounded-md flex-center flex-shrink-0 mt-0.5"
        :class="[ACCENTS[accent ?? 'amber'].iconBg, ACCENTS[accent ?? 'amber'].iconText]"
      >
        <div class="i-lucide-info text-[12px]" />
      </div>
      <div class="flex-1 min-w-0">
        <div class="flex items-baseline gap-1.5 flex-wrap">
          <span class="text-[10px] font-bold tracking-wider text-gray-500">DETAIL · 怎么做</span>
          <span class="text-[10px] font-mono text-gray-400">{{ detail.bullets.length }} 点</span>
        </div>
        <p class="text-[12px] text-gray-700 leading-relaxed mt-0.5"><InlineCode :text="detail.summary" /></p>
      </div>
      <div
        class="i-lucide-chevron-down text-gray-400 text-sm flex-shrink-0 mt-1 transition-transform"
        :class="{ 'rotate-180': open }"
      />
    </button>

    <!-- ─── body ─── -->
    <Transition
      enter-active-class="transition-all duration-200 ease-out overflow-hidden"
      leave-active-class="transition-all duration-150 ease-in overflow-hidden"
      enter-from-class="opacity-0 max-h-0"
      enter-to-class="opacity-100 max-h-[1200px]"
      leave-from-class="opacity-100 max-h-[1200px]"
      leave-to-class="opacity-0 max-h-0"
    >
      <div v-show="open" class="px-3 py-2.5 bg-white">
        <ul class="space-y-2">
          <li
            v-for="(b, i) in detail.bullets"
            :key="i"
            class="flex items-start gap-2.5"
          >
            <div
              class="w-5 h-5 rounded-md flex-center flex-shrink-0 mt-0.5 text-[10px] font-bold"
              :class="[
                ACCENTS[b.accent ?? accent ?? 'amber'].iconBg,
                ACCENTS[b.accent ?? accent ?? 'amber'].iconText,
              ]"
            >
              <div v-if="b.icon" :class="[b.icon, 'text-[11px]']" />
              <span v-else>{{ i + 1 }}</span>
            </div>
            <div class="flex-1 min-w-0">
              <div class="flex items-baseline gap-1.5 flex-wrap">
                <span
                  class="text-[12px] font-bold leading-tight"
                  :class="ACCENTS[b.accent ?? accent ?? 'amber'].text"
                >{{ b.label }}</span>
              </div>
              <p class="text-[12px] text-gray-700 leading-relaxed mt-0.5 whitespace-pre-line"><InlineCode :text="b.body" /></p>
            </div>
          </li>
        </ul>
        <div
          v-if="detail.closing"
          class="mt-2.5 pt-2 border-t border-dashed border-gray-200 flex items-start gap-1.5"
        >
          <div class="i-lucide-quote text-amber-500 text-[12px] flex-shrink-0 mt-0.5" />
          <p class="text-[12px] text-gray-700 leading-relaxed font-medium italic"><InlineCode :text="detail.closing" /></p>
        </div>
      </div>
    </Transition>
  </div>
</template>
