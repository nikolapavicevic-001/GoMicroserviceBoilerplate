// Custom JavaScript for web-app

// HTMX configuration
document.addEventListener('DOMContentLoaded', () => {
    console.log('Web app initialized');

    // Configure HTMX event listeners
    document.body.addEventListener('htmx:configRequest', (event) => {
        // Add custom headers if needed
        // event.detail.headers['X-Custom-Header'] = 'value';
    });

    document.body.addEventListener('htmx:beforeSwap', (event) => {
        // Handle errors
        if (event.detail.xhr.status === 401) {
            window.location.href = '/login';
        }
    });

    document.body.addEventListener('htmx:afterSwap', (event) => {
        // Reinitialize any plugins after content swap
        console.log('Content swapped');
    });

    // Global error handler
    document.body.addEventListener('htmx:responseError', (event) => {
        console.error('HTMX request failed:', event.detail);
        showToast('An error occurred. Please try again.', 'error');
    });
});

// Toast notification system
function showToast(message, type = 'info') {
    const toast = document.createElement('div');
    toast.className = `toast ${type}`;

    const colors = {
        info: 'bg-blue-500',
        success: 'bg-green-500',
        error: 'bg-red-500',
        warning: 'bg-yellow-500'
    };

    toast.innerHTML = `
        <div class="flex items-center space-x-3">
            <div class="flex-shrink-0">
                ${getToastIcon(type)}
            </div>
            <div class="flex-1">
                <p class="text-sm font-medium text-gray-900">${message}</p>
            </div>
            <button onclick="this.parentElement.parentElement.remove()" class="flex-shrink-0">
                <svg class="h-5 w-5 text-gray-400" fill="currentColor" viewBox="0 0 20 20">
                    <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd"/>
                </svg>
            </button>
        </div>
    `;

    document.body.appendChild(toast);

    // Auto remove after 5 seconds
    setTimeout(() => {
        toast.remove();
    }, 5000);
}

function getToastIcon(type) {
    const icons = {
        info: '<svg class="h-6 w-6 text-blue-500" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/></svg>',
        success: '<svg class="h-6 w-6 text-green-500" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/></svg>',
        error: '<svg class="h-6 w-6 text-red-500" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z"/></svg>',
        warning: '<svg class="h-6 w-6 text-yellow-500" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"/></svg>'
    };
    return icons[type] || icons.info;
}

// Utility: Debounce function for search inputs
function debounce(func, wait) {
    let timeout;
    return function executedFunction(...args) {
        const later = () => {
            clearTimeout(timeout);
            func(...args);
        };
        clearTimeout(timeout);
        timeout = setTimeout(later, wait);
    };
}

// Utility: Format date
function formatDate(dateString) {
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', {
        year: 'numeric',
        month: 'long',
        day: 'numeric'
    });
}

// Utility: Format time
function formatTime(dateString) {
    const date = new Date(dateString);
    return date.toLocaleTimeString('en-US', {
        hour: '2-digit',
        minute: '2-digit'
    });
}

// Export for use in other modules
window.app = {
    showToast,
    debounce,
    formatDate,
    formatTime
};
