// Telegram WebApp SDK integration
const tg = window.Telegram.WebApp;

// Initialize Telegram WebApp
function initTelegram() {
    // Expand the WebApp to full height
    tg.expand();

    // Enable closing confirmation
    tg.enableClosingConfirmation();

    // Set header color based on theme
    const isDark = tg.colorScheme === 'dark';
    document.body.classList.toggle('dark-theme', isDark);

    // Get user information
    const user = tg.initDataUnsafe?.user;

    return {
        user,
        initData: tg.initData,
        theme: tg.colorScheme,
        platform: tg.platform
    };
}

// Get authentication data for API requests
function getAuthData() {
    return {
        initData: tg.initData
    };
}

// Show/hide main button
function showMainButton(text, onClick) {
    tg.MainButton.setText(text);
    tg.MainButton.show();
    tg.MainButton.onClick(onClick);
}

function hideMainButton() {
    tg.MainButton.hide();
}

// Haptic feedback
function hapticFeedback(type = 'light') {
    if (tg.HapticFeedback) {
        tg.HapticFeedback.impactOccurred(type);
    }
}
