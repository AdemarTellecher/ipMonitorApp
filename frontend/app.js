// ==============================================================================
// IP Monitor - Lógica do Frontend (Estilo PC Manager)
// ==============================================================================

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
const refreshBtnText = document.getElementById('refreshBtnText');
const importBtn = document.getElementById('importBtn');
const fileInput = document.getElementById('fileInput');
const removeBtn = document.getElementById('removeBtn');
const themeToggleBtn = document.getElementById('themeToggleBtn');
const toast = document.getElementById('toast');

// Elementos da Sidebar
const navHome = document.getElementById('navHome');
const navFocusAdd = document.getElementById('navFocusAdd');
const navImport = document.getElementById('navImport');
const navRefresh = document.getElementById('navRefresh');
const navRemove = document.getElementById('navRemove');
const navAbout = document.getElementById('navAbout');

// Exibe feedback flutuante (Toast)
function showToast(msg, duration = 2800) {
    toast.textContent = msg;
    toast.classList.remove('hidden');
    setTimeout(() => {
        toast.classList.add('hidden');
    }, duration);
}

// Renderiza a lista de hosts na tabela Fluent
function renderHosts(ips) {
    currentIps = ips || [];
    let onlineCount = 0;
    let offlineCount = 0;

    hostsList.innerHTML = '';

    if (currentIps.length === 0) {
        hostsList.innerHTML = `<div style="text-align:center; padding: 30px 0; color: var(--text-dim); font-size: 0.8rem;">Nenhum dispositivo na lista.<br>Cadastre um IP acima ou importe um JSON.</div>`;
    } else {
        currentIps.forEach(device => {
            const isOnline = device.status === 'Online';
            const isOffline = device.status === 'Offline';

            if (isOnline) onlineCount++;
            if (isOffline) offlineCount++;

            const row = document.createElement('div');
            row.className = `host-row ${selectedIp === device.ip ? 'selected' : ''}`;
            row.onclick = () => selectIp(device.ip);

            const badgeClass = isOnline ? 'online' : (isOffline ? 'offline' : 'unknown');

            row.innerHTML = `
                <div class="host-ip-col">
                    <span class="host-radio-dot"></span>
                    <span>${device.ip}</span>
                </div>
                <div>
                    <span class="status-badge ${badgeClass}">
                        <span class="status-pip"></span>
                        ${device.status}
                    </span>
                </div>
            `;
            hostsList.appendChild(row);
        });
    }

    // Atualiza contadores
    metricTotal.textContent = currentIps.length;
    metricOnline.textContent = onlineCount;
    metricOffline.textContent = offlineCount;

    // Valida seleção ativa
    const isSelectedStillValid = currentIps.some(d => d.ip === selectedIp);
    if (!isSelectedStillValid) {
        selectedIp = null;
    }
    updateRemoveButtonsState();
}

// Atualiza o estado dos botões de remover
function updateRemoveButtonsState() {
    const hasSelection = Boolean(selectedIp);
    removeBtn.disabled = !hasSelection;
    if (hasSelection) {
        navRemove.classList.remove('disabled');
    } else {
        navRemove.classList.add('disabled');
    }
}

// Seleciona um IP para ação
function selectIp(ip) {
    selectedIp = (selectedIp === ip) ? null : ip;
    renderHosts(currentIps);
}

