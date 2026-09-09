// ==============================================================================
// IP Monitor - Lógica do Frontend (Estilo PC Manager)
// ==============================================================================

let selectedDevice = null;
let currentIps = [];
let modalMode = 'add'; // 'add' | 'edit'

// Elementos da UI
const metricTotal = document.getElementById('metricTotal');
const metricOnline = document.getElementById('metricOnline');
const metricOffline = document.getElementById('metricOffline');
const hostsList = document.getElementById('hostsList');
const ipInput = document.getElementById('ipInput');
const addBtn = document.getElementById('addBtn');
const addIpModal = document.getElementById('addIpModal');
const modalTitle = document.getElementById('modalTitle');
const modalDesc = document.getElementById('modalDesc');
const modalConfirmLabel = document.getElementById('modalConfirmLabel');
const networkOverviewCard = document.getElementById('networkOverviewCard');
const modalBackdrop = document.getElementById('modalBackdrop');
const cancelAddBtn = document.getElementById('cancelAddBtn');
const modalCloseXBtn = document.getElementById('modalCloseXBtn');
const fileInput = document.getElementById('fileInput');
const themeToggleBtn = document.getElementById('themeToggleBtn');
const toast = document.getElementById('toast');

// Elementos da Sidebar
const navFocusAdd = document.getElementById('navFocusAdd');
const navImport = document.getElementById('navImport');
const navRefresh = document.getElementById('navRefresh');
const navEdit = document.getElementById('navEdit');
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

// Renderiza a lista de hosts na tabela Fluent (Priorizando dispositivos Offline no topo)
function renderHosts(ips) {
    currentIps = (ips || []).slice().sort((a, b) => {
        // Prioridade: Offline (0) > Desconhecido/outros (1) > Online (2)
        const getPriority = (status) => {
            if (status === 'Offline') return 0;
            if (status === 'Online') return 2;
            return 1;
        };

        const diff = getPriority(a.status) - getPriority(b.status);
        if (diff !== 0) return diff;

        // Desempate alfanumérico pelo IP/Host
        return (a.ip || '').localeCompare(b.ip || '', undefined, { numeric: true, sensitivity: 'base' });
    });

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

            const isSelected = selectedDevice && selectedDevice.id === device.id;
            const row = document.createElement('div');
            row.className = `host-row ${isSelected ? 'selected' : ''}`;
            row.onclick = () => selectDevice(device);
            row.ondblclick = () => {
                selectDevice(device);
                openEditModal();
            };

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
    if (selectedDevice) {
        const matching = currentIps.find(d => d.id === selectedDevice.id);
        selectedDevice = matching || null;
    }
    updateActionButtonsState();
}

// Atualiza o estado dos botões de ação sensíveis à seleção (Editar e Remover)
function updateActionButtonsState() {
    const hasSelection = Boolean(selectedDevice);
    if (hasSelection) {
        navEdit.classList.remove('disabled');
        navRemove.classList.remove('disabled');
    } else {
        navEdit.classList.add('disabled');
        navRemove.classList.add('disabled');
    }
}

