<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import lucidLogo from '@/assets/lucid-logo.svg'
import LakebaseViz from './LakebaseViz.vue'
import SchemaLinkingViz from './SchemaLinkingViz.vue'
import ContextLifecycleViz from './ContextLifecycleViz.vue'
import AgentSelfMaintainViz from './AgentSelfMaintainViz.vue'

const router = useRouter()

// Scroll progress tracking
const scrollProgress = ref(0)
const showBackToTop = ref(false)

function handleScroll() {
  const scrollTop = window.scrollY
  const docHeight = document.documentElement.scrollHeight - window.innerHeight
  scrollProgress.value = docHeight > 0 ? (scrollTop / docHeight) * 100 : 0
  showBackToTop.value = scrollTop > 400
}

function scrollToTop() {
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

function goHome() {
  router.push('/')
}

onMounted(() => {
  window.addEventListener('scroll', handleScroll, { passive: true })
})

onUnmounted(() => {
  window.removeEventListener('scroll', handleScroll)
})
</script>

<template>
  <div class="min-h-screen bg-gradient-to-br from-slate-50 via-white to-blue-50/30">
    <!-- Scroll progress bar -->
    <div class="fixed top-14 left-0 right-0 z-40 h-0.5 bg-gray-100">
      <div
        class="h-full bg-gradient-to-r from-blue-500 via-violet-500 to-emerald-500 transition-all duration-100"
        :style="{ width: `${scrollProgress}%` }"
      />
    </div>

    <!-- Hero section -->
    <div class="relative overflow-hidden">
      <!-- Background blobs -->
      <div class="absolute inset-0 overflow-hidden pointer-events-none">
        <div class="absolute -top-40 -right-40 w-[700px] h-[700px] rounded-full bg-gradient-to-br from-blue-100/50 to-violet-100/40 blur-3xl" />
        <div class="absolute -bottom-40 -left-40 w-[600px] h-[600px] rounded-full bg-gradient-to-br from-emerald-100/40 to-cyan-100/30 blur-3xl" />
      </div>

      <div class="relative max-w-6xl mx-auto px-6 pt-20 pb-16 text-center">
        <!-- Logo -->
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

        <p class="text-lg text-gray-500 max-w-2xl mx-auto mb-8">
          A Lakebase-Unified Context-aware Intelligence system for Text-to-SQL, 
          built entirely inside MariaDB 12 with native VECTOR + HNSW support.
        </p>

        <!-- Quick nav pills -->
        <div class="inline-flex flex-wrap items-center justify-center gap-3">
          <a
            v-for="(item, idx) in [
              { label: 'Lakebase Storage', color: 'bg-blue-100 text-blue-700 hover:bg-blue-200', hash: '#lakebase' },
              { label: 'Schema Linking', color: 'bg-violet-100 text-violet-700 hover:bg-violet-200', hash: '#linking' },
              { label: 'Context Lifecycle', color: 'bg-emerald-100 text-emerald-700 hover:bg-emerald-200', hash: '#lifecycle' },
              { label: 'Self-Maintaining', color: 'bg-amber-100 text-amber-700 hover:bg-amber-200', hash: '#selfmaintain' },
            ]"
            :key="idx"
            :href="item.hash"
            class="px-4 py-2 rounded-full text-sm font-semibold transition-all duration-200"
            :class="item.color"
          >
            {{ item.label }}
          </a>
        </div>
      </div>
    </div>

    <!-- Feature Sections -->
    <div class="max-w-6xl mx-auto px-6 space-y-32 pb-24">
      <!-- Section 1: Lakebase -->
      <section id="lakebase">
        <LakebaseViz />
      </section>

      <!-- Section 2: Schema Linking -->
      <section id="linking">
        <SchemaLinkingViz />
      </section>

      <!-- Section 3: Context Lifecycle -->
      <section id="lifecycle">
        <ContextLifecycleViz />
      </section>

      <!-- Section 4: Self-Maintaining -->
      <section id="selfmaintain">
        <AgentSelfMaintainViz />
      </section>
    </div>

    <!-- Footer CTA -->
    <div class="bg-gradient-to-r from-gray-900 via-gray-800 to-gray-900 py-16">
      <div class="max-w-4xl mx-auto px-6 text-center">
        <h2 class="text-3xl font-bold text-white mb-4">Ready to Explore?</h2>
        <p class="text-gray-400 mb-8 text-lg">
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
    </div>

    <!-- Back to top -->
    <Transition name="fade">
      <button
        v-if="showBackToTop"
        class="fixed bottom-8 right-8 w-12 h-12 rounded-xl bg-white border border-gray-200 shadow-xl flex-center text-gray-600 hover:text-primary-600 hover:-translate-y-1 transition-all z-50"
        @click="scrollToTop"
      >
        <div class="i-lucide-chevron-up text-xl" />
      </button>
    </Transition>
  </div>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>

<style>
/* Smooth scroll for anchor navigation (global) */
html {
  scroll-behavior: smooth;
}
</style>
