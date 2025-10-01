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
        }
    );

    // Handle HTMX errors globally
    document.body.addEventListener( 'htmx:responseError',
        /** @param {HTMXErrorEvent} evt - HTMX error event */
        function ( evt )
        {
            console.error( 'HTMX Error:', evt.detail );
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
 * Set up offline/online event detection
 * Show connection status to user
 */
function setupOfflineDetection ()
{
    const statusElement = document.querySelector( '.status' );

    function updateConnectionStatus ()
    {
        if ( navigator.onLine )
        {
            console.log( 'App is online' );
            if ( statusElement )
            {
                statusElement.classList.remove( 'error' );
                statusElement.innerHTML = '<span class="indicator">●</span><span>Connected</span>';
            }
        } else
        {
            console.log( 'App is offline' );
            if ( statusElement )
            {
                statusElement.classList.add( 'error' );
                statusElement.innerHTML = '<span class="indicator">●</span><span>Offline</span>';
            }
        }
    }

    // Set initial status
    updateConnectionStatus();

    // Listen for connection changes
    window.addEventListener( 'online', updateConnectionStatus );
    window.addEventListener( 'offline', updateConnectionStatus );
}