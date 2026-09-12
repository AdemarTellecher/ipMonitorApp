// ==============================================================================
// IP Monitor - Componente: Modal (Cadastro e Edição de Hosts)
// ==============================================================================

import { state, setModalMode, setSelectedDevice } from '../state.js';
import { callGo } from '../services/api.js';
import { showToast } from './toast.js';
import { renderHosts } from './hostList.js';

// Elementos do Modal
const addIpModal = document.getElementById('addIpModal');
const modalTitle = document.getElementById('modalTitle');
const modalDesc = document.getElementById('modalDesc');
const modalConfirmLabel = document.getElementById('modalConfirmLabel');
const modalBackdrop = document.getElementById('modalBackdrop');
const networkOverviewCard = document.getElementById('networkOverviewCard');
const ipInput = document.getElementById('ipInput');
const nameInput = document.getElementById('nameInput');
const methodSelect = document.getElementById('methodSelect');
const thresholdInput = document.getElementById('thresholdInput');
const addBtn = document.getElementById('addBtn');

// Elementos da Sidebar com estado ativo durante modal
const navFocusAdd = document.getElementById('navFocusAdd');
const navEdit = document.getElementById('navEdit');

export async function loadIps() {
    try {
        const ips = await callGo('list');
        renderHosts(ips);
    } catch (err) {
        showToast('Erro ao carregar lista de IPs');
    }
}

export function openAddModal() {
    setModalMode('add');
    modalTitle.textContent = 'Cadastrar Novo Dispositivo';
    modalDesc.textContent = 'Preencha os dados do dispositivo para o monitoramento contínuo.';
    modalConfirmLabel.textContent = 'Cadastrar';
    ipInput.value = '';
    nameInput.value = '';
    methodSelect.value = 'PING';
    thresholdInput.value = '2000';

    modalBackdrop.classList.remove('hidden');
    if (networkOverviewCard) networkOverviewCard.classList.add('has-modal');
    addIpModal.classList.remove('hidden');
    if (navFocusAdd) navFocusAdd.classList.add('active');

    setTimeout(() => {
        ipInput.focus();
        ipInput.select();
    }, 50);
}

export function openEditModal() {
    if (!state.selectedDevice) {
        showToast('Selecione um host na lista para editar');
        return;
    }
    setModalMode('edit');
    modalTitle.textContent = 'Editar Dispositivo / Host';
    modalDesc.textContent = 'Atualize as informações do dispositivo selecionado.';
    modalConfirmLabel.textContent = 'Salvar';
    ipInput.value = state.selectedDevice.ip || '';
    nameInput.value = state.selectedDevice.name || '';
    methodSelect.value = state.selectedDevice.method || 'PING';
    thresholdInput.value = state.selectedDevice.thresholdMs ? state.selectedDevice.thresholdMs : '2000';

    modalBackdrop.classList.remove('hidden');
    if (networkOverviewCard) networkOverviewCard.classList.add('has-modal');
    addIpModal.classList.remove('hidden');
    if (navEdit) navEdit.classList.add('active');

    setTimeout(() => {
        ipInput.focus();
        ipInput.select();
    }, 50);
}

export function closeAddModal() {
    addIpModal.classList.add('hidden');
    if (networkOverviewCard) networkOverviewCard.classList.remove('has-modal');
    modalBackdrop.classList.add('hidden');
    if (navFocusAdd) navFocusAdd.classList.remove('active');
    if (navEdit) navEdit.classList.remove('active');
    ipInput.value = '';
    nameInput.value = '';
    methodSelect.value = 'PING';
    thresholdInput.value = '2000';
    setModalMode('add');
}

export async function handleModalSubmit() {
    const rawIp = ipInput.value.trim();
    if (!rawIp) {
        ipInput.focus();
        return;
    }

    const nameVal = nameInput.value.trim();
    const methodVal = methodSelect.value || 'PING';
    const thresholdVal = parseInt(thresholdInput.value, 10) || 2000;

    try {
        addBtn.disabled = true;
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
        addBtn.disabled = false;
    }
}
