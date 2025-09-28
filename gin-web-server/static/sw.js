// Service Worker 版本号 - 更新此版本号以触发更新
const CACHE_VERSION = 'v1.0.1';
const CACHE_NAME = 'web-tools-cache-' + CACHE_VERSION;
const RUNTIME_CACHE = 'runtime-cache-' + CACHE_VERSION;

// 检查是否为开发环境
const isDevelopment = location.hostname === 'localhost' || location.hostname === '127.0.0.1';

// 需要缓存的资源列表
const CACHE_URLS = [
  '/static/index.html',
  '/static/manifest.json',
  '/static/network_info_detector.html',
  '/static/ip_info_detector.html', 
  '/static/microphone_test.html',
  '/static/camera_test.html',
  '/static/socket_test.html',
  '/static/curl_test.html',
  '/static/port_scan.html',
  '/static/chrome_voice_chat_debug.html',
  // 可以根据需要添加更多静态资源
];

// Service Worker 安装事件
self.addEventListener('install', event => {
  console.log('Service Worker 正在安装...', CACHE_VERSION);
  
  event.waitUntil(
    caches.open(CACHE_NAME)
      .then(cache => {
        console.log('缓存已打开，开始预缓存资源');
        return cache.addAll(CACHE_URLS);
      })
      .then(() => {
        console.log('资源预缓存完成');
        // 强制激活新的Service Worker
        return self.skipWaiting();
      })
      .catch(error => {
        console.error('预缓存失败:', error);
      })
  );
});

// Service Worker 激活事件
self.addEventListener('activate', event => {
  console.log('Service Worker 正在激活...', CACHE_VERSION);
  
  event.waitUntil(
    caches.keys()
      .then(cacheNames => {
        // 删除旧版本的缓存
        return Promise.all(
          cacheNames.map(cacheName => {
            if (cacheName !== CACHE_NAME) {
              console.log('删除旧缓存:', cacheName);
              return caches.delete(cacheName);
            }
          })
        );
      })
      .then(() => {
        console.log('Service Worker 激活完成');
        // 立即控制所有客户端
        return self.clients.claim();
      })
      .then(() => {
        // 通知客户端有新版本可用
        return self.clients.matchAll();
      })
      .then(clients => {
        clients.forEach(client => {
          client.postMessage({
            type: 'NEW_VERSION_AVAILABLE',
            version: CACHE_VERSION
          });
        });
      })
  );
});

// 网络请求拦截
self.addEventListener('fetch', event => {
  // 只处理 GET 请求
  if (event.request.method !== 'GET') {
    return;
  }

  // 开发环境直接使用网络请求，不缓存
  if (isDevelopment) {
    event.respondWith(fetch(event.request));
    return;
  }

  const url = new URL(event.request.url);
  
  // HTML文件使用Network First策略（优先网络）
  if (event.request.destination === 'document' || 
      url.pathname.endsWith('.html') || 
      url.pathname === '/' || 
      url.pathname === '/static/') {
    event.respondWith(networkFirstStrategy(event.request));
  }
  // 静态资源使用Cache First策略
  else {
    event.respondWith(cacheFirstStrategy(event.request));
  }
});

// Network First 策略 - 优先使用网络，失败时使用缓存
async function networkFirstStrategy(request) {
  try {
    // 首先尝试网络请求
    const networkResponse = await fetch(request);
    
    if (networkResponse && networkResponse.status === 200) {
      // 成功获取网络响应，更新缓存
      const cache = await caches.open(RUNTIME_CACHE);
      cache.put(request, networkResponse.clone());
      return networkResponse;
    }
  } catch (error) {
    console.log('网络请求失败，尝试使用缓存:', request.url);
  }
  
  // 网络失败，使用缓存
  const cachedResponse = await caches.match(request);
  if (cachedResponse) {
    return cachedResponse;
  }
  
  // 缓存也没有，返回离线页面
  if (request.destination === 'document') {
    const offlineResponse = await caches.match('/static/index.html');
    return offlineResponse || new Response('离线模式 - 请检查网络连接', {
      status: 503,
      headers: { 'Content-Type': 'text/plain; charset=utf-8' }
    });
  }
  
  return new Response('资源不可用', { status: 404 });
}

