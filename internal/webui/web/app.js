import { render } from 'preact';
import { useState, useEffect } from 'preact/hooks';
import { html } from 'htm/preact';

// PWA Service Worker Registration and ESM Management
if ('serviceWorker' in navigator) {
    window.addEventListener('load', () => {
        navigator.serviceWorker.register('/web/sw.js')
            .then(registration => {
                console.log('✅ SW registered successfully:', registration.scope);
                
                // Check for updates
                registration.addEventListener('updatefound', () => {
                    const newWorker = registration.installing;
                    newWorker.addEventListener('statechange', () => {
                        if (newWorker.state === 'installed' && navigator.serviceWorker.controller) {
                            // Show update notification
                            showUpdateNotification();
                        }
                    });
                });
                
                // Pre-cache ESM imports when online
                if (navigator.onLine) {
                    precacheESMImports();
                }
            })
            .catch(error => {
                console.error('❌ SW registration failed:', error);
            });
    });
}

// Pre-cache ESM imports for offline use
async function precacheESMImports() {
    if ('caches' in window) {
        try {
            const cache = await caches.open('banned-cli-esm-v1');
            const esmUrls = [
                'https://esm.sh/preact@10.23.1',
                'https://esm.sh/preact@10.23.1/hooks',
                'https://esm.sh/htm@3.1.1/preact?external=preact'
            ];
            
            console.log('📦 Pre-caching ESM imports...');
            await cache.addAll(esmUrls);
            console.log('✅ ESM imports cached successfully');
        } catch (error) {
            console.warn('⚠️ Failed to pre-cache ESM imports:', error);
        }
    }
}

// PWA Install Prompt
let deferredPrompt;
window.addEventListener('beforeinstallprompt', (e) => {
    console.log('💡 PWA install prompt available');
    e.preventDefault();
    deferredPrompt = e;
    showInstallButton();
});

// PWA Installation Functions
function showInstallButton() {
    const installBtn = document.getElementById('install-btn');
    if (installBtn) {
        installBtn.style.display = 'block';
        installBtn.addEventListener('click', installPWA);
    }
}

async function installPWA() {
    if (deferredPrompt) {
        deferredPrompt.prompt();
        const { outcome } = await deferredPrompt.userChoice;
        console.log(`PWA install outcome: ${outcome}`);
        deferredPrompt = null;
        hideInstallButton();
    }
}

function hideInstallButton() {
    const installBtn = document.getElementById('install-btn');
    if (installBtn) {
        installBtn.style.display = 'none';
    }
}

function showUpdateNotification() {
    const notification = document.createElement('div');
    notification.className = 'update-notification';
    notification.innerHTML = `
        <div class="notification-content">
            <span>🔄 New version available!</span>
            <button onclick="updateApp()">Update</button>
            <button onclick="this.parentElement.parentElement.remove()">Later</button>
        </div>
    `;
    document.body.appendChild(notification);
}

window.updateApp = function() {
    if (navigator.serviceWorker.controller) {
        navigator.serviceWorker.controller.postMessage({ type: 'SKIP_WAITING' });
        window.location.reload();
    }
};

// PWA Background Sync
function requestBackgroundSync() {
    if ('serviceWorker' in navigator && 'sync' in window.ServiceWorkerRegistration.prototype) {
        navigator.serviceWorker.ready.then(registration => {
            return registration.sync.register('background-sync-data');
        }).catch(console.error);
    }
}

// Network Status Detection
function updateOnlineStatus() {
    const indicator = document.getElementById('online-status');
    if (indicator) {
        indicator.textContent = navigator.onLine ? '🟢 Online' : '🔴 Offline';
        indicator.className = navigator.onLine ? 'online' : 'offline';
    }
}

window.addEventListener('online', updateOnlineStatus);
window.addEventListener('offline', updateOnlineStatus);

// Helper function to get default search field for tables
function getDefaultSearchField ( tableName )
{
    const defaultFields = {
        'channels': 'name',
        'videos': 'title',
        'downloads': 'title'
    };
    return defaultFields[ tableName ] || 'name';
}

