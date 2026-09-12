// ==============================================================================
// IP Monitor - Ponto de Entrada da Aplicação (App Entrypoint)
// Orquestra e inicializa todos os módulos
// ==============================================================================

import { initTheme, toggleTheme } from './components/theme.js';
import { loadIps, openAddModal, closeAddModal, handleModalSubmit } from './components/modal.js';
import { initSidebar } from './components/sidebar.js';
import { fetchAppVersion } from './services/api.js';
import { setupWailsEventListeners, startPeriodicFallbackSync } from './services/wailsEvents.js';

// Elementos de Conexão Geral
const addBtn = document.getElementById('addBtn');
const cancelAddBtn = document.getElementById('cancelAddBtn');
const modalCloseXBtn = document.getElementById('modalCloseXBtn');
const modalBackdrop = document.getElementById('modalBackdrop');
const themeToggleBtn = document.getElementById('themeToggleBtn');
const ipInput = document.getElementById('ipInput');
const nameInput = document.getElementById('nameInput');
const thresholdInput = document.getElementById('thresholdInput');

// Conexões de Eventos do Modal
if (addBtn) addBtn.addEventListener('click', handleModalSubmit);
if (cancelAddBtn) cancelAddBtn.addEventListener('click', closeAddModal);
if (modalCloseXBtn) modalCloseXBtn.addEventListener('click', closeAddModal);
if (modalBackdrop) modalBackdrop.addEventListener('click', closeAddModal);

// Teclas de atalho no formulário (Enter/Escape)
[ipInput, nameInput, thresholdInput].forEach(field => {
    if (!field) return;
    field.addEventListener('keydown', (e) => {
        if (e.key === 'Enter') handleModalSubmit();
        if (e.key === 'Escape') closeAddModal();
    });
});

// Alternador de tema
if (themeToggleBtn) {
    themeToggleBtn.addEventListener('click', toggleTheme);
}

// Carrega versão dinâmica no card "Sobre"
async function initVersion() {
    const badge = document.getElementById('appVersionBadge');
    if (!badge) return;
    const version = await fetchAppVersion();
    if (version) {
        badge.textContent = version;
    }
}

// Inicialização da Aplicação
function initApp() {
    initTheme();
    initSidebar();
    loadIps();
    initVersion();
    setupWailsEventListeners();
    startPeriodicFallbackSync();
}

// Aguarda o DOM estar pronto
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initApp);
} else {
    initApp();
}
