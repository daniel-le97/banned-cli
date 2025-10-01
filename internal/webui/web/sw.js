// Banned CLI PWA Service Worker
const CACHE_NAME = 'banned-cli-v1';
const STATIC_CACHE = 'banned-cli-static-v1';
const DYNAMIC_CACHE = 'banned-cli-dynamic-v1';
const ESM_CACHE = 'banned-cli-esm-v1';

// Files to cache for offline use
const STATIC_FILES = [
  '/',
  '/index.html',
  '/style.css',
  '/app.js',
  '/manifest.json',
  '/icon-192.png',
  '/icon-512.png'
];

// ESM imports to cache for offline use
const ESM_IMPORTS = [
  'https://esm.sh/preact@10.23.1',
  'https://esm.sh/preact@10.23.1/hooks',
  'https://esm.sh/htm@3.1.1/preact?external=preact'
];

// API endpoints that can work offline (cached)
const CACHEABLE_APIS = [
  '/api/channels',
  '/api/videos',
  '/api/stats'
];

// ESM.sh domain patterns for caching
const ESM_PATTERNS = [
  /^https:\/\/esm\.sh\//,
  /^https:\/\/cdn\.skypack\.dev\//,
  /^https:\/\/unpkg\.com\//,
  /^https:\/\/cdn\.jsdelivr\.net\//
];

// Install event - cache static files and ESM imports
self.addEventListener('install', event => {
  console.log('[SW] Installing service worker...');
  
  event.waitUntil(
    Promise.all([
      // Cache static files
      caches.open(STATIC_CACHE).then(cache => {
        console.log('[SW] Caching static files...');
        return cache.addAll(STATIC_FILES);
      }),
      // Cache ESM imports
      caches.open(ESM_CACHE).then(cache => {
        console.log('[SW] Caching ESM imports...');
        return cache.addAll(ESM_IMPORTS);
      })
    ]).then(() => {
      console.log('[SW] Service worker installed successfully');
      return self.skipWaiting();
    })
  );
});

// Activate event - cleanup old caches
self.addEventListener('activate', event => {
  console.log('[SW] Activating service worker...');
  
  event.waitUntil(
    caches.keys()
      .then(cacheNames => {
        return Promise.all(
          cacheNames.map(cacheName => {
            if (cacheName !== STATIC_CACHE && cacheName !== DYNAMIC_CACHE && cacheName !== ESM_CACHE) {
              console.log('[SW] Deleting old cache:', cacheName);
              return caches.delete(cacheName);
            }
          })
        );
      })
      .then(() => {
        console.log('[SW] Service worker activated');
        return self.clients.claim();
      })
  );
});