// API Client
const api = {
    async get ( endpoint )
    {
        const response = await fetch( `/api${ endpoint }` );
        if ( !response.ok )
        {
            throw new Error( `HTTP ${ response.status }` );
        }
        return response.json();
    },

    async post ( endpoint, data )
    {
        const response = await fetch( `/api${ endpoint }`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify( data ),
        } );
        if ( !response.ok )
        {
            throw new Error( `HTTP ${ response.status }` );
        }
        return response.json();
    }
};

// Components
function Header ( { status, stats } )
{
    return html`
        <header className="header">
            <div className="header-main">
                <h1>🚫 Banned CLI Dashboard</h1>
                <div className="header-controls">
                    <button id="install-btn" className="install-btn" style="display: none;">
                        📱 Install App
                    </button>
                    <button onclick=${() => requestBackgroundSync()} className="sync-btn" title="Sync Data">
                        🔄 Sync
                    </button>
                    <div id="online-status" className="online-status">
                        ${navigator.onLine ? '🟢 Online' : '🔴 Offline'}
                    </div>
                </div>
            </div>
            <div className="header-status">
                <div className=${ `status ${ status === 'connected' ? '' : 'error' }` }>
                    <span className="indicator">${ status === 'connected' ? '●' : '●' }</span>
                    <span>${ status === 'connected' ? 'Connected' : 'Disconnected' }</span>
                    ${ stats && html`<span> • ${ stats.tables } tables</span>` }
                </div>
            </div>
        </header>
    `;
}

function StatsSection ( { stats, loading } )
{
    if ( loading )
    {
        return html`
            <div className="stats-section">
                <h3>Database Statistics</h3>
                <div className="loading"></div>
            </div>
        `;
    }

    if ( !stats )
    {
        return null;
    }

    return html`
        <div className="stats-section">
            <h3>Database Statistics</h3>
            <div className="stats-grid">
                <div className="stat-item">
                    <div className="stat-number">${ stats.tables }</div>
                    <div className="stat-label">Tables</div>
                </div>
                <div className="stat-item">
                    <div className="stat-number">${ stats.totalRows?.toLocaleString() || '0' }</div>
                    <div className="stat-label">Total Rows</div>
                </div>
                <div className="stat-item">
                    <div className="stat-number">${ stats.sizeMB } MB</div>
                    <div className="stat-label">DB Size</div>
                </div>
                <div className="stat-item">
                    <div className="stat-number">${ stats.version }</div>
                    <div className="stat-label">SQLite</div>
                </div>
            </div>
            ${ stats.databasePath && html`
                <div style="margin-top: 1rem; padding: 0.5rem; background: #333; border-radius: 4px; font-size: 0.8rem;">
                    <div style="color: #aaa; margin-bottom: 0.25rem;">Database Location:</div>
                    <div style="color: #e0e0e0; word-break: break-all;">${ stats.databasePath }</div>
                    ${ stats.sizeBytes && html`
                        <div style="color: #aaa; margin-top: 0.25rem;">
                            File Size: ${ ( stats.sizeBytes / ( 1024 * 1024 ) ).toFixed( 2 ) } MB (${ stats.sizeBytes.toLocaleString() } bytes)
                        </div>
                    ` }
                </div>
            ` }
        </div>
    `;
}

function TablesList ( { tables, selectedTable, onTableSelect, tableCounts } )
{
    if ( !tables || tables.length === 0 )
    {
        return html`<div><p>No tables found</p></div>`;
    }

    return html`
        <div>
            <h3>Tables</h3>
            <ul className="tables-list">
                ${ tables.map( table => html`
                    <li
                        key=${ table }
                        className=${ `table-item ${ selectedTable === table ? 'active' : '' }` }
                        onClick=${ () => onTableSelect( table ) }
                    >
                        <span>${ table }</span>
                        <span
                            className=${ `table-count ${ tableCounts[ table ] === 0 ? 'empty' :
            tableCounts[ table ] > 0 ? 'populated' : 'error'
        }` }
                        >
                            ${ tableCounts[ table ] !== undefined ? tableCounts[ table ] : '?' }
                        </span>
                    </li>
                `) }
            </ul>
        </div>
    `;
}



