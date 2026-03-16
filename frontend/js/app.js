// Main application logic

let currentOffset = 0;
const ITEMS_PER_PAGE = 50;

async function loadItems() {
    showLoading(true);
    const inventoryList = document.getElementById('inventory-list');

    try {
        const items = await api.getItems(ITEMS_PER_PAGE, currentOffset);

        if (items.length === 0 && currentOffset === 0) {
            inventoryList.innerHTML = '<p class="empty-state">Нет товаров на складе</p>';
        } else {
            items.forEach(item => {
                const card = createItemCard(item);
                inventoryList.appendChild(card);
            });

            currentOffset += items.length;
        }

        inventoryList.style.display = 'block';
    } catch (error) {
        showToast('Ошибка загрузки данных: ' + error.message, 'error');
    } finally {
        showLoading(false);
    }
}

// Initialize app
document.addEventListener('DOMContentLoaded', () => {
    const telegramData = initTelegram();

    if (!telegramData.initData) {
        console.warn('Not running in Telegram WebApp environment');
    }

    loadItems();
});

// Infinite scroll
window.addEventListener('scroll', () => {
    if ((window.innerHeight + window.scrollY) >= document.body.offsetHeight - 100) {
        loadItems();
    }
});
