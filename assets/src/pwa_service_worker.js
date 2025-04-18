const CACHE_NAME = 'vgn-cache-v1';
const urlsToCache = [
  '/images/*',
  '/runtime.js',
  '/vendors.js',
  '/admin.js',
  '/admin.css',
];

self.addEventListener('install', event => {
  event.waitUntil(
    caches.open(CACHE_NAME).then(cache => cache.addAll(urlsToCache))
  );
});

self.addEventListener('fetch', event => {
  event.respondWith(
    fetch(event.request).catch(() =>
      caches.match(event.request).then(response => response || caches.match('/offline.html'))
    )
  );
});
