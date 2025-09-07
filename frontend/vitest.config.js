import { defineConfig } from 'vitest/config';
import path from 'path';

export default defineConfig({
  resolve: {
    alias: {
      '$lib': path.resolve('./src/lib'),
      '$app': path.resolve('./node_modules/@sveltejs/kit/src/runtime/app')
    }
  },
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: ['./src/lib/api/__tests__/setup.js']
  }
});