import {
  defineConfig,
  presetUno,
  presetAttributify,
  presetIcons,
  transformerDirectives,
  transformerVariantGroup
} from 'unocss'

export default defineConfig({
  presets: [
    presetUno(),
    presetAttributify(),
    presetIcons({
      scale: 1.2,
      extraProperties: {
        'display': 'inline-block',
        'vertical-align': 'middle'
      },
      // pnpm's strict node_modules breaks presetIcons' auto-discovery of
      // @iconify-json/* collections, so no icon CSS was generated at all.
      // Provide the lucide collection explicitly (the only set we use).
      collections: {
        lucide: () => import('@iconify-json/lucide/icons.json').then(i => i.default),
      },
    })
  ],
  transformers: [
    transformerDirectives(),
    transformerVariantGroup()
  ],
  content: {
    pipeline: {
      // Default scans .vue/.tsx/etc only. Also scan plain .ts/.js so utility &
      // icon class names that live in data files (model/*.ts: ARCH_LAYERS,
      // flows, MODULES — icon names, ACCENTS) get generated. include REPLACES
      // the default, so keep the default patterns too.
      include: [
        /\.(vue|svelte|[jt]sx|mdx?|astro|elm|php|phtml|html)($|\?)/,
        /\.[jt]s($|\?)/,
      ],
    },
  },
  shortcuts: {
    'btn': 'px-3.5 py-2 rounded-lg font-medium transition-colors cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed text-sm',
    'btn-primary': 'btn bg-primary-600 text-white hover:bg-primary-700 active:bg-primary-800',
    'btn-secondary': 'btn bg-white text-gray-700 border border-gray-200 hover:bg-gray-50 hover:border-gray-300',
    'btn-success': 'btn bg-emerald-600 text-white hover:bg-emerald-700',
    'btn-ghost': 'btn text-gray-600 hover:bg-gray-100 hover:text-gray-900',
    'card': 'bg-white rounded-lg border border-gray-200 transition-colors',
    'card-hover': 'card hover:border-gray-300',
    'flex-center': 'flex items-center justify-center',
    'flex-between': 'flex items-center justify-between',
  },
  theme: {
    colors: {
      primary: {
        50: '#f0f4ff',
        100: '#dbe4ff',
        200: '#bac8ff',
        300: '#91a7ff',
        400: '#748ffc',
        500: '#5c7cfa',
        600: '#4c6ef5',
        700: '#4263eb',
        800: '#3b5bdb',
        900: '#364fc7',
        950: '#1e2a5e',
      }
    }
  }
})
