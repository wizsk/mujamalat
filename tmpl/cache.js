const CACHE_NAME = "mujamalat--cache-v1";

function shouldCache(request) {
  try {
    const u = new URL(request.url);
    if (u.pathname.startsWith('/pub/static')) {
      return false;
    }

    return u.pathname.startsWith('/content') ||
      (u.pathname.startsWith('/pub/') && !u.pathname.endsWith('/'));
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
      return await cloneWithHeader(cached, 'true');
    }

    // Otherwise fetch
    try {
      const fresh = await fetch(request);

      // Only store valid responses
      if (fresh.ok) {
        cache.put(request, fresh.clone());
      }

      return await cloneWithHeader(fresh, 'false');

    } catch (err) {
      // Offline + not cached => fail cleanly
      return new Response('Offline & Not Cached', { status: 503 });
    }
  })());
});

// Listen for messages to clear cache
self.addEventListener('message', event => {
  if (event.data && event.data.action === 'clearCache') {
    event.waitUntil(
      clearCache().then(() => {
        // Notify the requesting client
        if (event.source) {
          event.source.postMessage({ action: 'cacheCleared', success: true });
        }
      }).catch(err => {
        // Notify about error
        if (event.source) {
          event.source.postMessage({ action: 'cacheCleared', success: false, error: err.message });
        }
      })
    );
  }
});

async function clearCache() {
  try {
    const deleted = await caches.delete(CACHE_NAME);
    console.log(`Cache ${CACHE_NAME} deleted:`, deleted);
    return deleted;
  } catch (err) {
    console.error('Error clearing cache:', err);
    throw err;
  }
}

async function cloneWithHeader(response, wasCached) {
  const headers = new Headers(response.headers);
  headers.set('X-From-Cache', wasCached);

  const buf = await response.clone().arrayBuffer();

  return new Response(buf, {
    status: response.status,
    statusText: response.statusText,
    headers
  });
}