function Sidebar ( { tables, selectedTable, onTableSelect, stats, statsLoading, tableCounts, collapsed, onToggle } )
{
    return html`
        <aside className=${ `sidebar ${ collapsed ? 'collapsed' : '' }` }>
            <button 
                className="sidebar-toggle" 
                onClick=${ onToggle }
                title=${ collapsed ? 'Expand sidebar' : 'Collapse sidebar' }
            >
                ${ collapsed ? '→' : '←' }
            </button>
            <div className="sidebar-content">
                <${ StatsSection } stats=${ stats } loading=${ statsLoading } />
                <${ TablesList } tables=${ tables } selectedTable=${ selectedTable } onTableSelect=${ onTableSelect } tableCounts=${ tableCounts } />
            </div>
        </aside>
    `;
}

function DataTable ( { data, loading, error } )
{
    if ( loading )
    {
        return html`<div className="loading"></div>`;
    }

    if ( error )
    {
        return html`<div className="error">Error: ${ error }</div>`;
    }

    if ( !data || data.length === 0 )
    {
        return html`
            <div className="placeholder">
                <h3>No Data</h3>
                <p>This table appears to be empty.</p>
                <p>You can add data using SQL INSERT statements.</p>
            </div>
        `;
    }

    // Get all columns and sort them: ID columns first, then title columns, then others
    const allColumns = Object.keys( data[ 0 ] );
    const idColumns = allColumns.filter( col =>
        col.toLowerCase() === 'id' ||
        col.toLowerCase().endsWith( '_id' ) ||
        col.toLowerCase().endsWith( 'id' )
    );
    const titleColumns = allColumns.filter( col =>
        !col.toLowerCase().includes( 'id' ) && (
            col.toLowerCase() === 'title' ||
            col.toLowerCase() === 'name' ||
            col.toLowerCase().includes( 'title' ) ||
            col.toLowerCase().includes( 'name' )
        )
    );
    const otherColumns = allColumns.filter( col =>
        !col.toLowerCase().includes( 'id' ) &&
        !col.toLowerCase().includes( 'title' ) &&
        !col.toLowerCase().includes( 'name' )
    );
    const columns = [ ...idColumns, ...titleColumns, ...otherColumns ];

    return html`
        <div className="table-wrapper">
            <table className="data-table">
                <thead>
                    <tr>
                        ${ columns.map( col => html`<th key=${ col }>${ col }</th>` ) }
                    </tr>
                </thead>
                <tbody>
                    ${ data.map( ( row, index ) => html`
                        <tr key=${ index }>
                            ${ columns.map( col => html`
                                <td key=${ col } title=${ String( row[ col ] ) }>
                                    ${ String( row[ col ] ) }
                                </td>
                            `) }
                        </tr>
                    `) }
                </tbody>
            </table>
        </div>
    `;
}

