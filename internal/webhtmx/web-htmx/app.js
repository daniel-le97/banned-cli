/**
 * @fileoverview Database Editor HTMX JavaScript
 * Handles clipboard functionality, HTMX configuration, and user interactions
 * for the SQLite database web editor.
 * 
 * @version 1.0.0
 * @author Database Editor Team
 */

/**
 * @typedef {Object} HTMXEvent
 * @property {Object} detail - Event detail object
 * @property {Object} detail.headers - HTTP headers object
 * @property {string} detail.xhr - XMLHttpRequest object
 */

/**
 * @typedef {Object} HTMXErrorEvent
 * @property {Object} detail - Error event detail object
 * @property {number} detail.xhr.status - HTTP status code
 * @property {string} detail.xhr.statusText - HTTP status text
 */

/**
 * Global variable to store the currently selected table name
 * @type {string|null}
 * @global
 */
window.selectedTable = null;

/**
 * Connection state management
 * @type {Object}
 */
window.connectionState = {
    isOnline: navigator.onLine,
    lastSuccessfulRequest: Date.now(),
    retryCount: 0,
    maxRetries: 1, // Only try once before showing offline
    retryInterval: 3000, // 3 seconds
    healthCheckInterval: 15000, // 15 seconds
    isHealthChecking: false,
    offlineMode: false
};

/**
 * Main application initialization
 * Sets up event listeners and HTMX configuration when DOM is ready
 */
document.addEventListener( 'DOMContentLoaded', function ()
{
    console.log( 'Database Editor HTMX loaded' );

    setupHTMXConfiguration();
    setupClipboardFunctionality();
    registerServiceWorker();
} );

/**
 * Configure HTMX global settings and error handling
 * Adds custom headers and error event listeners
 */
function setupHTMXConfiguration ()
{
    // Global HTMX configuration - add HX-Request header to all requests
    document.body.addEventListener( 'htmx:configRequest',
        /** @param {HTMXEvent} evt - HTMX configuration event */
        function ( evt )
        {
            evt.detail.headers[ 'HX-Request' ] = 'true';
            // Add timestamp for request tracking
            evt.detail.headers[ 'X-Request-Time' ] = Date.now().toString();
        }
    );

    // Handle successful HTMX requests
    document.body.addEventListener( 'htmx:afterRequest',
        /** @param {HTMXEvent} evt - HTMX after request event */
        function ( evt )
        {
            if ( evt.detail.xhr.status >= 200 && evt.detail.xhr.status < 300 )
            {
                // Successful request - update connection state
                handleSuccessfulConnection();
            }
        }
    );

    // Handle HTMX errors globally
    document.body.addEventListener( 'htmx:responseError',
        /** @param {HTMXErrorEvent} evt - HTMX error event */
        function ( evt )
        {
            console.error( 'HTMX Error:', evt.detail );
            handleConnectionError( evt.detail.xhr );
        }
    );

    // Handle network errors (when server is completely unreachable)
    document.body.addEventListener( 'htmx:sendError',
        /** @param {HTMXErrorEvent} evt - HTMX send error event */
        function ( evt )
        {
            console.error( 'HTMX Network Error:', evt.detail );
            handleNetworkError();
        }
    );

    // Handle timeout errors
    document.body.addEventListener( 'htmx:timeout',
        /** @param {HTMXErrorEvent} evt - HTMX timeout event */
        function ( evt )
        {
            console.error( 'HTMX Timeout:', evt.detail );
            handleTimeoutError();
        }
    );
}

/**
 * Set up click-to-copy functionality for table cells
 * Attaches event listener to document for event delegation
 */
function setupClipboardFunctionality ()
{
    // Clipboard functionality for table cells using event delegation
    document.addEventListener( 'click',
        /** @param {MouseEvent} event - Click event */
        function ( event )
        {
            const target = /** @type {HTMLElement} */ ( event.target );

            if ( target && target.classList.contains( 'copyable-cell' ) )
            {
                handleCellCopy( target );
            }
        }
    );
}

/**
 * Handle copying cell content to clipboard
 * @param {HTMLElement} cell - The table cell element to copy from
 */
