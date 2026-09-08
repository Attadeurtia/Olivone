import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'

// En dev, `npm run dev` sert le frontend sur :5173 et proxifie /api vers le
// serveur Go (:8791). En prod, `npm run build` produit ./dist, servi par Go.
export default defineConfig({
  plugins: [svelte()],
  build: {
    outDir: 'dist',
    emptyOutDir: true,
  },
  server: {
    port: 5173,
    proxy: {
      '/api': 'http://localhost:8791',
    },
  },
})
