<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import atlasLogo from '@/assets/atlas-logo.svg'
import LakebaseViz from './LakebaseViz.vue'
import SchemaLinkingViz from './SchemaLinkingViz.vue'
import ContextLifecycleViz from './ContextLifecycleViz.vue'
import AgentSelfMaintainViz from './AgentSelfMaintainViz.vue'

const router = useRouter()

/* ─── Section metadata ─── */
const sections = [
{ id: 'hero', label: 'ATLAS' },
  { id: 'lakebase', label: 'Lakebase' },
  { id: 'linking', label: 'Linking' },
  { id: 'lifecycle', label: 'Context' },
  { id: 'selfmaintain', label: 'Agent' },
  { id: 'cta', label: 'Explore' },
]

/* ─── Reactive state ─── */
const current = ref(0)
const isTransitioning = ref(false)
const direction = ref<'up' | 'down'>('down')
const containerRef = ref<HTMLElement>()
// Tracks which slides have been revealed — once revealed, content stays visible permanently
const revealed = ref<Set<number>>(new Set([0]))
// The slide currently playing its entrance animation (or -1 if none)
const enteringSlide = ref(-1)

const TOTAL = sections.length
const DURATION = 800        // ms per transition
const COOLDOWN = 80         // ms after transition ends before accepting new input
const WHEEL_THRESHOLD = 30  // px delta to trigger page change
const TOUCH_THRESHOLD = 50  // px swipe distance to trigger

/* ─── Easing: custom smooth-stop curve (no final-frame discontinuity) ─── */
function easeOutQuart(t: number): number {
  return 1 - Math.pow(1 - t, 4)
}

/* ─── Core navigation ─── */
let animFrame = 0
let cooldownTimer = 0

function goTo(idx: number) {
  if (idx < 0 || idx >= TOTAL || idx === current.value || isTransitioning.value) return

  direction.value = idx > current.value ? 'down' : 'up'
  isTransitioning.value = true

  const container = containerRef.value
  if (!container) return

  const slideHeight = container.offsetHeight
  const startY = current.value * slideHeight
  const endY = idx * slideHeight
  const startTime = performance.now()

  cancelAnimationFrame(animFrame)

  let hasSwapped = false

  function animate(now: number) {
    const elapsed = now - startTime
    const progress = Math.min(elapsed / DURATION, 1)
    const eased = easeOutQuart(progress)

    container!.scrollTop = startY + (endY - startY) * eased

    // Swap active index early (at ~55%) so content entrance overlaps with scroll
    if (!hasSwapped && progress >= 0.55) {
      hasSwapped = true
      current.value = idx
      if (!revealed.value.has(idx)) {
        enteringSlide.value = idx
        // After animation completes, mark as permanently revealed
        setTimeout(() => {
          revealed.value.add(idx)
          enteringSlide.value = -1
        }, 500)
      }
    }

    if (progress < 1) {
      animFrame = requestAnimationFrame(animate)
    } else {
      // Ensure final position (curve naturally lands, no visible jump)
      container!.scrollTop = endY
      if (!hasSwapped) {
        current.value = idx
        if (!revealed.value.has(idx)) {
          enteringSlide.value = idx
          setTimeout(() => {
            revealed.value.add(idx)
            enteringSlide.value = -1
          }, 500)
        }
      }
      cooldownTimer = window.setTimeout(() => {
        isTransitioning.value = false
      }, COOLDOWN)
    }
  }

  animFrame = requestAnimationFrame(animate)
}

function next() { goTo(current.value + 1) }
function prev() { goTo(current.value - 1) }

/* ─── Wheel handler (debounced, direction-aware) ─── */
let wheelAccum = 0
let wheelTimer = 0

function onWheel(e: WheelEvent) {
  e.preventDefault()
  if (isTransitioning.value) return

  wheelAccum += e.deltaY
  clearTimeout(wheelTimer)

  wheelTimer = window.setTimeout(() => { wheelAccum = 0 }, 200)

  if (Math.abs(wheelAccum) >= WHEEL_THRESHOLD) {
    if (wheelAccum > 0) next()
    else prev()
    wheelAccum = 0
  }
}