function TableViewer ( { selectedTable, tableData, tableLoading, tableError, onRefresh } )
{
    const [ searchTerm, setSearchTerm ] = useState( '' );
    const [ searchField, setSearchField ] = useState( getDefaultSearchField( selectedTable ) );
    const [ showSearch, setShowSearch ] = useState( true );
    const [ availableFields, setAvailableFields ] = useState( [] );
    const [ filteredData, setFilteredData ] = useState( [] );

    // Update available fields and reset search when table changes
    useEffect( () =>
    {
        if ( tableData && tableData.length > 0 )
        {
            const fields = Object.keys( tableData[ 0 ] );
            setAvailableFields( fields );
            setSearchField( getDefaultSearchField( selectedTable ) || fields[ 0 ] );
            setFilteredData( tableData );
        } else
        {
            setFilteredData( [] );
        }
        // Reset search state when table changes
        setSearchTerm( '' );
    }, [ selectedTable, tableData ] );

    // Update filtered data when search term or field changes
    useEffect( () =>
    {
        const filtered = performClientSearch( searchTerm, searchField );
        setFilteredData( filtered );
    }, [ searchTerm, searchField, tableData ] );

    const performClientSearch = ( term, field ) =>
    {
        if ( !tableData || tableData.length === 0 ) return [];

        if ( !term.trim() ) return tableData;

        return tableData.filter( row =>
        {
            const fieldValue = row[ field ];
            if ( fieldValue === null || fieldValue === undefined ) return false;

            return String( fieldValue ).toLowerCase().includes( term.toLowerCase() );
        } );
    };

    const handleSearch = () =>
    {
        const filtered = performClientSearch( searchTerm, searchField );
        setFilteredData( filtered );
    };

    const handleKeyDown = ( e ) =>
    {
        if ( e.key === 'Enter' )
        {
            e.preventDefault();
            handleSearch();
        }
    };

    const clearSearch = () =>
    {
        setSearchTerm( '' );
        setFilteredData( tableData || [] );
    };

    if ( !selectedTable )
    {
        return html`
            <div className="placeholder">
                <h3>Select a Table</h3>
                <p>Choose a table from the sidebar to view its data.</p>
            </div>
        `;
    }

    return html`
        <div>
            <div className="table-header">
                <h3>
                    Table: ${ selectedTable }
                    ${ searchTerm && html`
                        <span className="search-status"> - Showing ${ filteredData.length } of ${ tableData?.length || 0 } records${ searchTerm ? ` matching "${ searchTerm }"` : '' }</span>
                    ` }
                </h3>
                <div className="table-controls">
                    <button onClick=${ onRefresh }>Refresh</button>
                    <button onClick=${ () => setShowSearch( !showSearch ) }>
                        ${ showSearch ? 'Hide Search' : 'Show Search' }
                    </button>
                    ${ searchTerm && html`
                        <button onClick=${ clearSearch } className="clear-search-btn">
                            Clear Search
                        </button>
                    ` }
                </div>
            </div>
            
            ${ showSearch && html`
                <div className="search-section">
                    <div className="search-controls">
                        <div className="search-field-select">
                            <label>Search in:</label>
                            <select 
                                value=${ searchField } 
                                onChange=${ ( e ) => setSearchField( e.target.value ) }
                            >
                                ${ availableFields.map( field => html`
                                    <option key=${ field } value=${ field }>${ field }</option>
                                ` ) }
                            </select>
                        </div>
                        <div className="search-input-container">
                            <input
                                type="text"
                                placeholder="Enter search term..."
                                value=${ searchTerm }
                                onChange=${ ( e ) => setSearchTerm( e.target.value ) }
                            />

                            ${ searchTerm && html`
                                <button
                                    className="clear-btn"
                                    onClick=${ clearSearch }
                                >
                                    Clear
                                </button>
                            ` }
                        </div>
                    </div>
                </div>
            ` }
            
            <div className="data-section">
                ${ ( () =>
        {
            if ( tableLoading )
            {
                return html`<div className="loading">Loading...</div>`;
            }

            if ( tableError )
            {
                return html`<div className="error">Error: ${ tableError }</div>`;
            }

            if ( searchTerm && filteredData.length === 0 && tableData && tableData.length > 0 )
            {
                return html`
                            <div className="no-results">
                                <h3>No Results Found</h3>
                                <p>No results found for "${ searchTerm }" in ${ searchField }</p>
                                <button onClick=${ clearSearch } className="clear-search-btn">
                                    Show All Records
                                </button>
                            </div>
                        `;
            }

            return html`<${ DataTable } data=${ filteredData } loading=${ false } error=${ null } />`;
        } )() }
            </div>
        </div>
    `;
}