// Fetch event - handle requests with cache-first strategy
self.addEventListener('fetch', event => {
  const requestUrl = new URL(event.request.url);
  
  // Handle ESM imports (cache-first with network fallback)
  if (ESM_PATTERNS.some(pattern => pattern.test(event.request.url))) {
    event.respondWith(
      caches.match(event.request)
        .then(response => {
          if (response) {
            console.log('[SW] ESM cache hit:', event.request.url);
            return response;
          }
          
          // Fetch from network and cache
          return fetch(event.request)
            .then(fetchResponse => {
              if (fetchResponse.ok) {
                const responseClone = fetchResponse.clone();
                caches.open(ESM_CACHE)
                  .then(cache => cache.put(event.request, responseClone));
                console.log('[SW] ESM cached:', event.request.url);
              }
              return fetchResponse;
            })
            .catch(error => {
              console.error('[SW] ESM fetch failed:', event.request.url, error);
              // Return cached version or create error response
              return caches.match(event.request).then(cachedResponse => {
                if (cachedResponse) {
                  return cachedResponse;
                }
                return new Response('// ESM module unavailable offline', {
                  status: 503,
                  statusText: 'Service Unavailable',
                  headers: { 'Content-Type': 'application/javascript' }
                });
              });
            });
        })
    );
    return;
  }
  
  // Handle static files (cache-first)
  if (STATIC_FILES.some(file => requestUrl.pathname === file || requestUrl.pathname.endsWith(file))) {
    event.respondWith(
      caches.match(event.request)
        .then(response => {
          return response || fetch(event.request)
            .then(fetchResponse => {
              return caches.open(STATIC_CACHE)
                .then(cache => {
                  cache.put(event.request, fetchResponse.clone());
                  return fetchResponse;
                });
            });
        })
        .catch(() => {
          // Return offline fallback if available
          if (requestUrl.pathname === '/' || requestUrl.pathname === '/index.html') {
            return caches.match('/offline.html');
          }
        })
    );
    return;
  }
  
  // Handle API requests (network-first with cache fallback)
  if (requestUrl.pathname.startsWith('/api/')) {
    event.respondWith(
      fetch(event.request)
        .then(response => {
          // Cache successful GET requests
          if (response.ok && event.request.method === 'GET' && 
              CACHEABLE_APIS.some(api => requestUrl.pathname.startsWith(api))) {
            const responseClone = response.clone();
            caches.open(DYNAMIC_CACHE)
              .then(cache => cache.put(event.request, responseClone));
          }
          return response;
        })
        .catch(() => {
          // Return cached data if network fails
          if (event.request.method === 'GET') {
            return caches.match(event.request)
              .then(cachedResponse => {
                if (cachedResponse) {
                  // Add offline indicator header
                  const headers = new Headers(cachedResponse.headers);
                  headers.set('X-Served-From', 'cache');
                  return new Response(cachedResponse.body, {
                    status: cachedResponse.status,
                    statusText: cachedResponse.statusText,
                    headers: headers
                  });
                }
                // Return offline message for API requests
                return new Response(
                  JSON.stringify({ 
                    error: 'Offline', 
                    message: 'This feature requires internet connection' 
                  }), 
                  { 
                    status: 503, 
                    headers: { 'Content-Type': 'application/json' } 
                  }
                );
              });
          }
        })
    );
    return;
  }
  
  // Default: try network first, fall back to cache
  event.respondWith(
    fetch(event.request)
      .then(response => {
        // Cache successful responses
        if (response.ok) {
          const responseClone = response.clone();
          caches.open(DYNAMIC_CACHE)
            .then(cache => cache.put(event.request, responseClone));
        }
        return response;
      })
      .catch(() => {
        return caches.match(event.request);
      })
  );
});

// Background sync for data updates
self.addEventListener('sync', event => {
  console.log('[SW] Background sync triggered:', event.tag);
  
  if (event.tag === 'background-sync-data') {
    event.waitUntil(syncData());
  }
});

// Push notifications for daemon status updates
self.addEventListener('push', event => {
  console.log('[SW] Push notification received');
  
  if (event.data) {
    const data = event.data.json();
    const options = {
      body: data.body || 'Banned CLI daemon update',
      icon: '/icon-192.png',
      badge: '/icon-192.png',
      tag: data.tag || 'daemon-update',
      requireInteraction: data.requireInteraction || false,
      actions: data.actions || []
    };
    
    event.waitUntil(
      self.registration.showNotification(data.title || 'Banned CLI', options)
    );
  }
});

// Handle notification clicks
self.addEventListener('notificationclick', event => {
  console.log('[SW] Notification clicked:', event.notification.tag);
  
  event.notification.close();
  
  // Open the app or focus existing window
  event.waitUntil(
    clients.matchAll({ type: 'window' })
      .then(clientList => {
        for (const client of clientList) {
          if (client.url.includes(self.location.origin) && 'focus' in client) {
            return client.focus();
          }
        }
        // Open new window if none exists
        if (clients.openWindow) {
          return clients.openWindow('/');
        }
      })
  );
});

// Background data sync function
async function syncData() {
  try {
    console.log('[SW] Starting background data sync...');
    
    // Sync channel data
    const channelsResponse = await fetch('/api/sync/channels');
    if (channelsResponse.ok) {
      console.log('[SW] Channels synced successfully');
    }
    
    // Sync video data
    const videosResponse = await fetch('/api/sync/videos');
    if (videosResponse.ok) {
      console.log('[SW] Videos synced successfully');
    }
    
    // Notify user of successful sync
    await self.registration.showNotification('Sync Complete', {
      body: 'Banned.video data has been updated',
      icon: '/icon-192.png',
      tag: 'sync-complete'
    });
    
  } catch (error) {
    console.error('[SW] Background sync failed:', error);
  }
}

// Handle messages from main thread
self.addEventListener('message', event => {
  console.log('[SW] Message received:', event.data);
  
  if (event.data && event.data.type === 'SKIP_WAITING') {
    self.skipWaiting();
  }
  
  if (event.data && event.data.type === 'GET_VERSION') {
    event.ports[0].postMessage({ version: CACHE_NAME });
  }
});