/* ─── Touch handler ─── */
let touchStartY = 0
let touchStartTime = 0

function onTouchStart(e: TouchEvent) {
  const t = e.touches[0]
  if (!t) return
  touchStartY = t.clientY
  touchStartTime = Date.now()
}

function onTouchEnd(e: TouchEvent) {
  if (isTransitioning.value) return
  const t = e.changedTouches[0]
  if (!t) return
  const dy = touchStartY - t.clientY
  const dt = Date.now() - touchStartTime
  // Require minimum distance OR fast flick
  if (Math.abs(dy) >= TOUCH_THRESHOLD || (Math.abs(dy) > 20 && dt < 300)) {
    if (dy > 0) next()
    else prev()
  }
}

/* ─── Keyboard handler ─── */
function onKeyDown(e: KeyboardEvent) {
  if (isTransitioning.value) return
  switch (e.key) {
    case 'ArrowDown':
    case 'PageDown':
    case ' ':
      e.preventDefault()
      next()
      break
    case 'ArrowUp':
    case 'PageUp':
      e.preventDefault()
      prev()
      break
    case 'Home':
      e.preventDefault()
      goTo(0)
      break
    case 'End':
      e.preventDefault()
      goTo(TOTAL - 1)
      break
  }
}

/* ─── Progress indicator ─── */
const progress = computed(() => ((current.value) / (TOTAL - 1)) * 100)

/* ─── Lifecycle ─── */
onMounted(() => {
  document.documentElement.classList.add('showcase-active')
  document.body.classList.add('showcase-active')

  const el = containerRef.value
  if (el) {
    el.addEventListener('wheel', onWheel, { passive: false })
    el.addEventListener('touchstart', onTouchStart, { passive: true })
    el.addEventListener('touchend', onTouchEnd, { passive: true })
  }
  window.addEventListener('keydown', onKeyDown)
})

onUnmounted(() => {
  document.documentElement.classList.remove('showcase-active')
  document.body.classList.remove('showcase-active')

  const el = containerRef.value
  if (el) {
    el.removeEventListener('wheel', onWheel)
    el.removeEventListener('touchstart', onTouchStart)
    el.removeEventListener('touchend', onTouchEnd)
  }
  window.removeEventListener('keydown', onKeyDown)
  cancelAnimationFrame(animFrame)
  clearTimeout(cooldownTimer)
  clearTimeout(wheelTimer)
})

function goHome() {
  router.push('/')
}
</script>

