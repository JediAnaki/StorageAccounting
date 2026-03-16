// API client for backend communication
const API_BASE_URL = window.location.origin + '/api';

async function apiRequest(endpoint, options = {}) {
    const authData = getAuthData();

    const defaultOptions = {
        headers: {
            'Content-Type': 'application/json',
            'X-Telegram-Init-Data': authData.initData
        }
    };

    const mergedOptions = {
        ...defaultOptions,
        ...options,
        headers: {
            ...defaultOptions.headers,
            ...options.headers
        }
    };

    try {
        const response = await fetch(API_BASE_URL + endpoint, mergedOptions);

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.message || 'Request failed');
        }

        return await response.json();
    } catch (error) {
        console.error('API request failed:', error);
        throw error;
    }
}

// API methods
const api = {
    // Items
    getItems: (limit = 50, offset = 0) =>
        apiRequest(`/items?limit=${limit}&offset=${offset}`),

    getItem: (id) =>
        apiRequest(`/items/${id}`),

    createItem: (data) =>
        apiRequest('/items', {
            method: 'POST',
            body: JSON.stringify(data)
        }),

    updateItemQuantity: (id, delta, note = '') =>
        apiRequest(`/items/${id}/quantity`, {
            method: 'PUT',
            body: JSON.stringify({ delta, note })
        }),

    // Transactions
    getTransactions: (itemId, limit = 50, offset = 0) =>
        apiRequest(`/items/${itemId}/transactions?limit=${limit}&offset=${offset}`)
};
