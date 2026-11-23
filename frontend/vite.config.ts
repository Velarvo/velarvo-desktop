import path from 'node:path'
import { svelte } from '@sveltejs/vite-plugin-svelte'
import tailwindcss from '@tailwindcss/vite'
import wails from '@wailsio/runtime/plugins/vite'
import { defineConfig } from 'vite'

export default defineConfig({
  server: {
    host: '127.0.0.1',
    port: Number(process.env.WAILS_VITE_PORT) || 9245,
    strictPort: true,
  },
  plugins: [svelte(), tailwindcss(), wails('./bindings')],
  resolve: {
    conditions: ['browser'],
    alias: {
      '@': path.resolve(import.meta.dirname, './src'),
      '@bindings': path.resolve(
        import.meta.dirname,
        './bindings/github.com/Velarvo/velarvo-desktop/internal',
      ),
      '@events': path.resolve(import.meta.dirname, './bindings/events'),
    },
  },
})