<template>
  <div
    ref="containerRef"
    class="slide-viewport"
  >
    <!-- ─── Progress bar ─── -->
    <div class="fixed top-[56px] left-0 right-0 z-50 h-[2px] bg-gray-200/60">
      <div
        class="h-full bg-gradient-to-r from-blue-500 via-violet-500 to-emerald-500 transition-all ease-out"
        :style="{ width: `${progress}%`, transitionDuration: `${DURATION}ms` }"
      />
    </div>

    <!-- ─── Side nav dots ─── -->
    <nav class="fixed right-5 top-1/2 -translate-y-1/2 z-50 flex flex-col items-center gap-3">
      <button
        v-for="(sec, idx) in sections"
        :key="sec.id"
        class="group relative flex items-center"
        :aria-label="sec.label"
        @click="goTo(idx)"
      >
        <!-- Track line between dots -->
        <div
          v-if="idx < sections.length - 1"
          class="absolute left-1/2 -translate-x-1/2 top-full w-px h-3 transition-colors duration-300"
          :class="idx < current ? 'bg-gray-700' : 'bg-gray-200'"
        />
        <!-- Dot -->
        <div class="relative">
          <div
            class="w-2.5 h-2.5 rounded-full border-2 transition-all duration-500"
            :class="current === idx
              ? 'bg-gray-800 border-gray-800 scale-[1.4]'
              : idx < current
                ? 'bg-gray-400 border-gray-400 scale-100'
                : 'bg-transparent border-gray-300 hover:border-gray-500 hover:scale-110'"
          />
          <!-- Active ring pulse -->
          <div
            v-if="current === idx"
            class="absolute inset-0 rounded-full border-2 border-gray-800 animate-ping opacity-20"
          />
        </div>
        <!-- Label tooltip -->
        <span
          class="absolute right-7 whitespace-nowrap text-xs font-medium px-2.5 py-1 rounded-lg bg-gray-900/90 text-white opacity-0 group-hover:opacity-100 pointer-events-none transition-all duration-200 backdrop-blur-sm -translate-x-1 group-hover:translate-x-0"
        >
          {{ sec.label }}
        </span>
      </button>
    </nav>

    <!-- ─── Section counter ─── -->
    <div class="fixed left-5 bottom-5 z-50 text-xs font-mono text-gray-400 select-none tracking-widest">
      <span class="text-gray-800 font-bold text-sm">{{ String(current + 1).padStart(2, '0') }}</span>
      <span class="mx-1">/</span>
      <span>{{ String(TOTAL).padStart(2, '0') }}</span>
    </div>

    <!-- ====== Slide 0: Hero ====== -->
    <section
      class="slide-page flex items-center justify-center bg-gradient-to-br from-slate-50 via-white to-blue-50/40"
      :class="{ 'is-entering': enteringSlide === 0, 'is-visible': revealed.has(0) }"
    >
      <div class="absolute inset-0 overflow-hidden pointer-events-none">
        <div class="absolute -top-40 -right-40 w-[700px] h-[700px] rounded-full bg-gradient-to-br from-blue-100/50 to-violet-100/40 blur-3xl" />
        <div class="absolute -bottom-40 -left-40 w-[600px] h-[600px] rounded-full bg-gradient-to-br from-emerald-100/40 to-cyan-100/30 blur-3xl" />
      </div>

      <div class="relative max-w-4xl mx-auto px-6 text-center">
        <div class="flex items-center justify-center mb-6">
<img :src="atlasLogo" alt="ATLAS" class="w-16 h-16 rounded-xl shadow-xl shadow-primary-500/20" />
        </div>

        <h1 class="text-5xl md:text-6xl font-extrabold text-gray-900 tracking-tight mb-5">
          <span class="bg-gradient-to-r from-blue-600 via-violet-600 to-emerald-600 bg-clip-text text-transparent">
            Four Innovations
          </span>
          <br>
