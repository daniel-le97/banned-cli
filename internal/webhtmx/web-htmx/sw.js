/**
 * Service Worker for Database Editor HTMX PWA
 * Handles caching, offline functionality, and background sync
 */

const CACHE_NAME = 'db-editor-htmx-v2';
const STATIC_CACHE = 'static-v2';
const API_CACHE = 'api-v2';

// Files to cache for offline use
const STATIC_FILES = [
    '/htmx/',
    '/web-htmx/style.css',
    '/web-htmx/app.js',
    '/web-htmx/htmx.js',
    '/web-htmx/hyperscript.js',
    '/web-htmx/manifest.json',
    '/web-htmx/icons/icon-192x192.png',
    '/web-htmx/icons/icon-512x512.png'
];

// API endpoints to cache (with shorter TTL)
const API_ENDPOINTS = [
    '/htmx/stats',
    '/htmx/tables',
    '/api/health'
];

/**
 * Service Worker Installation
 * Pre-cache essential files
 */
self.addEventListener( 'install', event =>
{
    console.log( 'Service Worker: Installing...' );

    event.waitUntil(
        Promise.all( [
            // Cache static files
            caches.open( STATIC_CACHE ).then( cache =>
            {
                console.log( 'Service Worker: Caching static files' );
                return cache.addAll( STATIC_FILES );
            } ),

            // Cache API endpoints
            caches.open( API_CACHE ).then( cache =>
            {
                console.log( 'Service Worker: Pre-caching API endpoints' );
                return Promise.all(
                    API_ENDPOINTS.map( url =>
                        cache.add( url ).catch( err =>
                        {
                            console.warn( 'Failed to cache:', url, err );
                            return null;
                        } )
                    )
                );
            } )
        ] ).then( () =>
        {
            console.log( 'Service Worker: Installation complete' );
            // Force activation
            return self.skipWaiting();
        } )
    );
} );

/**
 * Service Worker Activation
 * Clean up old caches
 */
self.addEventListener( 'activate', event =>
{
    console.log( 'Service Worker: Activating...' );

    event.waitUntil(
        caches.keys().then( cacheNames =>
        {
            return Promise.all(
                cacheNames.map( cacheName =>
                {
                    // Delete old caches
                    if ( cacheName !== STATIC_CACHE &&
                        cacheName !== API_CACHE &&
                        cacheName !== CACHE_NAME )
                    {
                        console.log( 'Service Worker: Deleting old cache:', cacheName );
                        return caches.delete( cacheName );
                    }
                } )
            );
        } ).then( () =>
        {
            console.log( 'Service Worker: Activation complete' );
            // Take control of all clients
            return self.clients.claim();
        } )
    );
} );

/**
 * Fetch Event Handler
 * Implement caching strategies
 */
self.addEventListener( 'fetch', event =>
{
    const { request } = event;
    const url = new URL( request.url );

    // Skip non-GET requests
    if ( request.method !== 'GET' )
    {
        return;
    }

    // Handle different types of requests
    if ( isStaticFile( url ) )
    {
        event.respondWith( handleStaticFile( request ) );
    } else if ( isAPIRequest( url ) )
    {
        event.respondWith( handleAPIRequest( request ) );
    } else if ( isHTMXPage( url ) )
    {
        event.respondWith( handleHTMXPage( request ) );
    }
} );

/**
 * Check if request is for a static file
 * @param {URL} url - Request URL
 * @returns {boolean}
 */
function isStaticFile ( url )
{
    return url.pathname.includes( '/web-htmx/' ) ||
        url.pathname.endsWith( '.css' ) ||
        url.pathname.endsWith( '.js' ) ||
        url.pathname.endsWith( '.png' );
}

/**
 * Check if request is for an API endpoint
 * @param {URL} url - Request URL
 * @returns {boolean}
 */
function isAPIRequest ( url )
{
    return url.pathname.startsWith( '/api/' ) ||
        url.pathname.startsWith( '/htmx/' );
}

/**
 * Check if request is for an HTMX page
 * @param {URL} url - Request URL
 * @returns {boolean}
 */
function isHTMXPage ( url )
{
    return url.pathname === '/htmx/' ||
        url.pathname === '/htmx';
}

/**
 * Handle static file requests
 * Cache-first strategy with network fallback
 * @param {Request} request
 * @returns {Promise<Response>}
 */