// Seleciona um dispositivo para ação
function selectDevice(device) {
    if (selectedDevice && selectedDevice.id === device.id) {
        selectedDevice = null;
    } else {
        selectedDevice = device;
    }
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

// Controle do Modal de Cadastro Sobreposto (Bloqueante)
function openAddModal() {
    modalMode = 'add';
    modalTitle.textContent = 'Cadastrar Novo Dispositivo';
    modalDesc.textContent = 'Digite o IP ou hostname para iniciar o monitoramento contínuo via ICMP Ping.';
    modalConfirmLabel.textContent = 'Cadastrar';
    ipInput.value = '';

    modalBackdrop.classList.remove('hidden');
    networkOverviewCard.classList.add('has-modal');
    addIpModal.classList.remove('hidden');
    navFocusAdd.classList.add('active');
    setTimeout(() => {
        ipInput.focus();
        ipInput.select();
    }, 50);
}

// Controle do Modal de Edição
function openEditModal() {
    if (!selectedDevice) {
        showToast('Selecione um host na lista para editar');
        return;
    }
    modalMode = 'edit';
    modalTitle.textContent = 'Editar Dispositivo / Host';
    modalDesc.textContent = 'Atualize o IP ou hostname do dispositivo selecionado.';
    modalConfirmLabel.textContent = 'Salvar';
    ipInput.value = selectedDevice.ip;

    modalBackdrop.classList.remove('hidden');
    networkOverviewCard.classList.add('has-modal');
    addIpModal.classList.remove('hidden');
    navEdit.classList.add('active');
    setTimeout(() => {
        ipInput.focus();
        ipInput.select();
    }, 50);
}

function closeAddModal() {
    addIpModal.classList.add('hidden');
    networkOverviewCard.classList.remove('has-modal');
    modalBackdrop.classList.add('hidden');
    navFocusAdd.classList.remove('active');
    navEdit.classList.remove('active');
    ipInput.value = '';
    modalMode = 'add';
}

// Submissão do Modal (Cadastrar ou Editar dependendo do modo)
async function handleModalSubmit() {
    const rawValue = ipInput.value.trim();
    if (!rawValue) {
        ipInput.focus();
        return;
    }

    try {
        addBtn.disabled = true;
        if (modalMode === 'add') {
            const result = await callGo('add', { ip: rawValue });
            if (result && result.error) {
                showToast(result.error);
            } else {
                closeAddModal();
                showToast(`Host ${rawValue} cadastrado com sucesso!`);
                await loadIps();
            }
        } else if (modalMode === 'edit') {
            if (!selectedDevice) return;
            const result = await callGo('edit', { id: selectedDevice.id, newIp: rawValue });
            if (result && result.error) {
                showToast(result.error);
            } else {
                closeAddModal();
                showToast(`Host alterado para ${rawValue} com sucesso!`);
                selectedDevice = null;
                await loadIps();
            }
        }
    } catch (err) {
        showToast(modalMode === 'edit' ? 'Erro ao editar host' : 'Erro ao cadastrar host');
    } finally {
        addBtn.disabled = false;
    }
}

// Atualiza o status de todos os IPs com animação na sidebar
async function handleRefresh() {
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

// Remove o IP selecionado
async function handleRemove() {
    if (!selectedDevice) return;
    const ipToRemove = selectedDevice.ip;

    try {
        navRemove.classList.add('disabled');
        navEdit.classList.add('disabled');
        const result = await callGo('remove', { ip: ipToRemove });
        if (result && result.error) {
            showToast(result.error);
        } else {
            showToast(`Host ${ipToRemove} removido`);
            selectedDevice = null;
            await loadIps();
        }
    } catch (err) {
        showToast('Erro ao remover host');
    } finally {
        updateActionButtonsState();
    }
}

// Tratador do input file para importação
async function handleFileSelected(event) {
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
        fileInput.value = '';
    }
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
addBtn.addEventListener('click', handleModalSubmit);
cancelAddBtn.addEventListener('click', closeAddModal);
modalCloseXBtn.addEventListener('click', closeAddModal);
modalBackdrop.addEventListener('click', closeAddModal);

ipInput.addEventListener('keydown', (e) => {
    if (e.key === 'Enter') handleModalSubmit();
    if (e.key === 'Escape') closeAddModal();
});

fileInput.addEventListener('change', handleFileSelected);
themeToggleBtn.addEventListener('click', toggleTheme);

// Conexões da Sidebar
navFocusAdd.addEventListener('click', () => {
    if (addIpModal.classList.contains('hidden')) {
        openAddModal();
    } else {
        closeAddModal();
    }
});
navImport.addEventListener('click', () => fileInput.click());
navRefresh.addEventListener('click', handleRefresh);
navEdit.addEventListener('click', () => {
    if (selectedDevice) {
        openEditModal();
    }
});
navRemove.addEventListener('click', handleRemove);
navAbout.addEventListener('click', () => {
    showToast('IP Monitor v2.2 • Wails v3 + SQLite3');
});

// Listener para eventos periódicos emitidos pelo backend Go (Wails v3)
if (window.wails && window.wails.Events) {
    window.wails.Events.On('ips-updated', (updatedIps) => {
        renderHosts(updatedIps);
    });
}

// Polling ativo no frontend para sincronização contínua de status em tempo real
// Garante atualização na tela caso o backend termine uma varredura automática a cada 1 minuto
setInterval(() => {
    // Não recarrega a tabela se o usuário estiver com um modal de adição/edição aberto
    if (!addIpModal.classList.contains('hidden')) {
        return;
    }
    loadIps();
}, 5000);

// Inicia aplicação
initTheme();
loadIps();