function SQLQueryEditor ( { onExecute, results, loading, error } )
{
    const [ query, setQuery ] = useState( 'SELECT * FROM channels LIMIT 10;' );

    const handleExecute = () =>
    {
        onExecute( query );
    };

    const handleKeyDown = ( e ) =>
    {
        if ( e.ctrlKey && e.key === 'Enter' )
        {
            e.preventDefault();
            handleExecute();
        }
    };

    return html`
        <div>
            <div style=${ { marginBottom: '1rem' } }>
                <textarea
                    id="query-input"
                    value=${ query }
                    onChange=${ ( e ) => setQuery( e.target.value ) }
                    onKeyDown=${ handleKeyDown }
                    placeholder="Enter SQL query..."
                />
                <button
                    className="execute-btn"
                    onClick=${ handleExecute }
                    disabled=${ loading }
                >
                    ${ loading ? 'Executing...' : 'Execute (Ctrl+Enter)' }
                </button>
            </div>
            <div className="query-results">
                ${ error && html`<div className="error">${ error }</div>` }
                ${ results && results.success && html`<div className="success">${ results.message }</div>` }
                ${ results && results.data && html`<${ DataTable } data=${ results.data } loading=${ false } />` }
            </div>
        </div>
    `;
}

function SchemaViewer ( { selectedTable, schema, loading, error } )
{
    if ( loading )
    {
        return html`<div className="loading"></div>`;
    }

    if ( error )
    {
        return html`<div className="error">Error: ${ error }</div>`;
    }

    if ( !selectedTable )
    {
        return html`
            <div className="placeholder">
                <h3>Select a Table</h3>
                <p>Choose a table from the sidebar to view its schema.</p>
            </div>
        `;
    }

    if ( !schema || schema.length === 0 )
    {
        return html`
            <div className="placeholder">
                <h3>No Schema Information</h3>
                <p>Unable to retrieve schema for this table.</p>
            </div>
        `;
    }

    return html`
        <div className="schema-container">
            <div className="schema-header">
                <h3>Table Schema: ${ selectedTable }</h3>
                <p className="schema-description">Fields and data types for records in this table</p>
            </div>
            
            <div className="schema-content">
                <div className="schema-table-wrapper">
                    <table className="schema-table">
                        <thead>
                            <tr>
                                <th className="col-name">Field Name</th>
                                <th className="col-type">Data Type</th>
                                <th className="col-constraints">Constraints</th>
                                <th className="col-default">Default Value</th>
                                <th className="col-description">Description</th>
                            </tr>
                        </thead>
                        <tbody>
                            ${ schema.map( ( col, index ) =>
    {
        const constraints = [];
        if ( col.primary ) constraints.push( 'PRIMARY KEY' );
        if ( col.notNull ) constraints.push( 'NOT NULL' );

        // Generate field descriptions based on common patterns
        const getFieldDescription = ( name, type ) =>
        {
            const lowerName = name.toLowerCase();
            if ( lowerName.includes( 'id' ) && col.primary ) return 'Unique identifier for this record';
            if ( lowerName === 'name' ) return 'Display name or title';
            if ( lowerName === 'title' ) return 'Title or heading text';
            if ( lowerName === 'description' ) return 'Detailed description or content';
            if ( lowerName === 'url' ) return 'Web address or link';
            if ( lowerName.includes( 'created' ) ) return 'Record creation timestamp';
            if ( lowerName.includes( 'updated' ) ) return 'Last modification timestamp';
            if ( lowerName.includes( 'date' ) || lowerName.includes( 'time' ) ) return 'Date/time information';
            if ( lowerName.includes( 'count' ) || lowerName.includes( 'num' ) ) return 'Numeric count or quantity';
            if ( type.toLowerCase().includes( 'text' ) ) return 'Text content';
            if ( type.toLowerCase().includes( 'int' ) ) return 'Integer number';
            if ( type.toLowerCase().includes( 'real' ) || type.toLowerCase().includes( 'float' ) ) return 'Decimal number';
            if ( type.toLowerCase().includes( 'blob' ) ) return 'Binary data';
            return 'Data field';
        };

        return html`
                                    <tr key=${ index } className="schema-row">
                                        <td className="field-name">
                                            <span className="field-name-text">${ col.name }</span>
                                            ${ col.primary ? html`<span className="primary-badge">PK</span>` : '' }
                                        </td>
                                        <td className="field-type">
                                            <span className="type-badge type-${ col.type.toLowerCase().replace( /[^a-z]/g, '' ) }">${ col.type }</span>
                                        </td>
                                        <td className="field-constraints">
                                            ${ constraints.length > 0 ?
                constraints.map( c => html`<span key=${ c } className="constraint-badge">${ c }</span>` ) :
                html`<span className="no-constraints">—</span>`
            }
                                        </td>
                                        <td className="field-default">
                                            ${ col.default ?
                html`<span className="default-value">${ col.default }</span>` :
                html`<span className="no-default">—</span>`
            }
                                        </td>
                                        <td className="field-description">
                                            <span className="description-text">${ getFieldDescription( col.name, col.type ) }</span>
                                        </td>
                                    </tr>
                                `;
    } ) }
                        </tbody>
                    </table>
                </div>
            </div>
        </div>
    `;
}

