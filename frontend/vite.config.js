import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
  plugins: [sveltekit()],
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true
      }
    }
  },
  build: {
    // 圧縮設定
    minify: 'terser',
    terserOptions: {
      compress: {
        drop_console: true, // 本番環境でconsole.logを削除
        drop_debugger: true
      }
    },
    // チャンクサイズ警告の閾値を設定
    chunkSizeWarningLimit: 1000,
    // ソースマップを本番環境では無効化
    sourcemap: false
  },
  // 依存関係の事前バンドル最適化
  optimizeDeps: {
    include: ['svelte'],
    exclude: []
  },
  // CSS最適化
  css: {
    devSourcemap: true
  }
});