function handleCellCopy ( cell )
{
    const text = cell.textContent?.trim() || '';

    if ( !text )
    {
        showCopyFeedback( cell, 'Nothing to copy' );
        return;
    }

    // Try modern clipboard API first
    if ( navigator.clipboard && navigator.clipboard.writeText )
    {
        navigator.clipboard.writeText( text )
            .then( () => showCopyFeedback( cell, 'Copied!' ) )
            .catch( ( err ) =>
            {
                console.error( 'Failed to copy: ', err );
                fallbackCopyTextToClipboard( text, cell );
            } );
    } else
    {
        // Use fallback for older browsers
        fallbackCopyTextToClipboard( text, cell );
    }
}

/**
 * Show visual feedback when content is copied
 * @param {HTMLElement} cell - The cell element to show feedback on
 * @param {string} message - The message to display in the tooltip
 */
function showCopyFeedback ( cell, message )
{
    // Add copied class for CSS animation
    cell.classList.add( 'copied' );

    // Create and show tooltip
    const tooltip = createTooltip( message );
    cell.appendChild( tooltip );

    // Remove feedback after animation completes
    setTimeout( () =>
    {
        removeCopyFeedback( cell, tooltip );
    }, 600 );
}

/**
 * Create a tooltip element for copy feedback
 * @param {string} message - The message to display
 * @returns {HTMLDivElement} The created tooltip element
 */
function createTooltip ( message )
{
    const tooltip = document.createElement( 'div' );
    tooltip.className = 'copy-tooltip show';
    tooltip.textContent = message;
    return tooltip;
}

/**
 * Remove copy feedback animation and tooltip
 * @param {HTMLElement} cell - The cell element
 * @param {HTMLElement} tooltip - The tooltip element to remove
 */
function removeCopyFeedback ( cell, tooltip )
{
    cell.classList.remove( 'copied' );

    if ( tooltip.parentNode )
    {
        tooltip.parentNode.removeChild( tooltip );
    }
}

/**
 * Fallback clipboard copy for older browsers using document.execCommand
 * @param {string} text - The text to copy to clipboard
 * @param {HTMLElement} cell - The cell element for feedback
 */
function fallbackCopyTextToClipboard ( text, cell )
{
    const textArea = createHiddenTextArea( text );
    document.body.appendChild( textArea );

    // Select and copy the text
    textArea.focus();
    textArea.select();

    try
    {
        const successful = document.execCommand( 'copy' );
        const message = successful ? 'Copied!' : 'Copy failed';
        showCopyFeedback( cell, message );
    } catch ( err )
    {
        console.error( 'Fallback copy failed: ', err );
        showCopyFeedback( cell, 'Copy failed' );
    } finally
    {
        // Clean up the temporary textarea
        document.body.removeChild( textArea );
    }
}

/**
 * Create a hidden textarea element for fallback clipboard operations
 * @param {string} text - The text to put in the textarea
 * @returns {HTMLTextAreaElement} The created textarea element
 */
function createHiddenTextArea ( text )
{
    const textArea = document.createElement( 'textarea' );
    textArea.value = text;

    // Position off-screen
    textArea.style.position = 'fixed';
    textArea.style.left = '-999999px';
    textArea.style.top = '-999999px';
    textArea.style.opacity = '0';
    textArea.setAttribute( 'aria-hidden', 'true' );

    return textArea;
}

/**
 * Register service worker for PWA functionality
 * Handles offline caching and background sync
 */
function registerServiceWorker ()
{
    // Check if service workers are supported
    if ( 'serviceWorker' in navigator )
    {
        window.addEventListener( 'load', async () =>
        {
            try
            {
                // Register the service worker
                const registration = await navigator.serviceWorker.register( '/web-htmx/sw.js' );

                console.log( 'Service Worker registered successfully:', registration.scope );

                // Handle service worker updates
                registration.addEventListener( 'updatefound', () =>
                {
                    const newWorker = registration.installing;
                    console.log( 'New service worker installing...' );

                    newWorker.addEventListener( 'statechange', () =>
                    {
                        if ( newWorker.state === 'installed' && navigator.serviceWorker.controller )
                        {
                            // New service worker installed, show update notification
                            showUpdateNotification();
                        }
                    } );
                } );

                // Handle offline/online events
                setupOfflineDetection();

                // Listen for service worker messages
                setupServiceWorkerMessages();

            } catch ( error )
            {
                console.error( 'Service Worker registration failed:', error );
            }
        } );
    } else
    {
        console.warn( 'Service Workers are not supported in this browser' );
    }
}