function TabContent ( { activeTab, selectedTable, tableData, tableLoading, tableError, onRefresh, schema, schemaLoading, schemaError } )
{
    switch ( activeTab )
    {
        case 'data':
            return html`<${ TableViewer }
                selectedTable=${ selectedTable }
                tableData=${ tableData }
                tableLoading=${ tableLoading }
                tableError=${ tableError }
                onRefresh=${ onRefresh }
            />`;
        case 'schema':
            return html`<${ SchemaViewer }
                selectedTable=${ selectedTable }
                schema=${ schema }
                loading=${ schemaLoading }
                error=${ schemaError }
            />`;
        default:
            return null;
    }
}

// Main App Component
function App ()
{
    const [ status, setStatus ] = useState( 'connecting' );
    const [ tables, setTables ] = useState( [] );
    const [ stats, setStats ] = useState( null );
    const [ statsLoading, setStatsLoading ] = useState( true );
    const [ selectedTable, setSelectedTable ] = useState( null );
    const [ activeTab, setActiveTab ] = useState( 'data' );
    const [ tableCounts, setTableCounts ] = useState( {} );
    const [ sidebarCollapsed, setSidebarCollapsed ] = useState( false );

    // Table data
    const [ tableData, setTableData ] = useState( [] );
    const [ tableLoading, setTableLoading ] = useState( false );
    const [ tableError, setTableError ] = useState( null );

    // Schema data
    const [ schema, setSchema ] = useState( [] );
    const [ schemaLoading, setSchemaLoading ] = useState( false );
    const [ schemaError, setSchemaError ] = useState( null );



    // Initialize app
    useEffect( () =>
    {
        loadInitialData();
    }, [] );

    // Load table data when selected table changes or when switching to data tab
    useEffect( () =>
    {
        if ( selectedTable && activeTab === 'data' )
        {
            loadTableData();
        }
    }, [ selectedTable, activeTab ] );

    // Load schema when selected table changes and schema tab is active
    useEffect( () =>
    {
        if ( selectedTable && activeTab === 'schema' )
        {
            loadSchema();
        }
    }, [ selectedTable, activeTab ] );

    const loadInitialData = async () =>
    {
        try
        {
            // Check health
            await api.get( '/health' );
            setStatus( 'connected' );

            // Load tables and stats
            const [ tablesData, statsData ] = await Promise.all( [
                api.get( '/tables' ),
                api.get( '/stats' )
            ] );

            setTables( tablesData.tables );
            setStats( statsData );
            setStatsLoading( false );

            // Use table counts from stats response
            if ( statsData.tableCounts )
            {
                setTableCounts( statsData.tableCounts );
            } else
            {
                // Fallback to individual queries if tableCounts not available
                const counts = {};
                for ( const table of tablesData.tables )
                {
                    try
                    {
                        const result = await api.get( `/data?table=${ table }&limit=0` );
                        counts[ table ] = result.count || 0;
                    } catch ( err )
                    {
                        console.warn( `Failed to get count for ${ table }:`, err );
                        counts[ table ] = '?';
                    }
                }
                setTableCounts( counts );
            }

        } catch ( error )
        {
            console.error( 'Failed to load initial data:', error );
            setStatus( 'error' );
            setStatsLoading( false );
        }
    };

    const loadTableData = async () =>
    {
        if ( !selectedTable ) return;

        setTableLoading( true );
        setTableError( null );

        try
        {
            const result = await api.get( `/data?table=${ selectedTable }&limit=100` );
            setTableData( result.data || [] );
        } catch ( error )
        {
            console.error( 'Failed to load table data:', error );
            setTableError( error.message );
        } finally
        {
            setTableLoading( false );
        }
    };

    const loadSchema = async () =>
    {
        if ( !selectedTable ) return;

        setSchemaLoading( true );
        setSchemaError( null );

        try
        {
            const result = await api.get( `/schema?table=${ selectedTable }` );
            setSchema( result.schema?.columns || [] );
        } catch ( error )
        {
            console.error( 'Failed to load schema:', error );
            setSchemaError( error.message );
        } finally
        {
            setSchemaLoading( false );
        }
    };

    const handleTableSelect = ( table ) =>
    {
        setSelectedTable( table );
        if ( activeTab === 'data' )
        {
            // Data will load via useEffect
        } else if ( activeTab === 'schema' )
        {
            // Schema will load via useEffect
        }
    };

    const handleTabChange = ( tab ) =>
    {
        setActiveTab( tab );

        // Trigger data loading for the new tab if a table is selected
        if ( selectedTable )
        {
            if ( tab === 'data' )
            {
                // Data will load via useEffect when activeTab changes
            } else if ( tab === 'schema' )
            {
                // Schema will load via useEffect when activeTab changes
            }
        }
    };

    const handleRefresh = () =>
    {
        if ( activeTab === 'data' )
        {
            loadTableData();
        } else if ( activeTab === 'schema' )
        {
            loadSchema();
        }
    };



    const toggleSidebar = () =>
    {
        setSidebarCollapsed( !sidebarCollapsed );
    };

    return html`
        <div>
            <${ Header } status=${ status } stats=${ stats } />
            <div className="layout">
                <${ Sidebar }
                    tables=${ tables }
                    selectedTable=${ selectedTable }
                    onTableSelect=${ handleTableSelect }
                    stats=${ stats }
                    statsLoading=${ statsLoading }
                    tableCounts=${ tableCounts }
                    collapsed=${ sidebarCollapsed }
                    onToggle=${ toggleSidebar }
                />
                <main className=${ `main-content ${ sidebarCollapsed ? 'sidebar-collapsed' : '' }` }>
                    <div className="tabs">
                        ${ [ 'data', 'schema' ].map( tab => html`
                            <button
                                key=${ tab }
                                className=${ `tab ${ activeTab === tab ? 'active' : '' }` }
                                onClick=${ () => handleTabChange( tab ) }
                            >
                                ${ tab.charAt( 0 ).toUpperCase() + tab.slice( 1 ) }
                            </button>
                        `) }
                    </div>
                    <div className="tab-content active">
                        <${ TabContent }
                            activeTab=${ activeTab }
                            selectedTable=${ selectedTable }
                            tableData=${ tableData }
                            tableLoading=${ tableLoading }
                            tableError=${ tableError }
                            onRefresh=${ handleRefresh }
                            schema=${ schema }
                            schemaLoading=${ schemaLoading }
                            schemaError=${ schemaError }
                        />
                    </div>
                </main>
            </div>
        </div>
    `;
}

// Initialize app
render( html`<${ App } />`, document.getElementById( 'app' ) );