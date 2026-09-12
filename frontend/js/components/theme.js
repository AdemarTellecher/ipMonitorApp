// ==============================================================================
// IP Monitor - Componente: Theme (Tema Claro / Escuro)
// ==============================================================================

const themeLabel = document.getElementById('themeLabel');
const darkIcon = document.getElementById('themeIconDark');
const lightIcon = document.getElementById('themeIconLight');

export function updateThemeUI(isLight) {
    if (isLight) {
        if (themeLabel) themeLabel.textContent = 'Tema Escuro';
        if (darkIcon) darkIcon.classList.remove('hidden');
        if (lightIcon) lightIcon.classList.add('hidden');
    } else {
        if (themeLabel) themeLabel.textContent = 'Tema Claro';
        if (darkIcon) darkIcon.classList.add('hidden');
        if (lightIcon) lightIcon.classList.remove('hidden');
    }
}

export function toggleTheme() {
    const isLight = document.body.classList.toggle('light-theme');
    localStorage.setItem('ipmonitor_theme', isLight ? 'light' : 'dark');
    updateThemeUI(isLight);
}

export function initTheme() {
    const saved = localStorage.getItem('ipmonitor_theme');
    const isLight = (saved === 'light');
    if (isLight) {
        document.body.classList.add('light-theme');
    } else {
        document.body.classList.remove('light-theme');
    }
    updateThemeUI(isLight);
}
