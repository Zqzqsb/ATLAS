import { createApp } from 'vue'
import { createPinia } from 'pinia'
import {
  create,
  NMessageProvider,
  NDialogProvider,
  NNotificationProvider,
  NConfigProvider
} from 'naive-ui'
import App from './App.vue'
import router from './router'

// UnoCSS
import 'virtual:uno.css'
import './style.css'

// Create app
const app = createApp(App)

// Pinia
const pinia = createPinia()
app.use(pinia)

// Router
app.use(router)

// Mount
app.mount('#app')
