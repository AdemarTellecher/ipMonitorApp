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

// Registra listener no Wails Events com retry caso o runtime carregue de forma assíncrona
export function setupWailsEventListeners() {
    if (window.wails && window.wails.Events && typeof window.wails.Events.On === 'function') {
        window.wails.Events.On('ips-updated', handleIpsUpdatedEvent);
        console.log('[IP Monitor] Listener Wails v3 para "ips-updated" registrado com sucesso.');
    } else {
        setTimeout(() => {
            if (window.wails && window.wails.Events && typeof window.wails.Events.On === 'function') {
                window.wails.Events.On('ips-updated', handleIpsUpdatedEvent);
                console.log('[IP Monitor] Listener Wails v3 para "ips-updated" registrado após retry.');
            } else {
                console.warn('[IP Monitor] Wails runtime não detectado para eventos nativos; operando com fallback.');
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
