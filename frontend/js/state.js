// ==============================================================================
// IP Monitor - Estado Global Compartilhado da Aplicação
// ==============================================================================

export const state = {
    selectedDevice: null,
    currentIps: [],
    modalMode: 'add', // 'add' | 'edit'
    expandedHostIds: new Set(),
};

// Funções utilitárias para mutação de estado controlada
export function setSelectedDevice(device) {
    state.selectedDevice = device;
}

export function setCurrentIps(ips) {
    state.currentIps = ips;
}

export function setModalMode(mode) {
    state.modalMode = mode;
}

export function toggleExpandedHostId(id) {
    if (state.expandedHostIds.has(id)) {
        state.expandedHostIds.delete(id);
        return false;
    } else {
        state.expandedHostIds.add(id);
        return true;
    }
}
