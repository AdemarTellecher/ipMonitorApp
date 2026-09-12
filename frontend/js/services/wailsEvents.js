// ==============================================================================
// IP Monitor - Serviço: Wails v3 Events & Sincronização Periódica
// ==============================================================================

import { renderHosts } from '../components/hostList.js';
import { callGo } from './api.js';

// Listener para eventos de atualização de status emitidos pelo backend Go (Wails v3)
export function handleIpsUpdatedEvent(eventOrData) {
    console.log('[IP Monitor] Evento ips-updated recebido do backend:', eventOrData);
    const data = (eventOrData && eventOrData.data !== undefined) ? eventOrData.data : eventOrData;
    if (data) {
        renderHosts(data);
    }
}

// Registra listener no Wails Events com importação modular ES e fallback
export async function setupWailsEventListeners() {
    try {
        // No Wails v3, o runtime é um ES Module que exporta { Events, Application, Window }
        const runtime = await import('/wails/runtime.js');
        if (runtime && runtime.Events && typeof runtime.Events.On === 'function') {
            runtime.Events.On('ips-updated', handleIpsUpdatedEvent);
            console.log('[IP Monitor] Listener Wails v3 (ES Module) para "ips-updated" registrado com sucesso.');
            return;
        }
    } catch (e) {
        console.log('[IP Monitor] Importação direta de /wails/runtime.js em andamento via fallback...');
    }

    // Fallback para caso o runtime defina window.wails globalmente
    if (window.wails && window.wails.Events && typeof window.wails.Events.On === 'function') {
        window.wails.Events.On('ips-updated', handleIpsUpdatedEvent);
        console.log('[IP Monitor] Listener Wails v3 (window.wails) para "ips-updated" registrado com sucesso.');
    } else {
        setTimeout(() => {
            if (window.wails && window.wails.Events && typeof window.wails.Events.On === 'function') {
                window.wails.Events.On('ips-updated', handleIpsUpdatedEvent);
                console.log('[IP Monitor] Listener Wails v3 registrado após retry.');
            }
        }, 500);
    }
}

// Fallback de sincronização periódica a cada 60 segundos
export function startPeriodicFallbackSync() {
    setInterval(async () => {
        try {
            const overview = await callGo('list');
            if (overview) {
                renderHosts(overview);
            }
        } catch (e) {
            console.warn('[IP Monitor] Fallback periódico de sincronização falhou:', e);
        }
    }, 60000);
}
