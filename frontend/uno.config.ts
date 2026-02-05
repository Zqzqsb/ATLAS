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
      }
    })
  ],
  transformers: [
    transformerDirectives(),
    transformerVariantGroup()
  ],
  shortcuts: {
    'btn': 'px-4 py-2 rounded-lg font-bold transition-all cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed text-sm tracking-wide',
    'btn-primary': 'btn bg-primary-600 text-white hover:bg-primary-700 active:bg-primary-800 shadow-sm hover:shadow',
    'btn-secondary': 'btn bg-white text-gray-700 border border-gray-200 hover:bg-gray-50 hover:border-gray-300 shadow-sm',
    'btn-success': 'btn bg-green-600 text-white hover:bg-green-700 shadow-sm',
    'card': 'bg-white rounded-xl border border-gray-200 shadow-sm transition-all',
    'card-hover': 'card hover:shadow-md hover:border-gray-300',
    'flex-center': 'flex items-center justify-center',
    'flex-between': 'flex items-center justify-between',
  },
  theme: {
    colors: {
      primary: {
        50: '#f0f5ff',
        100: '#e0ebff',
        200: '#c2d6ff',
        300: '#94b5ff',
        400: '#5e8aff',
        500: '#3b66f5', // Academic Blue
        600: '#254adb',
        700: '#1d3bbf',
        800: '#1e329b',
        900: '#1e2d7a',
        950: '#141b4b',
      }
    }
  }
})
