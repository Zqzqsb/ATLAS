<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import lucidLogo from '@/assets/lucid-logo.svg'
import LakebaseViz from './LakebaseViz.vue'
import SchemaLinkingViz from './SchemaLinkingViz.vue'
import ContextLifecycleViz from './ContextLifecycleViz.vue'
import AgentSelfMaintainViz from './AgentSelfMaintainViz.vue'

const router = useRouter()

const sections = [
  { id: 'hero', label: 'LUCID' },
  { id: 'lakebase', label: 'Lakebase' },
  { id: 'linking', label: 'Linking' },
  { id: 'lifecycle', label: 'Context' },
  { id: 'selfmaintain', label: 'Agent' },
  { id: 'cta', label: 'Explore' },
]

const currentSection = ref(0)
const containerRef = ref<HTMLElement>()

// Track current section via IntersectionObserver
let observers: IntersectionObserver[] = []

function setupObservers() {
  observers.forEach(o => o.disconnect())
  observers = []

  const sectionEls = containerRef.value?.querySelectorAll<HTMLElement>('.snap-section')
  if (!sectionEls) return

  sectionEls.forEach((el, idx) => {
    const obs = new IntersectionObserver(
      (entries) => {
        const entry = entries[0]
        if (entry && entry.isIntersecting) {
          currentSection.value = idx
        }
      },
      {
        root: containerRef.value,
        threshold: 0.6,
      }
    )
    obs.observe(el)
    observers.push(obs)
  })
}

function scrollToSection(idx: number) {
  const sectionEls = containerRef.value?.querySelectorAll<HTMLElement>('.snap-section')
  if (!sectionEls || !sectionEls[idx]) return
  sectionEls[idx]!.scrollIntoView({ behavior: 'smooth' })
}

function goHome() {
  router.push('/')
}

onMounted(async () => {
  // Lock body scroll while showcase is active
  document.documentElement.classList.add('showcase-active')
  document.body.classList.add('showcase-active')
  await nextTick()
  setupObservers()
})

onUnmounted(() => {
  observers.forEach(o => o.disconnect())
  // Restore body scroll
  document.documentElement.classList.remove('showcase-active')
  document.body.classList.remove('showcase-active')
})
</script>

