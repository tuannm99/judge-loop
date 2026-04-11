import { defineConfig } from 'vite'
import { qwikVite } from '@builder.io/qwik/optimizer'
import tailwindcss from '@tailwindcss/vite'
import path from 'path'

export default defineConfig({
  plugins: [
    tailwindcss(),
    qwikVite({
      csr: true,
      srcDir: 'src/qwik'
    })
  ],
  resolve: {
    alias: {
      '@builder.io/qwik-city': path.resolve(
        __dirname,
        './src/qwik/shims/qwik-city.tsx'
      ),
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: process.env.VITE_API_PROXY_TARGET ?? 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})