<span class="text-gray-800 text-3xl md:text-4xl font-bold">that Power ATLAS</span>
        </h1>

        <p class="text-lg text-gray-500 max-w-2xl mx-auto mb-10">
          A lifecycle-aware self-maintaining Text-to-SQL system,
          built entirely inside MariaDB with native VECTOR + HNSW support.
        </p>

        <div class="inline-flex flex-wrap items-center justify-center gap-3 mb-12">
          <button
            v-for="(item, idx) in [
              { label: 'Lakebase Storage', color: 'bg-blue-100 text-blue-700 hover:bg-blue-200', section: 1 },
              { label: 'Schema Linking', color: 'bg-violet-100 text-violet-700 hover:bg-violet-200', section: 2 },
              { label: 'Context Lifecycle', color: 'bg-emerald-100 text-emerald-700 hover:bg-emerald-200', section: 3 },
              { label: 'Self-Maintaining', color: 'bg-amber-100 text-amber-700 hover:bg-amber-200', section: 4 },
            ]"
            :key="idx"
            class="px-4 py-2 rounded-full text-sm font-semibold transition-all duration-200 hover:-translate-y-0.5"
            :class="item.color"
            @click="goTo(item.section)"
          >
            {{ item.label }}
          </button>
        </div>

        <div class="text-gray-400 cursor-pointer hover:text-gray-600 transition-colors" @click="next()">
          <div class="animate-bounce">
            <div class="i-lucide-chevrons-down text-2xl mx-auto" />
          </div>
          <div class="text-xs mt-1 tracking-wider uppercase">Scroll or press ↓</div>
        </div>
      </div>
    </section>

    <!-- ====== Slide 1: Lakebase ====== -->
    <section
      class="slide-page flex items-center bg-gradient-to-br from-blue-50/70 via-white to-indigo-50/40"
      :class="{ 'is-entering': enteringSlide === 1, 'is-visible': revealed.has(1) }"
    >
      <div class="max-w-6xl mx-auto px-6 py-12 w-full">
        <LakebaseViz />
      </div>
    </section>

    <!-- ====== Slide 2: Schema Linking ====== -->
    <section
      class="slide-page flex items-center bg-gradient-to-br from-violet-50/60 via-white to-purple-50/30"
      :class="{ 'is-entering': enteringSlide === 2, 'is-visible': revealed.has(2) }"
    >
      <div class="max-w-6xl mx-auto px-6 py-12 w-full">
        <SchemaLinkingViz />
      </div>
    </section>

    <!-- ====== Slide 3: Context Lifecycle ====== -->
    <section
      class="slide-page flex items-center bg-gradient-to-br from-emerald-50/50 via-white to-teal-50/30"
      :class="{ 'is-entering': enteringSlide === 3, 'is-visible': revealed.has(3) }"
    >
      <div class="max-w-6xl mx-auto px-6 py-12 w-full">
        <ContextLifecycleViz />
      </div>
    </section>

    <!-- ====== Slide 4: Agent Self-Maintain ====== -->
    <section
      class="slide-page flex items-center bg-gradient-to-br from-amber-50/50 via-white to-orange-50/30"
      :class="{ 'is-entering': enteringSlide === 4, 'is-visible': revealed.has(4) }"
    >
      <div class="max-w-6xl mx-auto px-6 py-6 w-full">
        <AgentSelfMaintainViz />
      </div>
    </section>

    <!-- ====== Slide 5: CTA ====== -->
    <section
      class="slide-page flex items-center justify-center bg-gradient-to-r from-gray-900 via-gray-800 to-gray-900"
      :class="{ 'is-entering': enteringSlide === 5, 'is-visible': revealed.has(5) }"
    >
      <div class="max-w-4xl mx-auto px-6 text-center">
        <h2 class="text-4xl font-bold text-white mb-4">Ready to Explore?</h2>
        <p class="text-gray-400 mb-10 text-lg">
          Connect your database and experience ATLAS's intelligent Text-to-SQL pipeline
        </p>
        <div class="flex items-center justify-center gap-4">
          <button
            class="px-6 py-3 rounded-xl bg-gradient-to-r from-blue-500 to-violet-500 text-white font-bold shadow-lg shadow-blue-500/30 hover:shadow-xl hover:-translate-y-0.5 transition-all"
            @click="goHome"
          >
            <div class="flex items-center gap-2">
              <div class="i-lucide-database" />
              Go to Databases
            </div>
          </button>
          <a
            href="https://github.com/zqzqsb/atlas"
            target="_blank"
            class="px-6 py-3 rounded-xl border border-gray-600 text-gray-300 font-bold hover:bg-gray-700/50 hover:-translate-y-0.5 transition-all"
          >
            <div class="flex items-center gap-2">
              <div class="i-lucide-github" />
              GitHub
            </div>
          </a>
        </div>
      </div>
    </section>
  </div>
</template>

<style scoped>
/* ─── Viewport: no CSS scroll-snap, JS controls everything ─── */
.slide-viewport {
  height: calc(100vh - 56px);
  height: calc(100dvh - 56px);
  overflow: hidden;                 /* JS handles scrollTop */
  position: relative;
}

/* ─── Each slide ─── */
.slide-page {
  width: 100%;
  height: calc(100vh - 56px);
  height: calc(100dvh - 56px);
  position: relative;
}

/* Content hidden until visited */
.slide-page > div {
  opacity: 0;
}

/* Once visited, content stays fully visible — no flicker on re-scroll */
.slide-page.is-visible > div {
  opacity: 1;
}

/* First-time entrance: gentle fade-up (only fires once per slide) */
.slide-page.is-entering > div {
  animation: slide-fade-in 500ms cubic-bezier(0.22, 1, 0.36, 1) forwards;
}

@keyframes slide-fade-in {
  from {
    opacity: 0;
    transform: translateY(16px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>

<style>
html.showcase-active,
html.showcase-active body {
  overflow: hidden !important;
}
</style>
