/**
 * Service Worker
 * オフライン対応とアセットキャッシュ戦略を提供
 */

const CACHE_NAME = 'tournament-app-v1';
const STATIC_CACHE_NAME = 'tournament-static-v1';
const DYNAMIC_CACHE_NAME = 'tournament-dynamic-v1';
const API_CACHE_NAME = 'tournament-api-v1';

// キャッシュするファイルのリスト
const STATIC_FILES = [
  '/',
  '/offline',
  '/manifest.json',
  // 重要なアセットをここに追加
];

// キャッシュ戦略の設定
const CACHE_STRATEGIES = {
  // 静的アセット: Cache First
  static: {
    pattern: /\.(js|css|woff|woff2|ttf|eot|ico|png|jpg|jpeg|gif|svg|webp)$/,
    strategy: 'cacheFirst',
    cacheName: STATIC_CACHE_NAME,
    maxAge: 7 * 24 * 60 * 60 * 1000, // 7日
    maxEntries: 100
  },
  
  // APIリクエスト: Network First
  api: {
    pattern: /^https?:\/\/.*\/api\/.*/,
    strategy: 'networkFirst',
    cacheName: API_CACHE_NAME,
    maxAge: 5 * 60 * 1000, // 5分
    maxEntries: 50
  },
  
  // ページ: Stale While Revalidate
  pages: {
    pattern: /^https?:\/\/.*\/(admin|login)?$/,
    strategy: 'staleWhileRevalidate',
    cacheName: DYNAMIC_CACHE_NAME,
    maxAge: 24 * 60 * 60 * 1000, // 1日
    maxEntries: 20
  }
};

/**
 * Service Worker インストール時の処理
 */
self.addEventListener('install', (event) => {
  console.log('[SW] Installing Service Worker');
  
  event.waitUntil(
    caches.open(STATIC_CACHE_NAME)
      .then((cache) => {
        console.log('[SW] Caching static files');
        return cache.addAll(STATIC_FILES);
      })
      .then(() => {
        console.log('[SW] Static files cached successfully');
        return self.skipWaiting();
      })
      .catch((error) => {
        console.error('[SW] Failed to cache static files:', error);
      })
  );
});

/**
 * Service Worker アクティベート時の処理
 */
self.addEventListener('activate', (event) => {
  console.log('[SW] Activating Service Worker');
  
  event.waitUntil(
    caches.keys()
      .then((cacheNames) => {
        return Promise.all(
          cacheNames.map((cacheName) => {
            // 古いキャッシュを削除
            if (cacheName !== STATIC_CACHE_NAME && 
                cacheName !== DYNAMIC_CACHE_NAME && 
                cacheName !== API_CACHE_NAME) {
              console.log('[SW] Deleting old cache:', cacheName);
              return caches.delete(cacheName);
            }
          })
        );
      })
      .then(() => {
        console.log('[SW] Service Worker activated');
        return self.clients.claim();
      })
  );
});

/**
 * フェッチイベントの処理
 */
self.addEventListener('fetch', (event) => {
  const { request } = event;
  const url = new URL(request.url);
  
  // 外部リソースは処理しない
  if (url.origin !== location.origin && !url.pathname.startsWith('/api')) {
    return;
  }
  
  // 適切なキャッシュ戦略を選択
  const strategy = getStrategy(request.url);
  
  if (strategy) {
    event.respondWith(handleRequest(request, strategy));
  }
});

/**
 * リクエストURLに基づいてキャッシュ戦略を取得
 */
function getStrategy(url) {
  for (const [name, config] of Object.entries(CACHE_STRATEGIES)) {
    if (config.pattern.test(url)) {
      return config;
    }
  }
  return null;
}

/**
 * キャッシュ戦略に基づいてリクエストを処理
 */
async function handleRequest(request, strategy) {
  switch (strategy.strategy) {
    case 'cacheFirst':
      return cacheFirstStrategy(request, strategy);
    
    case 'networkFirst':
      return networkFirstStrategy(request, strategy);
    
    case 'staleWhileRevalidate':
      return staleWhileRevalidateStrategy(request, strategy);
    
    default:
      return fetch(request);
  }
}

/**
 * Cache First 戦略
 */
async function cacheFirstStrategy(request, strategy) {
  try {
    const cache = await caches.open(strategy.cacheName);
    const cachedResponse = await cache.match(request);
    
    if (cachedResponse) {
      // キャッシュの有効期限をチェック
      const cacheDate = new Date(cachedResponse.headers.get('sw-cache-date') || 0);
      const isExpired = Date.now() - cacheDate.getTime() > strategy.maxAge;
      
      if (!isExpired) {
        return cachedResponse;
      }
    }
    
    // ネットワークから取得
    const networkResponse = await fetch(request);
    
    if (networkResponse.ok) {
      const responseToCache = networkResponse.clone();
      
      // キャッシュ日時を追加
      const headers = new Headers(responseToCache.headers);
      headers.set('sw-cache-date', new Date().toISOString());
      
      const modifiedResponse = new Response(responseToCache.body, {
        status: responseToCache.status,
        statusText: responseToCache.statusText,
        headers: headers
      });
      
      await cache.put(request, modifiedResponse);
      await cleanupCache(cache, strategy.maxEntries);
    }
    
    return networkResponse;
    
  } catch (error) {
    console.error('[SW] Cache First strategy failed:', error);
    
    // フォールバック: キャッシュから取得（期限切れでも）
    const cache = await caches.open(strategy.cacheName);
    const cachedResponse = await cache.match(request);
    
    if (cachedResponse) {
      return cachedResponse;
    }
    
    // オフラインページを返す
    if (request.destination === 'document') {
      return caches.match('/offline');
    }
    
    throw error;
  }
}

