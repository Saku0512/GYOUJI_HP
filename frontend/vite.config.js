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
    sourcemap: false,
    // アセット最適化
    assetsInlineLimit: 4096, // 4KB以下のアセットはインライン化
    // ロールアップオプション
    rollupOptions: {
      output: {
        // チャンク分割戦略
        manualChunks: {
          // ベンダーライブラリを分離
          vendor: ['svelte'],
          // APIクライアントを分離
          api: ['./src/lib/api/client.js', './src/lib/api/auth.js', './src/lib/api/tournament.js', './src/lib/api/matches.js'],
          // ユーティリティを分離
          utils: ['./src/lib/utils/imageOptimization.js', './src/lib/utils/assetCache.js', './src/lib/utils/performanceMonitor.js']
        },
        // アセットファイル名の設定
        assetFileNames: (assetInfo) => {
          const info = assetInfo.name.split('.');
          const ext = info[info.length - 1];
          
          // 画像ファイル
          if (/png|jpe?g|svg|gif|tiff|bmp|ico/i.test(ext)) {
            return `assets/images/[name]-[hash][extname]`;
          }
          
          // フォントファイル
          if (/woff2?|eot|ttf|otf/i.test(ext)) {
            return `assets/fonts/[name]-[hash][extname]`;
          }
          
          // CSSファイル
          if (ext === 'css') {
            return `assets/css/[name]-[hash][extname]`;
          }
          
          // その他のアセット
          return `assets/[name]-[hash][extname]`;
        },
        // JSファイル名の設定
        chunkFileNames: 'assets/js/[name]-[hash].js',
        entryFileNames: 'assets/js/[name]-[hash].js'
      }
    }
  },
  // 依存関係の事前バンドル最適化
  optimizeDeps: {
    include: ['svelte'],
    exclude: []
  },
  // CSS最適化
  css: {
    devSourcemap: true
  },
  // アセット処理の設定
  assetsInclude: ['**/*.webp', '**/*.avif'],
  // 実験的機能
  experimental: {
    // レンダリングの最適化
    renderBuiltUrl(filename, { hostType }) {
      if (hostType === 'js') {
        // JSファイルからの参照時はCDNを使用（本番環境）
        if (process.env.NODE_ENV === 'production' && process.env.VITE_CDN_URL) {
          return `${process.env.VITE_CDN_URL}/${filename}`;
        }
      }
      return { relative: true };
    }
  }
});
