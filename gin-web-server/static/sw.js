// Service Worker 版本号 - 更新此版本号以触发更新
const CACHE_VERSION = 'v1.0.0';
const CACHE_NAME = 'web-tools-cache-' + CACHE_VERSION;

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

  event.respondWith(
    caches.match(event.request)
      .then(response => {
        // 如果缓存中有，先返回缓存的版本
        if (response) {
          // 异步更新缓存
          fetch(event.request)
            .then(fetchResponse => {
              if (fetchResponse && fetchResponse.status === 200) {
                const responseClone = fetchResponse.clone();
                caches.open(CACHE_NAME)
                  .then(cache => {
                    cache.put(event.request, responseClone);
                  });
              }
            })
            .catch(() => {
              // 网络错误时忽略
            });
          
          return response;
        }

        // 缓存中没有，尝试网络请求
        return fetch(event.request)
          .then(fetchResponse => {
            // 检查是否是有效的响应
            if (!fetchResponse || fetchResponse.status !== 200 || fetchResponse.type !== 'basic') {
              return fetchResponse;
            }

            // 克隆响应，因为响应流只能使用一次
            const responseToCache = fetchResponse.clone();

            // 将响应添加到缓存
            caches.open(CACHE_NAME)
              .then(cache => {
                cache.put(event.request, responseToCache);
              });

            return fetchResponse;
          })
          .catch(() => {
            // 网络失败时，如果是HTML页面请求，返回离线页面
            if (event.request.destination === 'document') {
              return caches.match('/static/index.html');
            }
          });
      })
  );
});

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
});

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
