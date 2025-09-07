import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';
import { visualizer } from 'rollup-plugin-visualizer';

export default defineConfig({
  plugins: [
    sveltekit(),
    // バンドル分析用のビジュアライザー
    visualizer({
      filename: 'dist/bundle-analysis.html',
      open: true,
      gzipSize: true,
      brotliSize: true,
      template: 'treemap' // 'treemap', 'sunburst', 'network'
    })
  ],
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true
      }
    }
  },
  build: {
    // 分析用の詳細設定
    rollupOptions: {
      output: {
        // より詳細なチャンク分割
        manualChunks: {
          // フレームワーク関連
          'svelte-core': ['svelte'],
          'sveltekit-core': ['@sveltejs/kit'],
          
          // API関連
          'api-auth': ['./src/lib/api/auth.js'],
          'api-tournament': ['./src/lib/api/tournament.js'],
          'api-matches': ['./src/lib/api/matches.js'],
          'api-client': ['./src/lib/api/client.js'],
          
          // ストア関連
          'store-auth': ['./src/lib/stores/auth.js'],
          'store-tournament': ['./src/lib/stores/tournament.js'],
          'store-ui': ['./src/lib/stores/ui.js'],
          
          // ユーティリティ関連
          'utils-validation': ['./src/lib/utils/validation.js'],
          'utils-formatting': ['./src/lib/utils/formatting.js'],
          'utils-storage': ['./src/lib/utils/storage.js'],
          'utils-security': ['./src/lib/utils/security.js'],
          'utils-performance': ['./src/lib/utils/performance.js'],
          
          // コンポーネント関連
          'components-ui': [
            './src/lib/components/Button.svelte',
            './src/lib/components/Input.svelte',
            './src/lib/components/Select.svelte',
            './src/lib/components/Modal.svelte',
            './src/lib/components/LoadingSpinner.svelte'
          ],
          'components-layout': [
            './src/lib/components/ResponsiveLayout.svelte',
            './src/lib/components/ResponsiveGrid.svelte',
            './src/lib/components/ResponsiveNavigation.svelte'
          ],
          'components-tournament': [
            './src/lib/components/TournamentBracket.svelte',
            './src/lib/components/MatchCard.svelte'
          ],
          'components-admin': [
            './src/lib/components/AdminMatchForm.svelte'
          ],
          'components-animation': [
            './src/lib/components/AnimatedTransition.svelte',
            './src/lib/components/PageTransition.svelte',
            './src/lib/components/StaggeredList.svelte'
          ]
        },
        // ファイル名にハッシュを追加
        chunkFileNames: 'assets/[name]-[hash].js',
        entryFileNames: 'assets/[name]-[hash].js',
        assetFileNames: 'assets/[name]-[hash].[ext]'
      }
    },
    // 分析用の設定
    minify: 'terser',
    terserOptions: {
      compress: {
        drop_console: false, // 分析時はconsole.logを残す
        drop_debugger: true
      }
    },
    // ソースマップを生成（分析用）
    sourcemap: true,
    // 詳細な分析情報を出力
    reportCompressedSize: true,
    chunkSizeWarningLimit: 500
  },
  // 依存関係の最適化
  optimizeDeps: {
    include: ['svelte', '@sveltejs/kit'],
    exclude: []
  }
});