<template>
  <div
    ref="containerRef"
    class="snap-container"
  >
    <!-- ====== Side navigation dots ====== -->
    <div class="fixed right-5 top-1/2 -translate-y-1/2 z-50 flex flex-col items-center gap-3">
      <button
        v-for="(sec, idx) in sections"
        :key="sec.id"
        class="group relative flex items-center"
        @click="scrollToSection(idx)"
      >
        <!-- Dot -->
        <div
          class="w-2.5 h-2.5 rounded-full border-2 transition-all duration-300"
          :class="currentSection === idx
            ? 'bg-gray-800 border-gray-800 scale-125'
            : 'bg-transparent border-gray-300 hover:border-gray-500 hover:scale-110'"
        />
        <!-- Label tooltip -->
        <span
          class="absolute right-6 whitespace-nowrap text-xs font-medium px-2 py-1 rounded-md bg-gray-800 text-white opacity-0 group-hover:opacity-100 pointer-events-none transition-opacity duration-200"
        >
          {{ sec.label }}
        </span>
      </button>
    </div>

    <!-- ====== Slide 0: Hero ====== -->
    <section class="snap-section relative flex items-center justify-center bg-gradient-to-br from-slate-50 via-white to-blue-50/40">
      <!-- Background blobs -->
      <div class="absolute inset-0 overflow-hidden pointer-events-none">
        <div class="absolute -top-40 -right-40 w-[700px] h-[700px] rounded-full bg-gradient-to-br from-blue-100/50 to-violet-100/40 blur-3xl" />
        <div class="absolute -bottom-40 -left-40 w-[600px] h-[600px] rounded-full bg-gradient-to-br from-emerald-100/40 to-cyan-100/30 blur-3xl" />
      </div>

      <div class="relative max-w-4xl mx-auto px-6 text-center">
        <div class="flex items-center justify-center mb-6">
          <img :src="lucidLogo" alt="LUCID" class="w-16 h-16 rounded-xl shadow-xl shadow-primary-500/20" />
        </div>

        <h1 class="text-5xl md:text-6xl font-extrabold text-gray-900 tracking-tight mb-5">
          <span class="bg-gradient-to-r from-blue-600 via-violet-600 to-emerald-600 bg-clip-text text-transparent">
            Four Innovations
          </span>
          <br>
          <span class="text-gray-800 text-3xl md:text-4xl font-bold">that Power LUCID</span>
        </h1>

        <p class="text-lg text-gray-500 max-w-2xl mx-auto mb-10">
          A Lakebase-Unified Context-aware Intelligence system for Text-to-SQL,
          built entirely inside MariaDB with native VECTOR + HNSW support.
        </p>

        <!-- Nav pills -->
        <div class="inline-flex flex-wrap items-center justify-center gap-3 mb-12">
          <button
            v-for="(item, idx) in [
              { label: 'Lakebase Storage', color: 'bg-blue-100 text-blue-700 hover:bg-blue-200', section: 1 },
              { label: 'Schema Linking', color: 'bg-violet-100 text-violet-700 hover:bg-violet-200', section: 2 },
              { label: 'Context Lifecycle', color: 'bg-emerald-100 text-emerald-700 hover:bg-emerald-200', section: 3 },
              { label: 'Self-Maintaining', color: 'bg-amber-100 text-amber-700 hover:bg-amber-200', section: 4 },
            ]"
            :key="idx"
            class="px-4 py-2 rounded-full text-sm font-semibold transition-all duration-200"
            :class="item.color"
            @click="scrollToSection(item.section)"
          >
            {{ item.label }}
          </button>
        </div>

        <!-- Scroll hint -->
        <div class="animate-bounce text-gray-400">
          <div class="i-lucide-chevrons-down text-2xl mx-auto" />
        </div>
      </div>
    </section>

    <!-- ====== Slide 1: Lakebase — warm blue tint ====== -->
    <section class="snap-section relative flex items-center bg-gradient-to-br from-blue-50/70 via-white to-indigo-50/40">
      <div class="max-w-6xl mx-auto px-6 py-12 w-full">
        <LakebaseViz />
      </div>
    </section>

    <!-- ====== Slide 2: Schema Linking — soft violet tint ====== -->
    <section class="snap-section relative flex items-center bg-gradient-to-br from-violet-50/60 via-white to-purple-50/30">
      <div class="max-w-6xl mx-auto px-6 py-12 w-full">
        <SchemaLinkingViz />
      </div>
    </section>

    <!-- ====== Slide 3: Context Lifecycle — subtle green tint ====== -->
    <section class="snap-section relative flex items-center bg-gradient-to-br from-emerald-50/50 via-white to-teal-50/30">
      <div class="max-w-6xl mx-auto px-6 py-12 w-full">
        <ContextLifecycleViz />
      </div>
    </section>

    <!-- ====== Slide 4: Agent Self-Maintain — warm amber tint ====== -->
    <section class="snap-section relative flex items-center bg-gradient-to-br from-amber-50/50 via-white to-orange-50/30">
      <div class="max-w-6xl mx-auto px-6 py-12 w-full">
        <AgentSelfMaintainViz />
      </div>
    </section>

    <!-- ====== Slide 5: CTA Footer ====== -->
    <section class="snap-section relative flex items-center justify-center bg-gradient-to-r from-gray-900 via-gray-800 to-gray-900">
      <div class="max-w-4xl mx-auto px-6 text-center">
        <h2 class="text-4xl font-bold text-white mb-4">Ready to Explore?</h2>
        <p class="text-gray-400 mb-10 text-lg">
          Connect your database and experience LUCID's intelligent Text-to-SQL pipeline
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
            href="https://github.com/zqzqsb/lucid"
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
.snap-container {
  height: calc(100vh - 56px);
  height: calc(100dvh - 56px);
  overflow-y: auto;
  scroll-snap-type: y mandatory;
  -webkit-overflow-scrolling: touch;
}

.snap-section {
  scroll-snap-align: start;
  scroll-snap-stop: always;
  min-height: calc(100vh - 56px);
  min-height: calc(100dvh - 56px);
}
</style>

<style>
/* Applied dynamically via JS classList */
html.showcase-active,
html.showcase-active body {
  overflow: hidden !important;
}
</style>
