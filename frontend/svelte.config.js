import adapter from '@sveltejs/adapter-static';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
  // Consult https://kit.svelte.dev/docs/integrations#preprocessors
  // for more information about preprocessors
  preprocess: vitePreprocess(),

  kit: {
    // 静的サイト生成用のアダプター（Docker/Nginx用）
    adapter: adapter({
      // 出力ディレクトリ
      pages: 'build',
      assets: 'build',
      fallback: 'index.html', // SPA用のフォールバック
      precompress: false, // Nginxで圧縮するため無効化
      strict: true
    }),
    
    alias: {
      $lib: 'src/lib'
    },
    
    // プリレンダリング設定
    prerender: {
      handleHttpError: 'warn',
      handleMissingId: 'warn',
      entries: [
        '/',
        '/login'
        // 管理者ページは認証が必要なためプリレンダリングしない
      ]
    },
    
    // CSP設定（セキュリティ強化）
    csp: {
      mode: 'auto',
      directives: {
        'default-src': ['self'],
        'script-src': ['self', 'unsafe-inline', 'unsafe-eval'],
        'style-src': ['self', 'unsafe-inline'],
        'img-src': ['self', 'data:', 'https:'],
        'font-src': ['self', 'data:'],
        'connect-src': ['self', 'https:'],
        'frame-ancestors': ['self']
      }
    }
  }
};

export default config;
