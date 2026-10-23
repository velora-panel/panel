import { resolve } from 'node:path'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'
import devtools from 'vite-plugin-vue-devtools'
export default defineConfig({
  plugins: [vue(), tailwindcss(), devtools()],
 resolve: {alias:{'@':resolve(import.meta.dirname,'src')}},
  server: {
    fs: {
      deny: [
        '**/.env',
        '**/.env.*',
        '**/*.{crt,pem}',
        '**/.git/**',
        '**/assets_temp/**',
        resolve(import.meta.dirname, '../../assets/**'),
      ],
    },
    proxy: { '/api': process.env.CORE_PROXY ?? 'http://127.0.0.1:8080' },
  },
})