async function handleStaticFile ( request )
{
    try
    {
        // Try cache first
        const cachedResponse = await caches.match( request );
        if ( cachedResponse )
        {
            return cachedResponse;
        }

        // Fallback to network
        const networkResponse = await fetch( request );
        if ( networkResponse.ok )
        {
            // Cache successful responses
            const cache = await caches.open( STATIC_CACHE );
            cache.put( request, networkResponse.clone() );
        }

        return networkResponse;
    } catch ( error )
    {
        console.error( 'Static file fetch failed:', error );

        // Return offline fallback if available
        if ( request.destination === 'document' )
        {
            return caches.match( '/htmx/' );
        }

        // Return a basic error response
        return new Response( 'Offline', {
            status: 503,
            statusText: 'Service Unavailable'
        } );
    }
}

/**
 * Handle API requests
 * Network-first strategy with cache fallback
 * @param {Request} request
 * @returns {Promise<Response>}
 */
async function handleAPIRequest ( request )
{
    try
    {
        // Try network first for fresh data
        const networkResponse = await fetch( request );

        if ( networkResponse.ok )
        {
            // Cache successful API responses
            const cache = await caches.open( API_CACHE );
            cache.put( request, networkResponse.clone() );
            return networkResponse;
        }

        // If network fails, try cache
        const cachedResponse = await caches.match( request );
        if ( cachedResponse )
        {
            // Add offline indicator header
            const response = cachedResponse.clone();
            response.headers.append( 'X-Served-By', 'ServiceWorker-Cache' );
            return response;
        }

        return networkResponse;
    } catch ( error )
    {
        console.error( 'API fetch failed:', error );

        // Try cached version
        const cachedResponse = await caches.match( request );
        if ( cachedResponse )
        {
            return cachedResponse;
        }

        // Return offline message for specific endpoints
        if ( request.url.includes( '/htmx/stats' ) )
        {
            return new Response( JSON.stringify( {
                offline: true,
                message: 'Server offline - showing cached data'
            } ), {
                headers: { 'Content-Type': 'application/json' },
                status: 503
            } );
        }

        return new Response( 'Offline', {
            status: 503,
            statusText: 'Service Unavailable'
        } );
    }
}

/**
 * Handle HTMX page requests
 * Cache-first with network update
 * @param {Request} request
 * @returns {Promise<Response>}
 */
async function handleHTMXPage ( request )
{
    try
    {
        // Try cache first for instant loading
        const cachedResponse = await caches.match( request );

        // Always try to update from network in background
        const networkPromise = fetch( request ).then( response =>
        {
            if ( response.ok )
            {
                const cache = caches.open( STATIC_CACHE );
                cache.then( c => c.put( request, response.clone() ) );
            }
            return response;
        } );

        // Return cached version immediately if available
        if ( cachedResponse )
        {
            return cachedResponse;
        }

        // Otherwise wait for network
        return await networkPromise;
    } catch ( error )
    {
        console.error( 'HTMX page fetch failed:', error );

        // Try to return cached version
        const cachedResponse = await caches.match( request );
        if ( cachedResponse )
        {
            return cachedResponse;
        }

        // Return basic offline page
        return new Response( `
      <!DOCTYPE html>
      <html>
        <head>
          <title>Database Editor - Offline</title>
          <meta name="viewport" content="width=device-width, initial-scale=1.0">
          <style>
            body { 
              font-family: Arial, sans-serif; 
              text-align: center; 
              padding: 50px;
              background: #1a1a1a;
              color: #e0e0e0;
            }
            .offline-message {
              background: #2d2d2d;
              padding: 2rem;
              border-radius: 8px;
              max-width: 400px;
              margin: 0 auto;
            }
          </style>
        </head>
        <body>
          <div class="offline-message">
            <h1>Database Editor</h1>
            <h2>You're Offline</h2>
            <p>The database editor is currently unavailable.</p>
            <p>Please check your connection and try again.</p>
            <button onclick="window.location.reload()">Retry</button>
          </div>
        </body>
      </html>
    `, {
            headers: { 'Content-Type': 'text/html' },
            status: 503
        } );
    }
}

/**
 * Handle background sync events
 */
self.addEventListener( 'sync', event =>
{
    console.log( 'Service Worker: Background sync triggered' );

    if ( event.tag === 'background-sync' )
    {
        event.waitUntil(
            // Refresh cache when connection is restored
            refreshCache()
        );
    }
} );

/**
 * Refresh cached data
 */
async function refreshCache ()
{
    try
    {
        const cache = await caches.open( API_CACHE );

        // Update API endpoints
        await Promise.all(
            API_ENDPOINTS.map( async url =>
            {
                try
                {
                    const response = await fetch( url );
                    if ( response.ok )
                    {
                        await cache.put( url, response );
                    }
                } catch ( error )
                {
                    console.warn( 'Failed to refresh cache for:', url, error );
                }
            } )
        );

        console.log( 'Service Worker: Cache refreshed' );
    } catch ( error )
    {
        console.error( 'Service Worker: Cache refresh failed:', error );
    }
}