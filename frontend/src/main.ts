import '@/i18n'
import '@fontsource/inter/400.css'
import '@fontsource/jetbrains-mono/400.css'
import './index.css'
import App from './App.svelte'
import { mount } from 'svelte'
import { installGlobalLoggerBridge, logger } from '@/lib/logger'

installGlobalLoggerBridge()

const target = document.getElementById('root')

if (!target) {
  logger.error('Root element not found')
  throw new Error('Root element not found')
}

const app = mount(App, {
  target,
})

export default app
