import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, '.', '')

  // the browser only ever talks to vite, which passes /api/* on to the go server.
  // under WSL this matters: a windows browser reaches WSL ports through localhost
  // forwarding, which quietly loses to any windows program already on the same port
  const proxy = {
    '/api': {
      target: env.API_PROXY_TARGET || 'http://localhost:8080',
      rewrite: (path: string) => path.replace(/^\/api/, ''),
    },
  }

  return {
    plugins: [react()],
    server: { proxy },
    preview: { proxy },
  }
})
