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

function showQuantityChangeForm(itemId, action, currentQuantity) {
    hapticFeedback('light');

    const isAdd = action === 'add';
    const title = isAdd ? 'Добавить товар' : 'Списать товар';
    const btnClass = isAdd ? 'btn-success' : 'btn-danger';
    const btnText = isAdd ? 'Добавить' : 'Списать';

    const formHtml = `
        <form id="quantity-change-form" class="quantity-change-form">
            <div class="form-group">
                <label for="quantity">Количество:</label>
                <input type="number" id="quantity" name="quantity" min="1" required>
            </div>
            <div class="form-group">
                <label for="note">Примечание (необязательно):</label>
                <textarea id="note" name="note" rows="3" placeholder="Добавьте комментарий..."></textarea>
            </div>
            ${!isAdd && currentQuantity > 0 ? `
                <div id="negative-warning" class="warning-message" style="display: none;">
                    <p>⚠️ Внимание! Количество станет отрицательным (${currentQuantity} → <span id="projected-quantity"></span>)</p>
                    <p>Разрешить отрицательный остаток?</p>
                </div>
            ` : ''}
            <div class="form-actions">
                <button type="button" class="btn btn-secondary" onclick="this.closest('.modal').remove()">Отмена</button>
                <button type="submit" class="btn ${btnClass}">${btnText}</button>
            </div>
        </form>
    `;

    showModal(title, formHtml);

    // Setup form handling
    const form = document.getElementById('quantity-change-form');
    const quantityInput = document.getElementById('quantity');
    const noteInput = document.getElementById('note');

    // Show negative quantity warning for reductions
    if (!isAdd) {
        quantityInput.addEventListener('input', () => {
            const quantity = parseInt(quantityInput.value) || 0;
            const projectedQuantity = currentQuantity - quantity;

            const warning = document.getElementById('negative-warning');
            const projectedSpan = document.getElementById('projected-quantity');

            if (projectedQuantity < 0 && warning && projectedSpan) {
                projectedSpan.textContent = projectedQuantity;
                warning.style.display = 'block';
            } else if (warning) {
                warning.style.display = 'none';
            }
        });
    }

    form.addEventListener('submit', async (e) => {
        e.preventDefault();

        const quantity = parseInt(quantityInput.value);
        const note = noteInput.value.trim() || null;

        if (!quantity || quantity <= 0) {
            showToast('Введите корректное количество', 'error');
            return;
        }

        await handleQuantityChange(itemId, action, quantity, note);
        form.closest('.modal').remove();
    });
}

async function handleQuantityChange(itemId, action, quantity, note) {
    showLoading(true);

    try {
        const response = action === 'add'
            ? await api.addStock(itemId, quantity, note)
            : await api.removeStock(itemId, quantity, note);

        showToast(response.message || 'Количество обновлено', 'success');
        hapticFeedback('success');

        // Refresh the item details
        await showItemDetails(itemId);

    } catch (error) {
        // Handle concurrent update errors
        if (error.message.includes('modified by another user') || error.message.includes('409')) {
            showToast('Товар был изменен другим пользователем. Пожалуйста, обновите страницу.', 'error');
            hapticFeedback('error');
        } else {
            showToast('Ошибка обновления: ' + error.message, 'error');
            hapticFeedback('error');
        }
    } finally {
        showLoading(false);
    }
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
                <div class="quantity-actions">
                    <button class="btn btn-success" onclick="showQuantityChangeForm('${itemId}', 'add', ${item.current_quantity})">
                        <span>+ Добавить</span>
                    </button>
                    <button class="btn btn-danger" onclick="showQuantityChangeForm('${itemId}', 'remove', ${item.current_quantity})">
                        <span>− Списать</span>
                    </button>
                </div>
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