// Chamadas Go via rota HTTP interna /api/...
async function callGo(endpoint, data = null) {
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

// Carrega os IPs ao inicializar
async function loadIps() {
    try {
        const ips = await callGo('list');
        renderHosts(ips);
    } catch (err) {
        showToast('Erro ao carregar lista de IPs');
    }
}

// Adiciona um novo IP
async function handleAddIp() {
    const ip = ipInput.value.trim();
    if (!ip) {
        ipInput.focus();
        return;
    }

    try {
        addBtn.disabled = true;
        const result = await callGo('add', { ip });
        if (result && result.error) {
            showToast(result.error);
        } else {
            ipInput.value = '';
            showToast(`Host ${ip} adicionado com sucesso!`);
            await loadIps();
        }
    } catch (err) {
        showToast('Erro ao adicionar host');
    } finally {
        addBtn.disabled = false;
        ipInput.focus();
    }
}

// Atualiza o status de todos os IPs com animação
async function handleRefresh() {
    try {
        refreshBtn.classList.add('spinning');
        refreshBtn.disabled = true;
        refreshBtnText.textContent = 'Verificando...';

        const updated = await callGo('update');
        renderHosts(updated);
        showToast('Varredura ICMP Ping concluída!');
    } catch (err) {
        showToast('Falha ao atualizar status dos hosts');
    } finally {
        refreshBtn.classList.remove('spinning');
        refreshBtn.disabled = false;
        refreshBtnText.textContent = 'Atualizar Todos';
    }
}

// Remove o IP selecionado
async function handleRemove() {
    if (!selectedIp) return;
    const ipToRemove = selectedIp;

    try {
        removeBtn.disabled = true;
        const result = await callGo('remove', { ip: ipToRemove });
        if (result && result.error) {
            showToast(result.error);
        } else {
            showToast(`Host ${ipToRemove} removido`);
            selectedIp = null;
            await loadIps();
        }
    } catch (err) {
        showToast('Erro ao remover host');
    } finally {
        updateRemoveButtonsState();
    }
}

// Importa arquivo JSON
async function handleFileSelected(event) {
    const file = event.target.files[0];
    if (!file) return;

    const reader = new FileReader();
    reader.onload = async (e) => {
        try {
            const rawContent = e.target.result;
            // Envia o JSON bruto diretamente ao backend Go (que suporta {sites:[...]} ou listas)
            const result = await callGo('import', rawContent);
            if (result && result.error) {
                showToast(`Erro na importação: ${result.error}`);
            } else if (result && typeof result.count === 'number') {
                showToast(`${result.count} novo(s) host(s) importado(s) com sucesso!`);
                await loadIps();
            } else {
                showToast("Arquivo JSON importado com sucesso!");
                await loadIps();
            }
        } catch (err) {
            console.error('Erro ao processar arquivo JSON:', err);
            showToast('Falha ao importar o arquivo JSON');
        } finally {
            fileInput.value = '';
        }
    };
    reader.readAsText(file);
}

// Alterna tema claro/escuro
function toggleTheme() {
    const isLight = document.body.classList.toggle('light-theme');
    localStorage.setItem('ipmonitor_theme', isLight ? 'light' : 'dark');
}

// Inicializa preferências de tema
function initTheme() {
    const saved = localStorage.getItem('ipmonitor_theme');
    if (saved === 'light') {
        document.body.classList.add('light-theme');
    }
}

// Conexões de Eventos da Interface
addBtn.addEventListener('click', handleAddIp);
ipInput.addEventListener('keydown', (e) => {
    if (e.key === 'Enter') handleAddIp();
});

refreshBtn.addEventListener('click', handleRefresh);
removeBtn.addEventListener('click', handleRemove);
importBtn.addEventListener('click', () => fileInput.click());
fileInput.addEventListener('change', handleFileSelected);
themeToggleBtn.addEventListener('click', toggleTheme);

// Conexões da Sidebar
navHome.addEventListener('click', () => {
    window.scrollTo({ top: 0, behavior: 'smooth' });
});
navFocusAdd.addEventListener('click', () => {
    ipInput.focus();
    ipInput.select();
});
navImport.addEventListener('click', () => fileInput.click());
navRefresh.addEventListener('click', handleRefresh);
navRemove.addEventListener('click', handleRemove);
navAbout.addEventListener('click', () => {
    showToast('IP Monitor v2.1 • Wails v3 + SQLite3');
});

// Listener para eventos periódicos emitidos pelo backend Go (Wails v3)
if (window.wails && window.wails.Events) {
    window.wails.Events.On('ips-updated', (updatedIps) => {
        renderHosts(updatedIps);
    });
}

// Inicia aplicação
initTheme();
loadIps();
