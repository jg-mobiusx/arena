import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      // Proxy API requests to the Go daemon
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        // Uncomment rewrite if Go serves API from root:
        // rewrite: (path) => path.replace(/^\/api/, '')
      }
    }
  }
})