/**
 * Network First 戦略
 */
async function networkFirstStrategy(request, strategy) {
  try {
    const networkResponse = await fetch(request);
    
    if (networkResponse.ok) {
      const cache = await caches.open(strategy.cacheName);
      const responseToCache = networkResponse.clone();
      
      // キャッシュ日時を追加
      const headers = new Headers(responseToCache.headers);
      headers.set('sw-cache-date', new Date().toISOString());
      
      const modifiedResponse = new Response(responseToCache.body, {
        status: responseToCache.status,
        statusText: responseToCache.statusText,
        headers: headers
      });
      
      await cache.put(request, modifiedResponse);
      await cleanupCache(cache, strategy.maxEntries);
    }
    
    return networkResponse;
    
  } catch (error) {
    console.error('[SW] Network First strategy failed:', error);
    
    // フォールバック: キャッシュから取得
    const cache = await caches.open(strategy.cacheName);
    const cachedResponse = await cache.match(request);
    
    if (cachedResponse) {
      return cachedResponse;
    }
    
    throw error;
  }
}

/**
 * Stale While Revalidate 戦略
 */
async function staleWhileRevalidateStrategy(request, strategy) {
  const cache = await caches.open(strategy.cacheName);
  const cachedResponse = await cache.match(request);
  
  // バックグラウンドでネットワークから更新
  const networkResponsePromise = fetch(request)
    .then(async (networkResponse) => {
      if (networkResponse.ok) {
        const responseToCache = networkResponse.clone();
        
        // キャッシュ日時を追加
        const headers = new Headers(responseToCache.headers);
        headers.set('sw-cache-date', new Date().toISOString());
        
        const modifiedResponse = new Response(responseToCache.body, {
          status: responseToCache.status,
          statusText: responseToCache.statusText,
          headers: headers
        });
        
        await cache.put(request, modifiedResponse);
        await cleanupCache(cache, strategy.maxEntries);
      }
      return networkResponse;
    })
    .catch((error) => {
      console.error('[SW] Background update failed:', error);
    });
  
  // キャッシュがあれば即座に返す
  if (cachedResponse) {
    return cachedResponse;
  }
  
  // キャッシュがない場合はネットワークを待つ
  return networkResponsePromise;
}

/**
 * キャッシュのクリーンアップ
 */
async function cleanupCache(cache, maxEntries) {
  const keys = await cache.keys();
  
  if (keys.length > maxEntries) {
    const keysToDelete = keys.slice(0, keys.length - maxEntries);
    await Promise.all(keysToDelete.map(key => cache.delete(key)));
  }
}

/**
 * メッセージイベントの処理
 */
self.addEventListener('message', (event) => {
  const { type, payload } = event.data;
  
  switch (type) {
    case 'SKIP_WAITING':
      self.skipWaiting();
      break;
    
    case 'GET_CACHE_INFO':
      getCacheInfo().then(info => {
        event.ports[0].postMessage({ type: 'CACHE_INFO', payload: info });
      });
      break;
    
    case 'CLEAR_CACHE':
      clearAllCaches().then(() => {
        event.ports[0].postMessage({ type: 'CACHE_CLEARED' });
      });
      break;
    
    default:
      console.warn('[SW] Unknown message type:', type);
  }
});

/**
 * キャッシュ情報を取得
 */
async function getCacheInfo() {
  const cacheNames = await caches.keys();
  const info = {};
  
  for (const cacheName of cacheNames) {
    const cache = await caches.open(cacheName);
    const keys = await cache.keys();
    info[cacheName] = keys.length;
  }
  
  return info;
}

/**
 * 全キャッシュをクリア
 */
async function clearAllCaches() {
  const cacheNames = await caches.keys();
  await Promise.all(cacheNames.map(cacheName => caches.delete(cacheName)));
}

/**
 * 同期イベントの処理（バックグラウンド同期）
 */
self.addEventListener('sync', (event) => {
  if (event.tag === 'background-sync') {
    event.waitUntil(doBackgroundSync());
  }
});

/**
 * バックグラウンド同期処理
 */
async function doBackgroundSync() {
  try {
    // 必要に応じてバックグラウンド同期処理を実装
    console.log('[SW] Background sync completed');
  } catch (error) {
    console.error('[SW] Background sync failed:', error);
  }
}

/**
 * プッシュイベントの処理（将来の拡張用）
 */
self.addEventListener('push', (event) => {
  if (event.data) {
    const data = event.data.json();
    
    event.waitUntil(
      self.registration.showNotification(data.title, {
        body: data.body,
        icon: data.icon || '/icon-192x192.png',
        badge: '/badge-72x72.png',
        tag: data.tag || 'default',
        data: data.data
      })
    );
  }
});

/**
 * 通知クリックイベントの処理
 */
self.addEventListener('notificationclick', (event) => {
  event.notification.close();
  
  event.waitUntil(
    clients.openWindow(event.notification.data?.url || '/')
  );
});