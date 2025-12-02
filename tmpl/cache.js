const CACHE_NAME = "mujamalat--cache-v1";

function shouldCache(request) {
  try {
    const u = new URL(request.url);
    return u.pathname.startsWith('/content') || u.pathname.startsWith('/pub/');
  } catch {
    return false;
  }
}

self.addEventListener('install', () => self.skipWaiting());
self.addEventListener('activate',
  event => event.waitUntil(self.clients.claim()));

self.addEventListener('fetch', event => {
  const request = event.request;
  if (!shouldCache(request)) return;

  event.respondWith((async () => {
    const cache = await caches.open(CACHE_NAME);

    // Serve from cache if available
    const cached = await cache.match(request);
    if (cached) {
      return cloneWithHeader(cached, 'true');
    }

    // Otherwise fetch
    try {
      const fresh = await fetch(request);

      // Only store valid responses
      if (fresh.ok) {
        cache.put(request, fresh.clone());
      }

      return cloneWithHeader(fresh, 'false');

    } catch (err) {
      // Offline + not cached => fail cleanly
      return new Response('Offline & Not Cached', { status: 503 });
    }
  })());
});

function cloneWithHeader(response, wasCached) {
  return new Response(response.body, {
    status: response.status,
    statusText: response.statusText,
    headers: {
      ...Object.fromEntries(response.headers.entries()),
      'X-From-Cache': wasCached
    }
  });
}

