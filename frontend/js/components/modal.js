// ==============================================================================
// IP Monitor - Componente: Modais (Cadastro/Edição, Importação e Sobre)
// ==============================================================================

import { state, setModalMode, setSelectedDevice } from '../state.js';
import { callGo, fetchAppVersion } from '../services/api.js';
import { showToast } from './toast.js';
import { renderHosts } from './hostList.js';

// Backdrop Global
const modalBackdrop = document.getElementById('modalBackdrop');
const networkOverviewCard = document.getElementById('networkOverviewCard');

// Modal 1: Cadastro / Edição
const addIpModal = document.getElementById('addIpModal');
const modalTitle = document.getElementById('modalTitle');
const modalConfirmLabel = document.getElementById('modalConfirmLabel');
const ipInput = document.getElementById('ipInput');
const nameInput = document.getElementById('nameInput');
const methodSelect = document.getElementById('methodSelect');
const thresholdInput = document.getElementById('thresholdInput');
const addBtn = document.getElementById('addBtn');
const cancelAddBtn = document.getElementById('cancelAddBtn');
const modalCloseXBtn = document.getElementById('modalCloseXBtn');

// Modal 2: Importar (.JSON)
const importModal = document.getElementById('importModal');
const importPickFileBtn = document.getElementById('importPickFileBtn');
const fileInput = document.getElementById('fileInput');
const importJsonTextarea = document.getElementById('importJsonTextarea');
const cancelImportBtn = document.getElementById('cancelImportBtn');
const confirmImportBtn = document.getElementById('confirmImportBtn');
const importCloseXBtn = document.getElementById('importCloseXBtn');

// Modal 3: Sobre
const aboutModal = document.getElementById('aboutModal');
const aboutVersionText = document.getElementById('aboutVersionText');
const aboutCloseXBtn = document.getElementById('aboutCloseXBtn');

// Elementos da Sidebar com estado ativo
const navFocusAdd = document.getElementById('navFocusAdd');
const navEdit = document.getElementById('navEdit');
const navImport = document.getElementById('navImport');
const navAbout = document.getElementById('navAbout');

export async function loadIps() {
    try {
        const ips = await callGo('list');
        renderHosts(ips);
    } catch (err) {
        showToast('Erro ao carregar lista de IPs');
    }
}

// -----------------------------------------------------------------------------
// Funções do Modal 1: Cadastro / Edição
// -----------------------------------------------------------------------------
export function openAddModal() {
    closeAllModals();
    setModalMode('add');
    if (modalTitle) modalTitle.textContent = 'Novo Dispositivo';
    if (modalConfirmLabel) modalConfirmLabel.textContent = 'Cadastrar';
    if (ipInput) ipInput.value = '';
    if (nameInput) nameInput.value = '';
    if (methodSelect) methodSelect.value = 'PING';
    if (thresholdInput) thresholdInput.value = '2000';

    if (modalBackdrop) modalBackdrop.classList.remove('hidden');
    if (networkOverviewCard) networkOverviewCard.classList.add('has-modal');
    if (addIpModal) addIpModal.classList.remove('hidden');
    if (navFocusAdd) navFocusAdd.classList.add('active');

    setTimeout(() => {
        if (ipInput) {
            ipInput.focus();
            ipInput.select();
        }
    }, 50);
}

export function openEditModal() {
    if (!state.selectedDevice) {
        showToast('Selecione um host na lista para editar');
        return;
    }
    closeAllModals();
    setModalMode('edit');
    if (modalTitle) modalTitle.textContent = 'Editar Dispositivo / Host';
    if (modalConfirmLabel) modalConfirmLabel.textContent = 'Salvar';
    if (ipInput) ipInput.value = state.selectedDevice.ip || '';
    if (nameInput) nameInput.value = state.selectedDevice.name || '';
    if (methodSelect) methodSelect.value = state.selectedDevice.method || 'PING';
    if (thresholdInput) thresholdInput.value = state.selectedDevice.thresholdMs ? state.selectedDevice.thresholdMs : '2000';

    if (modalBackdrop) modalBackdrop.classList.remove('hidden');
    if (networkOverviewCard) networkOverviewCard.classList.add('has-modal');
    if (addIpModal) addIpModal.classList.remove('hidden');
    if (navEdit) navEdit.classList.add('active');

    setTimeout(() => {
        if (ipInput) {
            ipInput.focus();
            ipInput.select();
        }
    }, 50);
}

export function closeAddModal() {
    if (addIpModal) addIpModal.classList.add('hidden');
    if (networkOverviewCard) networkOverviewCard.classList.remove('has-modal');
    if (modalBackdrop) modalBackdrop.classList.add('hidden');
    if (navFocusAdd) navFocusAdd.classList.remove('active');
    if (navEdit) navEdit.classList.remove('active');
    if (ipInput) ipInput.value = '';
    if (nameInput) nameInput.value = '';
    if (methodSelect) methodSelect.value = 'PING';
    if (thresholdInput) thresholdInput.value = '2000';
    setModalMode('add');
}