// Cache First 策略 - 优先使用缓存
async function cacheFirstStrategy(request) {
  // 先检查缓存
  const cachedResponse = await caches.match(request);
  if (cachedResponse) {
    // 异步更新缓存（可选）
    fetch(request).then(response => {
      if (response && response.status === 200) {
        caches.open(RUNTIME_CACHE).then(cache => {
          cache.put(request, response.clone());
        });
      }
    }).catch(() => {});
    
    return cachedResponse;
  }

  // 缓存中没有，尝试网络请求
  try {
    const networkResponse = await fetch(request);
    
    if (networkResponse && networkResponse.status === 200) {
      // 将响应添加到缓存
      const cache = await caches.open(RUNTIME_CACHE);
      cache.put(request, networkResponse.clone());
      return networkResponse;
    }
    
    return networkResponse;
  } catch (error) {
    return new Response('网络错误', { status: 503 });
  }
}

// 监听消息事件
self.addEventListener('message', event => {
  if (event.data && event.data.type === 'SKIP_WAITING') {
    self.skipWaiting();
  }
  
  if (event.data && event.data.type === 'GET_VERSION') {
    event.ports[0].postMessage({
      type: 'VERSION_INFO',
      version: CACHE_VERSION
    });
  }

  // 清理所有缓存
  if (event.data && event.data.type === 'CLEAR_CACHE') {
    event.waitUntil(clearAllCaches().then(() => {
      event.ports[0].postMessage({
        type: 'CACHE_CLEARED',
        success: true
      });
    }));
  }

  // 强制更新缓存
  if (event.data && event.data.type === 'FORCE_UPDATE') {
    event.waitUntil(forceUpdateCache().then(() => {
      event.ports[0].postMessage({
        type: 'CACHE_UPDATED',
        success: true
      });
    }));
  }
});

// 清理所有缓存
async function clearAllCaches() {
  try {
    const cacheNames = await caches.keys();
    await Promise.all(cacheNames.map(cacheName => caches.delete(cacheName)));
    console.log('所有缓存已清理');
    return true;
  } catch (error) {
    console.error('清理缓存失败:', error);
    return false;
  }
}

// 强制更新缓存
async function forceUpdateCache() {
  try {
    const cache = await caches.open(CACHE_NAME);
    
    // 删除旧缓存并重新获取
    for (const url of CACHE_URLS) {
      await cache.delete(url);
      try {
        const response = await fetch(url + '?t=' + Date.now()); // 添加时间戳防止浏览器缓存
        if (response.ok) {
          await cache.put(url, response);
        }
      } catch (fetchError) {
        console.warn('无法更新缓存:', url, fetchError);
      }
    }
    
    console.log('缓存强制更新完成');
    return true;
  } catch (error) {
    console.error('强制更新缓存失败:', error);
    return false;
  }
}

// 后台同步（如果支持）
if ('sync' in self.registration) {
  self.addEventListener('sync', event => {
    if (event.tag === 'background-sync') {
      event.waitUntil(doBackgroundSync());
    }
  });
}

// 后台同步处理函数
async function doBackgroundSync() {
  try {
    // 这里可以添加后台同步逻辑
    console.log('执行后台同步');
  } catch (error) {
    console.error('后台同步失败:', error);
  }
}

// 推送通知（如果需要）
self.addEventListener('push', event => {
  const options = {
    body: event.data ? event.data.text() : '您有新的消息',
    icon: '/static/icons/icon-192x192.png',
    badge: '/static/icons/icon-72x72.png',
    vibrate: [100, 50, 100],
    data: {
      dateOfArrival: Date.now(),
      primaryKey: 1
    },
    actions: [
      {
        action: 'explore',
        title: '查看详情',
        icon: '/static/icons/checkmark.png'
      },
      {
        action: 'close',
        title: '关闭',
        icon: '/static/icons/xmark.png'
      }
    ]
  };

  event.waitUntil(
    self.registration.showNotification('网页工具通知', options)
  );
});

// 通知点击事件
self.addEventListener('notificationclick', event => {
  event.notification.close();

  if (event.action === 'explore') {
    event.waitUntil(
      clients.openWindow('/static/index.html')
    );
  }
});

console.log('Service Worker 脚本已加载，版本:', CACHE_VERSION);


