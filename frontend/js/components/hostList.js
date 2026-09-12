// ==============================================================================
// IP Monitor - Componente: Host List & Cards
// ==============================================================================

import { state, setSelectedDevice, setCurrentIps, toggleExpandedHostId } from '../state.js';
import { openEditModal } from './modal.js';

// Elementos da UI
const metricTotal = document.getElementById('metricTotal');
const metricOnline = document.getElementById('metricOnline');
const metricOffline = document.getElementById('metricOffline');
const hostsList = document.getElementById('hostsList');

// Elementos da Sidebar com dependência de seleção
const navEdit = document.getElementById('navEdit');
const navRemove = document.getElementById('navRemove');

// Atualiza o estado dos botões de ação sensíveis à seleção (Editar e Remover)
export function updateActionButtonsState() {
    const hasSelection = Boolean(state.selectedDevice);
    if (navEdit && navRemove) {
        if (hasSelection) {
            navEdit.classList.remove('disabled');
            navRemove.classList.remove('disabled');
        } else {
            navEdit.classList.add('disabled');
            navRemove.classList.add('disabled');
        }
    }
}

// Atualiza visualmente a classe 'selected' nos cards existentes sem recriar o DOM
export function updateSelectedRowUI() {
    if (!hostsList) return;
    const cards = hostsList.querySelectorAll('.host-item-card');
    cards.forEach(c => {
        const rowId = parseInt(c.dataset.id, 10);
        if (state.selectedDevice && rowId === state.selectedDevice.id) {
            c.classList.add('selected');
        } else {
            c.classList.remove('selected');
        }
    });
    updateActionButtonsState();
}

// Seleciona um dispositivo para ação mantendo o DOM estável
export function selectDevice(device) {
    if (state.selectedDevice && state.selectedDevice.id === device.id) {
        setSelectedDevice(null);
    } else {
        setSelectedDevice(device);
    }
    updateSelectedRowUI();
}

// Alterna o estado expandido/recolhido de um host
export function toggleExpandHost(event, deviceId) {
    if (event) {
        event.stopPropagation();
    }
    const card = document.getElementById(`host-card-${deviceId}`);
    const isNowExpanded = toggleExpandedHostId(deviceId);
    if (card) {
        if (isNowExpanded) {
            card.classList.add('is-expanded');
        } else {
            card.classList.remove('is-expanded');
        }
    }
}

// Renderiza a lista de hosts e o sumário de rede com dados entregues pelo Go
export function renderHosts(data) {
    if (!hostsList) return;

    let devices = [];
    let total = 0;
    let online = 0;
    let offline = 0;

    if (data && Array.isArray(data.devices)) {
        devices = data.devices;
        total = data.total;
        online = data.online;
        offline = data.offline;
    } else if (Array.isArray(data)) {
        devices = data;
        total = devices.length;
        online = devices.filter(d => d.status === 'Online').length;
        offline = devices.filter(d => d.status === 'Offline').length;
    }

    setCurrentIps(devices);
    hostsList.innerHTML = '';

    if (devices.length === 0) {
        hostsList.innerHTML = `<div style="text-align:center; padding: 30px 0; color: var(--text-dim); font-size: 0.8rem;">Nenhum dispositivo na lista.<br>Cadastre um IP acima ou importe um JSON.</div>`;
    } else {
        devices.forEach(device => {
            const isOnline = device.status === 'Online';
            const isOffline = device.status === 'Offline';

            const isSelected = state.selectedDevice && state.selectedDevice.id === device.id;
            const isExpanded = state.expandedHostIds.has(device.id);

            const card = document.createElement('div');
            card.className = `host-item-card ${isSelected ? 'selected' : ''} ${isExpanded ? 'is-expanded' : ''}`;
            card.id = `host-card-${device.id}`;
            card.dataset.id = device.id;

            const badgeClass = isOnline ? 'online' : (isOffline ? 'offline' : 'unknown');
            const hostName = device.name ? device.name : 'Não especificado (Host manual)';
            const hostMethod = device.method ? device.method : 'PING';
            const thresholdMs = device.thresholdMs ? `${device.thresholdMs} ms` : '2000 ms';
            const hostUUID = device.uuid ? device.uuid : '—';

            card.innerHTML = `
                <div class="host-main-row">
                    <div class="host-ip-col">
                        <button class="host-expand-btn" title="Expandir/Recolher Detalhes" type="button">
                            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
                                <polyline points="9 18 15 12 9 6"></polyline>
                            </svg>
                        </button>
                        <span>${device.ip}</span>
                    </div>
                    <div>
                        <span class="status-badge ${badgeClass}">
                            <span class="status-pip"></span>
                            ${device.status}
                        </span>
                    </div>
                </div>
                <div class="host-details-drawer">
                    <div class="details-grid">
                        <div class="detail-item full-width">
                            <span class="detail-label">NOME / IDENTIFICAÇÃO</span>
                            <span class="detail-value host-name-value" style="${!device.name ? 'color: var(--text-dim); font-style: italic;' : ''}">${hostName}</span>
                        </div>
                        <div class="detail-item">
                            <span class="detail-label">MÉTODO DE TESTE</span>
                            <span class="detail-tag">
                                <svg width="11" height="11" viewBox="0 0 24 24" fill="currentColor" stroke="none"><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"/></svg>
                                ${hostMethod} (ICMP)
                            </span>
                        </div>
                        <div class="detail-item">
                            <span class="detail-label">LIMITE TIMEOUT</span>
                            <span class="detail-value mono timeout-value">${thresholdMs}</span>
                        </div>
                        <div class="detail-item full-width">
                            <span class="detail-label">ID DO SISTEMA / UUID</span>
                            <span class="detail-value mono uuid-value">${hostUUID}</span>
                        </div>
                    </div>
                </div>
            `;

            const mainRow = card.querySelector('.host-main-row');
            const expandBtn = card.querySelector('.host-expand-btn');

            expandBtn.addEventListener('click', (e) => {
                toggleExpandHost(e, device.id);
            });

            mainRow.addEventListener('click', () => {
                selectDevice(device);
            });

            mainRow.addEventListener('dblclick', (e) => {
                e.preventDefault();
                e.stopPropagation();
                setSelectedDevice(device);
                updateSelectedRowUI();
                openEditModal();
            });

            hostsList.appendChild(card);
        });
    }

    if (metricTotal) metricTotal.textContent = total;
    if (metricOnline) metricOnline.textContent = online;
    if (metricOffline) metricOffline.textContent = offline;

    if (state.selectedDevice) {
        const matching = state.currentIps.find(d => d.id === state.selectedDevice.id);
        setSelectedDevice(matching || null);
    }
    updateActionButtonsState();
}
