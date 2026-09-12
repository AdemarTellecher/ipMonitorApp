// ==============================================================================
// IP Monitor - Componente: Toast (Notificações Flutuantes)
// ==============================================================================

const toastEl = document.getElementById('toast');
let toastTimer = null;

export function showToast(message, duration = 2800) {
    if (!toastEl) return;
    
    if (toastTimer) {
        clearTimeout(toastTimer);
    }

    toastEl.textContent = message;
    toastEl.classList.remove('hidden');

    toastTimer = setTimeout(() => {
        toastEl.classList.add('hidden');
        toastTimer = null;
    }, duration);
}