/**
 * Show notification when app update is available
 */
function showUpdateNotification ()
{
    // Create update notification
    const notification = document.createElement( 'div' );
    notification.className = 'update-notification';
    notification.innerHTML = `
        <div class="update-content">
            <span>🔄 New version available!</span>
            <button onclick="window.location.reload()">Update</button>
            <button onclick="this.parentElement.parentElement.remove()">Later</button>
        </div>
    `;

    // Add to page
    document.body.appendChild( notification );

    // Auto-hide after 10 seconds
    setTimeout( () =>
    {
        if ( notification.parentNode )
        {
            notification.remove();
        }
    }, 10000 );
}

/**
 * Set up enhanced offline/online event detection with server health monitoring
 * Show detailed connection status to user and handle offline mode
 */
function setupOfflineDetection ()
{
    const statusElement = document.querySelector( '.status' );

    function updateConnectionStatus ()
    {
        const state = window.connectionState;

        if ( !navigator.onLine )
        {
            // Browser reports offline
            console.log( 'Browser is offline' );
            state.isOnline = false;
            state.offlineMode = true;
            updateStatusDisplay( statusElement, 'offline', 'No Internet Connection' );
            showOfflineMessage();
        }
        else if ( state.offlineMode )
        {
            // Browser came back online, but we need to check server
            console.log( 'Browser back online - checking server...' );
            state.isOnline = true;
            updateStatusDisplay( statusElement, 'checking', 'Reconnecting...' );
            checkServerHealth();
        }
        else
        {
            // Normal online state
            console.log( 'App is online' );
            state.isOnline = true;
            state.offlineMode = false;
            updateStatusDisplay( statusElement, 'online', 'Connected' );
            hideOfflineMessage();
        }
    }

    // Set initial status
    updateConnectionStatus();

    // Listen for connection changes
    window.addEventListener( 'online', updateConnectionStatus );
    window.addEventListener( 'offline', updateConnectionStatus );

    // Start periodic health checks
    startHealthChecks();
}

/**
 * Handle successful connection - reset retry counters and update state
 */
function handleSuccessfulConnection ()
{
    const state = window.connectionState;
    state.lastSuccessfulRequest = Date.now();
    state.retryCount = 0;
    state.isOnline = true;

    if ( state.offlineMode )
    {
        state.offlineMode = false;
        const statusElement = document.querySelector( '.status' );
        updateStatusDisplay( statusElement, 'online', 'Connected' );
        hideOfflineMessage();
        console.log( 'Connection restored!' );
    }
}

/**
 * Handle connection errors (4xx, 5xx responses)
 * @param {XMLHttpRequest} xhr - The failed request
 */
function handleConnectionError ( xhr )
{
    const state = window.connectionState;
    const timeSinceLastSuccess = Date.now() - state.lastSuccessfulRequest;

    // If it's been more than 30 seconds since last successful request, consider it a connection issue
    if ( timeSinceLastSuccess > 30000 && xhr.status >= 500 )
    {
        handleServerUnavailable();
    }
}

/**
 * Handle network errors (server completely unreachable)
 */
function handleNetworkError ()
{
    console.log( 'Network error - server may be offline' );
    handleServerUnavailable();
}

/**
 * Handle timeout errors
 */
function handleTimeoutError ()
{
    console.log( 'Request timeout - server may be slow or offline' );
    handleServerUnavailable();
}

/**
 * Handle server unavailable state
 */
function handleServerUnavailable ()
{
    const state = window.connectionState;
    const statusElement = document.querySelector( '.status' );

    if ( navigator.onLine && !state.offlineMode )
    {
        state.retryCount++;
        console.log( `Server unavailable (attempt ${ state.retryCount }/${ state.maxRetries })` );

        if ( state.retryCount >= state.maxRetries )
        {
            // Enter offline mode after max retries
            state.offlineMode = true;
            updateStatusDisplay( statusElement, 'server-offline', 'Server Offline' );
            showOfflineMessage( 'Server is offline. Please restart the banned-cli application.' );
        }
        else
        {
            // Show retry state briefly
            updateStatusDisplay( statusElement, 'retrying', 'Reconnecting...' );
            setTimeout( () => checkServerHealth(), state.retryInterval );
        }
    }
}

