// ==============================================================================
// IP Monitor - Serviço de Comunicação HTTP com o Backend Go (/api/...)
// ==============================================================================

export async function callGo(endpoint, data = null) {
    try {
        const options = {
            method: data ? 'POST' : 'GET',
            headers: { 'Content-Type': 'application/json' },
        };
        if (data) {
            options.body = (typeof data === 'string') ? data : JSON.stringify(data);
        }
        const response = await fetch(`/api/${endpoint}`, options);
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${await response.text()}`);
        }
        return await response.json();
    } catch (err) {
        console.error(`Erro ao chamar /api/${endpoint}:`, err);
        throw err;
    }
}

export async function fetchAppVersion() {
    try {
        const res = await fetch('/api/version');
        if (res.ok) {
            const data = await res.json();
            return data && data.version ? data.version : null;
        }
    } catch (e) {
        console.warn('Não foi possível buscar a versão via API:', e);
    }
    return null;
}
