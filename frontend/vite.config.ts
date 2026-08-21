import path from 'path'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import { defineConfig } from 'vite'
import nodeCrypto, { webcrypto } from 'node:crypto'

const cryptoModule = nodeCrypto as unknown as {
  getRandomValues?: (typedArray: Uint8Array) => Uint8Array
}

if (!cryptoModule.getRandomValues) {
  cryptoModule.getRandomValues = (typedArray: Uint8Array) => webcrypto.getRandomValues(typedArray)
}

if (!globalThis.crypto?.getRandomValues) {
  globalThis.crypto = webcrypto
}

export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    port: 3005,
    open: true,
  },
})
