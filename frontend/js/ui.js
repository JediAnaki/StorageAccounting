// UI component helpers

function showLoading(show = true) {
    const loading = document.getElementById('loading');
    if (loading) {
        loading.style.display = show ? 'block' : 'none';
    }
}

function showToast(message, type = 'info') {
    const toast = document.createElement('div');
    toast.className = `toast toast-${type}`;
    toast.textContent = message;
    document.body.appendChild(toast);

    setTimeout(() => {
        toast.classList.add('show');
    }, 100);

    setTimeout(() => {
        toast.classList.remove('show');
        setTimeout(() => toast.remove(), 300);
    }, 3000);
}

function showModal(title, content) {
    const modal = document.createElement('div');
    modal.className = 'modal';
    modal.innerHTML = `
        <div class="modal-content">
            <div class="modal-header">
                <h2>${title}</h2>
                <button class="modal-close" onclick="this.closest('.modal').remove()">&times;</button>
            </div>
            <div class="modal-body">
                ${content}
            </div>
        </div>
    `;
    document.body.appendChild(modal);

    setTimeout(() => modal.classList.add('show'), 100);
}

function createItemCard(item) {
    const card = document.createElement('div');
    card.className = 'item-card';
    if (item.current_quantity === 0) {
        card.classList.add('out-of-stock');
    }

    card.innerHTML = `
        <div class="item-photo">
            ${item.photo_url ?
                `<img src="${item.photo_url}" alt="${item.name}">` :
                '<div class="photo-placeholder">📦</div>'
            }
        </div>
        <div class="item-info">
            <h3>${item.name}</h3>
            <p class="item-quantity">Количество: <span>${item.current_quantity}</span></p>
        </div>
    `;

    card.addEventListener('click', () => {
        hapticFeedback('light');
        showItemDetails(item.id);
    });

    return card;
}

async function showItemDetails(itemId) {
    showLoading(true);
    try {
        const item = await api.getItem(itemId);
        const transactions = await api.getTransactions(itemId);

        const transactionsHtml = transactions.map(t => `
            <div class="transaction ${t.delta > 0 ? 'addition' : 'reduction'}">
                <div class="transaction-header">
                    <span class="transaction-delta">${t.delta > 0 ? '+' : ''}${t.delta}</span>
                    <span class="transaction-date">${new Date(t.timestamp).toLocaleString('ru-RU')}</span>
                </div>
                ${t.note ? `<div class="transaction-note">${t.note}</div>` : ''}
                <div class="transaction-user">Пользователь: ${t.user_id}</div>
            </div>
        `).join('');

        showModal(item.name, `
            <div class="item-details">
                ${item.photo_url ? `<img src="${item.photo_url}" alt="${item.name}" class="item-photo-large">` : ''}
                <p><strong>Текущее количество:</strong> ${item.current_quantity}</p>
                <h3>История изменений</h3>
                <div class="transactions-list">
                    ${transactionsHtml || '<p>Нет истории</p>'}
                </div>
            </div>
        `);
    } catch (error) {
        showToast('Ошибка загрузки данных: ' + error.message, 'error');
    } finally {
        showLoading(false);
    }
}
