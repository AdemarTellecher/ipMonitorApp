// ==============================================================================
// IP Monitor - Ponto de Entrada da Aplicação (App Entrypoint)
// Orquestra e inicializa todos os módulos
// ==============================================================================

import { initTheme, toggleTheme } from './components/theme.js';
import { loadIps, initModals } from './components/modal.js';
import { initSidebar } from './components/sidebar.js';
import { setupWailsEventListeners, startPeriodicFallbackSync } from './services/wailsEvents.js';

// Alternador de tema
const themeToggleBtn = document.getElementById('themeToggleBtn');
if (themeToggleBtn) {
    themeToggleBtn.addEventListener('click', toggleTheme);
}

// Inicialização da Aplicação
function initApp() {
    initTheme();
    initModals();
    initSidebar();
    loadIps();
    setupWailsEventListeners();
    startPeriodicFallbackSync();
}

// Aguarda o DOM estar pronto
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initApp);
} else {
    initApp();
}
