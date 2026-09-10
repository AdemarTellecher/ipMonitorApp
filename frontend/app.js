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
const nameInput = document.getElementById('nameInput');
const methodSelect = document.getElementById('methodSelect');
const thresholdInput = document.getElementById('thresholdInput');
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

// Conjunto que armazena os IDs dos hosts atualmente expandidos
const expandedHostIds = new Set();

// Alterna o estado expandido/recolhido de um host
function toggleExpandHost(event, deviceId) {
    if (event) {
        event.stopPropagation();
    }
    const card = document.getElementById(`host-card-${deviceId}`);
    if (expandedHostIds.has(deviceId)) {
        expandedHostIds.delete(deviceId);
        if (card) card.classList.remove('is-expanded');
    } else {
        expandedHostIds.add(deviceId);
        if (card) card.classList.add('is-expanded');
    }
}

// Renderiza a lista de hosts e o sumário de rede com dados prontos entregues pelo Go
function renderHosts(data) {
    let devices = [];
    let total = 0;
    let online = 0;
    let offline = 0;

    // Se receber a estrutura completa NetworkOverview calculada no Go
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

    // Mantém a ordem canônica retornada pelo SQLite/Go (Offline > Desconhecido > Online > Host)
    currentIps = devices;

    hostsList.innerHTML = '';

    if (currentIps.length === 0) {
        hostsList.innerHTML = `<div style="text-align:center; padding: 30px 0; color: var(--text-dim); font-size: 0.8rem;">Nenhum dispositivo na lista.<br>Cadastre um IP acima ou importe um JSON.</div>`;
    } else {
        currentIps.forEach(device => {
            const isOnline = device.status === 'Online';
            const isOffline = device.status === 'Offline';

            const isSelected = selectedDevice && selectedDevice.id === device.id;
            const isExpanded = expandedHostIds.has(device.id);

            const card = document.createElement('div');
            card.className = `host-item-card ${isSelected ? 'selected' : ''} ${isExpanded ? 'is-expanded' : ''}`;
            card.id = `host-card-${device.id}`;
            card.dataset.id = device.id;

            const badgeClass = isOnline ? 'online' : (isOffline ? 'offline' : 'unknown');

            // Formatação dos metadados ricos do host
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
                            <span class="detail-label">Nome / Identificação</span>
                            <span class="detail-value" style="${!device.name ? 'color: var(--text-dim); font-style: italic;' : ''}">${hostName}</span>
                        </div>
                        <div class="detail-item">
                            <span class="detail-label">Método de Teste</span>
                            <span class="detail-tag">
                                <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"/></svg>
                                ${hostMethod} (ICMP)
                            </span>
                        </div>
                        <div class="detail-item">
                            <span class="detail-label">Limite Timeout</span>
                            <span class="detail-value mono">${thresholdMs}</span>
                        </div>
                        <div class="detail-item full-width">
                            <span class="detail-label">ID do Sistema / UUID</span>
                            <span class="detail-value mono">${hostUUID}</span>
                        </div>
                    </div>
                </div>
            `;

            const mainRow = card.querySelector('.host-main-row');
            const expandBtn = card.querySelector('.host-expand-btn');

            // Clique na setinha alterna expansão
            expandBtn.addEventListener('click', (e) => {
                toggleExpandHost(e, device.id);
            });

            // Clique simples na linha seleciona/deseleciona
            mainRow.addEventListener('click', () => {
                selectDevice(device);
            });

            // Duplo clique rápido na linha abre diretamente a edição
            mainRow.addEventListener('dblclick', (e) => {
                e.preventDefault();
                e.stopPropagation();
                selectedDevice = device;
                updateSelectedRowUI();
                openEditModal();
            });

            hostsList.appendChild(card);
        });
    }

    // Atualiza contadores com métricas consolidadas pelo backend Go
    metricTotal.textContent = total;
    metricOnline.textContent = online;
    metricOffline.textContent = offline;

    // Valida seleção ativa
    if (selectedDevice) {
        const matching = currentIps.find(d => d.id === selectedDevice.id);
        selectedDevice = matching || null;
    }
    updateActionButtonsState();
}

// Atualiza visualmente a classe 'selected' nos cards existentes sem recriar o DOM
function updateSelectedRowUI() {
    const cards = hostsList.querySelectorAll('.host-item-card');
    cards.forEach(c => {
        const rowId = parseInt(c.dataset.id, 10);
        if (selectedDevice && rowId === selectedDevice.id) {
            c.classList.add('selected');
        } else {
            c.classList.remove('selected');
        }
    });
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

// Seleciona um dispositivo para ação mantendo o DOM estável para cliques duplos rápidos
function selectDevice(device) {
    if (selectedDevice && selectedDevice.id === device.id) {
        selectedDevice = null;
    } else {
        selectedDevice = device;
    }
    updateSelectedRowUI();
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
    modalDesc.textContent = 'Preencha os dados do dispositivo para o monitoramento contínuo.';
    modalConfirmLabel.textContent = 'Cadastrar';
    ipInput.value = '';
    nameInput.value = '';
    methodSelect.value = 'PING';
    thresholdInput.value = '2000';

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
    modalDesc.textContent = 'Atualize as informações do dispositivo selecionado.';
    modalConfirmLabel.textContent = 'Salvar';
    ipInput.value = selectedDevice.ip || '';
    nameInput.value = selectedDevice.name || '';
    methodSelect.value = selectedDevice.method || 'PING';
    thresholdInput.value = selectedDevice.thresholdMs ? selectedDevice.thresholdMs : '2000';

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
    nameInput.value = '';
    methodSelect.value = 'PING';
    thresholdInput.value = '2000';
    modalMode = 'add';
}

// Submissão do Modal (Cadastrar ou Editar dependendo do modo)
async function handleModalSubmit() {
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
        if (modalMode === 'add') {
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
        } else if (modalMode === 'edit') {
            if (!selectedDevice) return;
            const result = await callGo('edit', {
                id: selectedDevice.id,
                newIp: rawIp,
                name: nameVal,
                method: methodVal,
                thresholdMs: thresholdVal,
                uuid: selectedDevice.uuid || '',
            });
            if (result && result.error) {
                showToast(result.error);
            } else {
                closeAddModal();
                showToast(`Host alterado para ${rawIp} com sucesso!`);
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

[ipInput, nameInput, thresholdInput].forEach(field => {
    if (!field) return;
    field.addEventListener('keydown', (e) => {
        if (e.key === 'Enter') handleModalSubmit();
        if (e.key === 'Escape') closeAddModal();
    });
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

// Listener para eventos de atualização de status emitidos pelo backend Go (Wails v3)
function handleIpsUpdatedEvent(eventOrData) {
    console.log('[IP Monitor] Evento ips-updated recebido do backend:', eventOrData);
    const data = (eventOrData && eventOrData.data !== undefined) ? eventOrData.data : eventOrData;
    if (data) {
        renderHosts(data);
    }
}

// Registra listener no Wails Events com retry caso o runtime carregue de forma assíncrona
function setupWailsEventListeners() {
    if (window.wails && window.wails.Events && typeof window.wails.Events.On === 'function') {
        window.wails.Events.On('ips-updated', handleIpsUpdatedEvent);
        console.log('[IP Monitor] Listener Wails v3 para "ips-updated" registrado com sucesso.');
    } else {
        // Tenta novamente caso o script do runtime demore alguns milissegundos
        setTimeout(() => {
            if (window.wails && window.wails.Events && typeof window.wails.Events.On === 'function') {
                window.wails.Events.On('ips-updated', handleIpsUpdatedEvent);
                console.log('[IP Monitor] Listener Wails v3 para "ips-updated" registrado após retry.');
            } else {
                console.warn('[IP Monitor] Wails runtime não detectado para eventos nativos; operando com fallback.');
            }
        }, 500);
    }
}

// Fallback de polling a cada 60 segundos (garante que a UI sempre atualize mesmo se eventos IPC falharem)
function startPeriodicFallbackSync() {
    setInterval(async () => {
        try {
            const overview = await callGo('list');
            if (overview) {
                renderHosts(overview);
            }
        } catch (e) {
            console.warn('[IP Monitor] Fallback periódico de sincronização falhou:', e);
        }
    }, 60000);
}

setupWailsEventListeners();
startPeriodicFallbackSync();

// Carrega a versão dinâmica exposta pelo backend Go
async function loadVersion() {
    const badge = document.getElementById('appVersionBadge');
    if (!badge) return;
    try {
        const res = await fetch('/api/version');
        if (res.ok) {
            const data = await res.json();
            if (data && data.version) {
                badge.textContent = data.version;
            }
        }
    } catch (e) {
        console.warn('Não foi possível carregar a versão dinamicamente:', e);
    }
}

// Inicia aplicação carregando os dados iniciais
initTheme();
loadIps();
loadVersion();
