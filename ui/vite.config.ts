import { fileURLToPath, URL } from 'node:url'

import tailwindcss from '@tailwindcss/vite'
import { tanstackRouter } from '@tanstack/router-plugin/vite'
import react from '@vitejs/plugin-react'
import { defineConfig, lazyPlugins } from 'vite-plus'

// O frontend fala exclusivamente com o main (API gateway).
// Em dev, /api/* é repassado para o main; no Docker o nginx faz o mesmo.
const mainApiUrl = process.env.MAIN_API_URL ?? 'http://localhost:8080'

export default defineConfig({
  plugins: lazyPlugins(() => [
    tanstackRouter({ target: 'react', autoCodeSplitting: true }),
    react(),
    tailwindcss(),
  ]),
  resolve: {
    alias: { '@': fileURLToPath(new URL('./src', import.meta.url)) },
  },
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: mainApiUrl,
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, ''),
      },
    },
  },
  fmt: {
    singleQuote: true,
    semi: false,
    ignorePatterns: ['src/routeTree.gen.ts'],
  },
  lint: {
    plugins: ['react', 'typescript', 'oxc'],
    ignorePatterns: ['dist/**', 'src/routeTree.gen.ts', 'src/components/ui/**'],
    rules: {
      'react/rules-of-hooks': 'error',
      'react/only-export-components': ['warn', { allowConstantExport: true }],
      'vite-plus/prefer-vite-plus-imports': 'error',
    },
    overrides: [
      {
        // Arquivos de rota exportam `Route`; o code-splitting do TanStack Router cuida do HMR.
        files: ['src/routes/**'],
        rules: { 'react/only-export-components': 'off' },
      },
    ],
    options: {
      typeAware: true,
      typeCheck: true,
    },
    jsPlugins: [{ name: 'vite-plus', specifier: 'vite-plus/oxlint-plugin' }],
  },
})
