// ==============================================================================
// IP Monitor - Componente: Sidebar (Ações, Varredura, Remoção, Importação)
// ==============================================================================

import { state, setSelectedDevice } from '../state.js';
import { callGo } from '../services/api.js';
import { showToast } from './toast.js';
import { renderHosts, updateActionButtonsState } from './hostList.js';
import { openAddModal, openEditModal, closeAddModal, loadIps } from './modal.js';

const navFocusAdd = document.getElementById('navFocusAdd');
const navImport = document.getElementById('navImport');
const navRefresh = document.getElementById('navRefresh');
const navEdit = document.getElementById('navEdit');
const navRemove = document.getElementById('navRemove');
const fileInput = document.getElementById('fileInput');
const addIpModal = document.getElementById('addIpModal');

export async function handleRefresh() {
    if (!navRefresh) return;
    try {
        navRefresh.classList.add('spinning');
        showToast('Executando varredura ICMP Ping...');

        const updated = await callGo('update');
        renderHosts(updated);
        showToast('Varredura ICMP Ping concluída!');
    } catch (err) {
        showToast('Falha ao atualizar status dos hosts');
    } finally {
        navRefresh.classList.remove('spinning');
    }
}

export async function handleRemove() {
    if (!state.selectedDevice) return;
    const ipToRemove = state.selectedDevice.ip;

    try {
        if (navRemove) navRemove.classList.add('disabled');
        if (navEdit) navEdit.classList.add('disabled');
        const result = await callGo('remove', { ip: ipToRemove });
        if (result && result.error) {
            showToast(result.error);
        } else {
            showToast(`Host ${ipToRemove} removido`);
            setSelectedDevice(null);
            await loadIps();
        }
    } catch (err) {
        showToast('Erro ao remover host');
    } finally {
        updateActionButtonsState();
    }
}

export async function handleFileSelected(event) {
    const file = event.target.files[0];
    if (!file) return;

    try {
        const text = await file.text();
        const result = await callGo('import', text);
        if (result && result.success) {
            showToast(`Importação concluída: ${result.count || 0} hosts adicionados!`);
            await loadIps();
        } else {
            showToast(result.error || 'Erro na importação de JSON');
        }
    } catch (err) {
        showToast('Falha ao processar arquivo JSON');
    } finally {
        if (fileInput) fileInput.value = '';
    }
}

export function initSidebar() {
    if (navFocusAdd) {
        navFocusAdd.addEventListener('click', () => {
            if (addIpModal && addIpModal.classList.contains('hidden')) {
                openAddModal();
            } else {
                closeAddModal();
            }
        });
    }

    if (navImport && fileInput) {
        navImport.addEventListener('click', () => fileInput.click());
        fileInput.addEventListener('change', handleFileSelected);
    }

    if (navRefresh) {
        navRefresh.addEventListener('click', handleRefresh);
    }

    if (navEdit) {
        navEdit.addEventListener('click', () => {
            if (state.selectedDevice) {
                openEditModal();
            }
        });
    }

    if (navRemove) {
        navRemove.addEventListener('click', handleRemove);
    }
}
