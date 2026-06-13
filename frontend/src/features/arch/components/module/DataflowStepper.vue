<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { ACCENTS } from '../../model/architecture'
import type { FlowDef } from '../../model/flows'

const props = defineProps<{ flow: FlowDef }>()

const active = ref(0)
const playing = ref(true)
const STEP_MS = 2600
let timer: number | null = null

const steps = computed(() => props.flow.steps)
const current = computed(() => steps.value[active.value])
const a = computed(() => ACCENTS[current.value?.accent ?? props.flow.accent])

function clear() {
  if (timer) { clearInterval(timer); timer = null }
}
function play() {
  clear()
  playing.value = true
  timer = window.setInterval(() => {
    if (active.value >= steps.value.length - 1) {
      playing.value = false
      clear()
    } else {
      active.value++
    }
  }, STEP_MS)
}
function pause() { playing.value = false; clear() }
function toggle() { playing.value ? pause() : (active.value >= steps.value.length - 1 ? replay() : play()) }
function goto(i: number) { pause(); active.value = i }
function replay() { active.value = 0; play() }

// Restart whenever the flow changes (e.g. drilling into a different module)
watch(() => props.flow.id, () => { active.value = 0; play() })

onMounted(() => { play() })
onUnmounted(clear)
</script>

<template>
  <div class="grid grid-cols-1 lg:grid-cols-5 gap-6">
    <!-- Left rail: step list -->
    <div class="lg:col-span-2 space-y-2">
      <button
        v-for="(step, idx) in steps"
        :key="step.id"
        class="w-full text-left flex items-start gap-3 rounded-xl border px-3 py-2.5 transition-all duration-300"
        :class="active === idx
          ? `${ACCENTS[step.accent].surface} shadow-sm`
          : active > idx
            ? 'border-gray-200 bg-gray-50/60'
            : 'border-gray-200 bg-white hover:border-gray-300'"
        @click="goto(idx)"
      >
        <div class="flex flex-col items-center gap-1 flex-shrink-0">
          <div
            class="w-8 h-8 rounded-lg flex-center text-white bg-gradient-to-br transition-all duration-300"
            :class="active >= idx ? ACCENTS[step.accent].gradient : 'from-gray-300 to-gray-400'"
          >
            <div v-if="active > idx" class="i-lucide-check text-base" />
            <div v-else :class="[step.icon, 'text-base', active === idx ? 'animate-pulse' : '']" />
          </div>
          <span class="text-[10px] font-mono text-gray-400">{{ String(idx + 1).padStart(2, '0') }}</span>
        </div>
        <div class="flex-1 min-w-0 pt-0.5">
          <div class="text-sm font-bold leading-tight" :class="active === idx ? 'text-gray-900' : 'text-gray-600'">
            {{ step.title }}
          </div>
          <div class="text-xs text-gray-400 mt-0.5">{{ step.summary }}</div>
        </div>
        <!-- connector -->
        <div
          v-if="idx < steps.length - 1"
          class="absolute"
        />
      </button>
    </div>

    <!-- Right: active step detail -->
    <div class="lg:col-span-3">
      <Transition name="flow-detail" mode="out-in">
        <div :key="current?.id" class="rounded-2xl border bg-white p-5 h-full" :class="a.surface">
          <div class="flex items-center gap-3 mb-3">
            <div class="w-10 h-10 rounded-xl flex-center text-white bg-gradient-to-br" :class="a.gradient">
              <div :class="[current?.icon, 'text-lg']" />
            </div>
            <div>
              <div class="text-base font-bold text-gray-900 leading-tight">{{ current?.title }}</div>
              <div class="text-xs" :class="a.text">{{ current?.subtitle }}</div>
            </div>
          </div>

          <p class="text-sm text-gray-600 leading-relaxed mb-4">{{ current?.detail }}</p>

          <!-- IO row -->
          <div class="flex items-stretch gap-2 mb-4">
            <div class="flex-1 rounded-lg bg-gray-50 border border-gray-100 px-3 py-2">
              <div class="text-[10px] uppercase tracking-wide text-gray-400 font-semibold mb-0.5">Input</div>
              <div class="text-xs text-gray-700 font-medium">{{ current?.artifact.input ?? '—' }}</div>
            </div>
            <div class="flex-center text-gray-300">
              <div class="i-lucide-arrow-right" />
            </div>
            <div class="flex-1 rounded-lg border px-3 py-2" :class="a.chip">
              <div class="text-[10px] uppercase tracking-wide font-semibold mb-0.5 opacity-70">Output</div>
              <div class="text-xs font-medium">{{ current?.artifact.output ?? '—' }}</div>
            </div>
          </div>

          <!-- store tag -->
          <div v-if="current?.artifact.store" class="flex items-center gap-1.5 mb-3 text-xs">
            <div class="i-lucide-hard-drive text-gray-400" />
            <span class="text-gray-400">写入 / 调用</span>
            <code class="px-1.5 py-0.5 rounded bg-gray-100 text-gray-700 font-mono text-[11px]">{{ current?.artifact.store }}</code>
          </div>

          <!-- code snippet -->
          <div v-if="current?.artifact.code" class="rounded-xl bg-gray-900 overflow-hidden">
            <div class="flex items-center gap-1.5 px-3 py-1.5 border-b border-white/10">
              <div class="w-2 h-2 rounded-full bg-red-400/70" />
              <div class="w-2 h-2 rounded-full bg-amber-400/70" />
              <div class="w-2 h-2 rounded-full bg-emerald-400/70" />
              <span class="ml-1.5 text-[10px] font-mono text-gray-400 uppercase">{{ current?.artifact.lang ?? 'text' }}</span>
            </div>
            <pre class="px-3.5 py-3 text-[12px] leading-relaxed text-gray-100 font-mono overflow-x-auto whitespace-pre">{{ current?.artifact.code }}</pre>
          </div>
        </div>
      </Transition>

      <!-- playback controls -->
      <div class="flex items-center justify-between mt-3 px-1">
        <div class="flex items-center gap-1.5">
          <button class="w-8 h-8 rounded-lg flex-center text-gray-500 hover:bg-gray-100 disabled:opacity-30" :disabled="active === 0" @click="goto(active - 1)">
            <div class="i-lucide-chevron-left" />
          </button>
          <button class="w-9 h-9 rounded-lg flex-center text-white bg-gradient-to-br" :class="a.gradient" @click="toggle">
            <div :class="playing ? 'i-lucide-pause' : (active >= steps.length - 1 ? 'i-lucide-rotate-ccw' : 'i-lucide-play')" />
          </button>
          <button class="w-8 h-8 rounded-lg flex-center text-gray-500 hover:bg-gray-100 disabled:opacity-30" :disabled="active >= steps.length - 1" @click="goto(active + 1)">
            <div class="i-lucide-chevron-right" />
          </button>
        </div>
        <!-- progress dots -->
        <div class="flex items-center gap-1.5">
          <button
            v-for="(s, i) in steps"
            :key="s.id"
            class="h-1.5 rounded-full transition-all duration-300"
            :class="i === active ? `w-5 ${a.bar}` : 'w-1.5 bg-gray-200 hover:bg-gray-300'"
            @click="goto(i)"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.flow-detail-enter-active,
.flow-detail-leave-active {
  transition: opacity 0.25s ease, transform 0.25s ease;
}
.flow-detail-enter-from {
  opacity: 0;
  transform: translateY(8px);
}
.flow-detail-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}
</style>