export async function handleModalSubmit() {
    if (!ipInput) return;
    const rawIp = ipInput.value.trim();
    if (!rawIp) {
        ipInput.focus();
        return;
    }

    const nameVal = nameInput ? nameInput.value.trim() : '';
    const methodVal = methodSelect ? methodSelect.value || 'PING' : 'PING';
    const thresholdVal = thresholdInput ? parseInt(thresholdInput.value, 10) || 2000 : 2000;

    try {
        if (addBtn) addBtn.disabled = true;
        if (state.modalMode === 'add') {
            const result = await callGo('add', {
                ip: rawIp,
                name: nameVal,
                method: methodVal,
                thresholdMs: thresholdVal,
                uuid: '',
            });
            if (result && result.error) {
                showToast(result.error);
            } else {
                closeAddModal();
                showToast(`Host ${rawIp} cadastrado com sucesso!`);
                await loadIps();
            }
        } else if (state.modalMode === 'edit') {
            if (!state.selectedDevice) return;
            const result = await callGo('edit', {
                id: state.selectedDevice.id,
                newIp: rawIp,
                name: nameVal,
                method: methodVal,
                thresholdMs: thresholdVal,
                uuid: state.selectedDevice.uuid || '',
            });
            if (result && result.error) {
                showToast(result.error);
            } else {
                closeAddModal();
                showToast(`Host alterado para ${rawIp} com sucesso!`);
                setSelectedDevice(null);
                await loadIps();
            }
        }
    } catch (err) {
        showToast(state.modalMode === 'edit' ? 'Erro ao editar host' : 'Erro ao cadastrar host');
    } finally {
        if (addBtn) addBtn.disabled = false;
    }
}

// -----------------------------------------------------------------------------
// Funções do Modal 2: Importar (.JSON)
// -----------------------------------------------------------------------------
export function openImportModal() {
    closeAllModals();
    if (importJsonTextarea) importJsonTextarea.value = '';
    if (modalBackdrop) modalBackdrop.classList.remove('hidden');
    if (networkOverviewCard) networkOverviewCard.classList.add('has-modal');
    if (importModal) importModal.classList.remove('hidden');
    if (navImport) navImport.classList.add('active');
}

export function closeImportModal() {
    if (importModal) importModal.classList.add('hidden');
    if (networkOverviewCard) networkOverviewCard.classList.remove('has-modal');
    if (modalBackdrop) modalBackdrop.classList.add('hidden');
    if (navImport) navImport.classList.remove('active');
    if (importJsonTextarea) importJsonTextarea.value = '';
}

export async function handleConfirmImport() {
    if (!importJsonTextarea) return;
    const content = importJsonTextarea.value.trim();
    if (!content) {
        showToast('Cole o conteúdo JSON ou selecione um arquivo');
        importJsonTextarea.focus();
        return;
    }

    try {
        if (confirmImportBtn) confirmImportBtn.disabled = true;
        const result = await callGo('import', content);
        if (result && result.success) {
            closeImportModal();
            showToast(`Importação concluída: ${result.count || 0} hosts adicionados!`);
            await loadIps();
        } else {
            showToast(result.error || 'Erro na importação de JSON');
        }
    } catch (err) {
        showToast('Falha ao processar arquivo JSON');
    } finally {
        if (confirmImportBtn) confirmImportBtn.disabled = false;
    }
}

// -----------------------------------------------------------------------------
// Funções do Modal 3: Sobre o IP Monitor
// -----------------------------------------------------------------------------
export async function openAboutModal() {
    closeAllModals();
    const version = await fetchAppVersion();
    if (aboutVersionText && version) {
        aboutVersionText.textContent = version;
    }
    if (modalBackdrop) modalBackdrop.classList.remove('hidden');
    if (networkOverviewCard) networkOverviewCard.classList.add('has-modal');
    if (aboutModal) aboutModal.classList.remove('hidden');
    if (navAbout) navAbout.classList.add('active');
}

export function closeAboutModal() {
    if (aboutModal) aboutModal.classList.add('hidden');
    if (networkOverviewCard) networkOverviewCard.classList.remove('has-modal');
    if (modalBackdrop) modalBackdrop.classList.add('hidden');
    if (navAbout) navAbout.classList.remove('active');
}

// Fecha qualquer modal que esteja aberto
export function closeAllModals() {
    closeAddModal();
    closeImportModal();
    closeAboutModal();
}

// -----------------------------------------------------------------------------
// Inicialização de Listeners dos Modais
// -----------------------------------------------------------------------------
export function initModals() {
    // Modal 1: Cadastro / Edição
    if (addBtn) addBtn.addEventListener('click', handleModalSubmit);
    if (cancelAddBtn) cancelAddBtn.addEventListener('click', closeAddModal);
    if (modalCloseXBtn) modalCloseXBtn.addEventListener('click', closeAddModal);

    [ipInput, nameInput, thresholdInput].forEach(field => {
        if (!field) return;
        field.addEventListener('keydown', (e) => {
            if (e.key === 'Enter') handleModalSubmit();
            if (e.key === 'Escape') closeAddModal();
        });
    });

    // Modal 2: Importar (.JSON)
    if (importPickFileBtn && fileInput) {
        importPickFileBtn.addEventListener('click', () => fileInput.click());
    }
    if (fileInput) {
        fileInput.addEventListener('change', async (event) => {
            const file = event.target.files[0];
            if (!file) return;
            try {
                const text = await file.text();
                if (importJsonTextarea) {
                    importJsonTextarea.value = text;
                }
            } catch (e) {
                showToast('Falha ao ler o arquivo selecionado');
            } finally {
                fileInput.value = '';
            }
        });
    }
    if (cancelImportBtn) cancelImportBtn.addEventListener('click', closeImportModal);
    if (confirmImportBtn) confirmImportBtn.addEventListener('click', handleConfirmImport);
    if (importCloseXBtn) importCloseXBtn.addEventListener('click', closeImportModal);

    // Modal 3: Sobre
    if (aboutCloseXBtn) aboutCloseXBtn.addEventListener('click', closeAboutModal);

    // Backdrop Global: fecha qualquer modal ativo
    if (modalBackdrop) {
        modalBackdrop.addEventListener('click', closeAllModals);
    }
}