/**
 * Update the status display element
 * @param {HTMLElement} statusElement - The status display element
 * @param {string} state - The connection state (online, offline, checking, retrying, server-offline)
 * @param {string} message - The status message to display
 */
function updateStatusDisplay ( statusElement, state, message )
{
    if ( !statusElement ) return;

    // Remove all state classes
    statusElement.classList.remove( 'error', 'offline', 'checking', 'retrying', 'server-offline' );

    // Add appropriate class
    if ( state !== 'online' )
    {
        statusElement.classList.add( state );
    }

    // Update content
    statusElement.innerHTML = `<span class="indicator">●</span><span>${ message }</span>`;
}

/**
 * Start periodic health checks
 */
function startHealthChecks ()
{
    setInterval( () =>
    {
        const state = window.connectionState;
        const timeSinceLastSuccess = Date.now() - state.lastSuccessfulRequest;

        // If we haven't had a successful request in a while and we're not already checking
        if ( timeSinceLastSuccess > state.healthCheckInterval && !state.isHealthChecking && navigator.onLine )
        {
            checkServerHealth();
        }
    }, state.healthCheckInterval );
}

/**
 * Check server health with a lightweight request
 */
function checkServerHealth ()
{
    const state = window.connectionState;
    if ( state.isHealthChecking ) return;

    state.isHealthChecking = true;

    fetch( '/api/health', {
        method: 'GET',
        headers: { 'X-Health-Check': 'true' },
        cache: 'no-cache',
        timeout: 5000
    } )
        .then( response =>
        {
            if ( response.ok )
            {
                handleSuccessfulConnection();
            }
            else
            {
                throw new Error( `Health check failed: ${ response.status }` );
            }
        } )
        .catch( error =>
        {
            console.log( 'Health check failed:', error );
            handleServerUnavailable();
        } )
        .finally( () =>
        {
            state.isHealthChecking = false;
        } );
}

/**
 * Show offline notification banner
 * @param {string} customMessage - Optional custom message
 */
function showOfflineMessage ( customMessage )
{
    // Remove existing offline message
    hideOfflineMessage();

    const message = customMessage || 'Server is offline. Please restart the application.';

    const offlineBanner = document.createElement( 'div' );
    offlineBanner.id = 'offline-banner';
    offlineBanner.className = 'offline-banner';
    offlineBanner.innerHTML = `
        <div class="offline-banner-content">
            <span class="offline-banner-icon">⚠️</span>
            <span class="offline-banner-message">${ message }</span>
            <button onclick="window.location.reload()" class="offline-banner-button">Retry</button>
            <button onclick="hideOfflineMessage()" class="offline-banner-close">×</button>
        </div>
    `;

    document.body.appendChild( offlineBanner );

    // Show with animation
    setTimeout( () =>
    {
        offlineBanner.classList.add( 'show' );
    }, 100 );
}

/**
 * Hide offline notification banner
 */
function hideOfflineMessage ()
{
    const existingBanner = document.getElementById( 'offline-banner' );
    if ( existingBanner )
    {
        existingBanner.classList.remove( 'show' );
        setTimeout( () =>
        {
            if ( existingBanner.parentNode )
            {
                existingBanner.remove();
            }
        }, 300 );
    }
}

/**
 * Set up service worker message handling
 * Coordinate between service worker and main app connection state
 */
function setupServiceWorkerMessages ()
{
    if ( 'serviceWorker' in navigator )
    {
        navigator.serviceWorker.addEventListener( 'message', event =>
        {
            const { data } = event;

            if ( data && data.type === 'sw-event' )
            {
                console.log( 'Service Worker Event:', data.event );

                switch ( data.event )
                {
                    case 'connection-restored':
                        handleSuccessfulConnection();
                        break;

                    case 'connection-failed':
                        // Let the main app handle this through HTMX error events
                        // to avoid duplicate error handling
                        break;
                }
            }
        } );
    }
}