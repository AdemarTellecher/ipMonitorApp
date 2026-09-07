// Lógica reativa Vanilla JS para o frontend embutido do Wails v3

let selectedIp = null;
let currentIps = [];

// Elementos da UI
const metricTotal = document.getElementById('metricTotal');
const metricOnline = document.getElementById('metricOnline');
const metricOffline = document.getElementById('metricOffline');
const hostsList = document.getElementById('hostsList');
const ipInput = document.getElementById('ipInput');
const addBtn = document.getElementById('addBtn');
const refreshBtn = document.getElementById('refreshBtn');
const importBtn = document.getElementById('importBtn');
const fileInput = document.getElementById('fileInput');
const removeBtn = document.getElementById('removeBtn');
const themeToggleBtn = document.getElementById('themeToggleBtn');
const toast = document.getElementById('toast');

// Mostra feedback toast sutil
function showToast(msg, duration = 3000) {
    toast.textContent = msg;
    toast.classList.remove('hidden');
    setTimeout(() => {
        toast.classList.add('hidden');
    }, duration);
}

// Renderiza a lista de hosts na tabela
function renderHosts(ips) {
    currentIps = ips || [];
    let onlineCount = 0;
    let offlineCount = 0;

    hostsList.innerHTML = '';

    if (currentIps.length === 0) {
        hostsList.innerHTML = `<div style="text-align:center; padding: 25px 0; color: var(--text-dim); font-size: 0.8rem;">Nenhum IP cadastrado.<br>Adicione acima ou importe um arquivo JSON.</div>`;
    } else {
        currentIps.forEach(device => {
            const isOnline = device.status === 'Online';
            const isOffline = device.status === 'Offline';

            if (isOnline) onlineCount++;
            if (isOffline) offlineCount++;

            const row = document.createElement('div');
            row.className = `table-row ${selectedIp === device.ip ? 'selected' : ''}`;
            row.onclick = () => selectIp(device.ip);

            const badgeClass = isOnline ? 'badge-online' : (isOffline ? 'badge-offline' : 'badge-unknown');
            const dotClass = isOnline ? 'dot-online' : (isOffline ? 'dot-offline' : '');

            row.innerHTML = `
                <div class="col-ip font-mono">${device.ip}</div>
                <div class="col-status">
                    <span class="badge ${badgeClass}">
                        ${dotClass ? `<span class="dot ${dotClass}"></span>` : ''}
                        ${device.status}
                    </span>
                </div>
            `;
            hostsList.appendChild(row);
        });
    }

    // Atualiza contadores dos Cards
    metricTotal.textContent = currentIps.length;
    metricOnline.textContent = onlineCount;
    metricOffline.textContent = offlineCount;

    // Atualiza estado do botão Remover
    const isSelectedStillValid = currentIps.some(d => d.ip === selectedIp);
    if (!isSelectedStillValid) {
        selectedIp = null;
    }
    removeBtn.disabled = !selectedIp;
}

// Seleciona um IP para remoção
function selectIp(ip) {
    if (selectedIp === ip) {
        selectedIp = null;
    } else {
        selectedIp = ip;
    }
    renderHosts(currentIps);
}

// Chamadas Go via rota HTTP interna /api/... servida nativamente pelo Wails v3
async function callGo(endpoint, data = null) {
    try {
        const options = {
            method: data ? 'POST' : 'GET',
            headers: { 'Content-Type': 'application/json' },
        };
        if (data) {
            if (typeof data === 'string') {
                options.body = data;
            } else {
                options.body = JSON.stringify(data);
            }
        }
        const res = await fetch(`/api/${endpoint}`, options);
        if (!res.ok) {
            throw new Error(`HTTP ${res.status}`);
        }
        return await res.json();
    } catch (err) {
        console.error(`Erro na chamada /api/${endpoint}:`, err);
        throw err;
    }
}

// Carrega IPs
async function loadIPs() {
    try {
        const ips = await callGo('list');
        if (ips) {
            renderHosts(ips);
        }
    } catch (err) {
        console.error("Erro ao listar IPs:", err);
    }
}

// Adiciona IP
async function handleAddIP() {
    const ip = ipInput.value.trim();
    if (!ip) return;

    try {
        const res = await callGo('add', { ip });
        if (res && res.success) {
            ipInput.value = '';
            showToast(`IP ${ip} adicionado com sucesso!`);
            await loadIPs();
        } else if (res && res.error) {
            showToast(res.error);
        }
    } catch (err) {
        showToast(err.toString());
    }
}

// Remove IP selecionado
async function handleRemoveIP() {
    if (!selectedIp) return;
    const toRemove = selectedIp;

    try {
        const res = await callGo('remove', { ip: toRemove });
        selectedIp = null;
        showToast(`IP ${toRemove} removido.`);
        await loadIPs();
    } catch (err) {
        showToast(err.toString());
    }
}

// Atualiza status de todos via Ping
async function handleRefreshAll() {
    refreshBtn.disabled = true;
    showToast("Atualizando status dos IPs via ICMP Ping...");
    try {
        const updated = await callGo('update');
        if (updated) {
            renderHosts(updated);
        } else {
            await loadIPs();
        }
        showToast("Status atualizado com sucesso!");
    } catch (err) {
        showToast("Erro ao atualizar status: " + err);
    } finally {
        refreshBtn.disabled = false;
    }
}

// Importar JSON
importBtn.onclick = () => fileInput.click();

fileInput.onchange = async (e) => {
    const file = e.target.files[0];
    if (!file) return;

    const reader = new FileReader();
    reader.onload = async (evt) => {
        try {
            const content = evt.target.result;
            const res = await callGo('import', content);
            if (res && typeof res.count === 'number') {
                showToast(`${res.count} novo(s) IP(s) importado(s)!`);
            } else if (res && res.error) {
                showToast(res.error);
            } else {
                showToast("Arquivo JSON processado.");
            }
            await loadIPs();
        } catch (err) {
            showToast("Erro ao importar JSON: " + err);
        }
    };
    reader.readAsText(file);
    fileInput.value = '';
};

// Eventos de botões e input
addBtn.onclick = handleAddIP;
ipInput.onkeydown = (e) => {
    if (e.key === 'Enter') handleAddIP();
};
refreshBtn.onclick = handleRefreshAll;
removeBtn.onclick = handleRemoveIP;

// Alternar Tema
let isDark = true;
themeToggleBtn.onclick = () => {
    isDark = !isDark;
    document.body.classList.toggle('light-theme', !isDark);
};

// Ouvinte de evento periódico de ping vindo do Go (Wails events)
if (window.wails && window.wails.Events) {
    window.wails.Events.On('ips-updated', (data) => {
        renderHosts(data);
    });
}

// Inicialização
document.addEventListener('DOMContentLoaded', () => {
    loadIPs();
});
