var token = localStorage.getItem('token');
var userId = localStorage.getItem('userId');
var cartCount = 0;
var productsCache = {};

// Auth
async function register() {
    var email = document.getElementById('authEmail').value.trim();
    var password = document.getElementById('authPassword').value.trim();

    if (!email || !password) {
        ui.showAlert('authAlert', 'Please enter email and password', 'error');
        return;
    }

    try {
        var data = await api.register(email, password);
        token = data.token;
        userId = data.user.id;
        localStorage.setItem('token', token);
        localStorage.setItem('userId', userId);
        ui.showToast('Account created successfully!', 'success');
        showAppScreen();
        loadProducts();
    } catch (err) {
        ui.showAlert('authAlert', err.message, 'error');
    }
}

async function login() {
    var email = document.getElementById('authEmail').value.trim();
    var password = document.getElementById('authPassword').value.trim();

    if (!email || !password) {
        ui.showAlert('authAlert', 'Please enter email and password', 'error');
        return;
    }

    try {
        var data = await api.loginUser(email, password);
        token = data.token;
        userId = data.user.id;
        localStorage.setItem('token', token);
        localStorage.setItem('userId', userId);
        ui.showToast('Welcome back!', 'success');
        showAppScreen();
        loadProducts();
    } catch (err) {
        ui.showAlert('authAlert', err.message, 'error');
    }
}

async function demoLogin() {
    var demoEmail = 'demo-' + Date.now() + '@firdousshops.com';
    var demoPassword = 'demo123456';
    try {
        var data = await api.register(demoEmail, demoPassword);
        token = data.token;
        userId = data.user.id;
        localStorage.setItem('token', token);
        localStorage.setItem('userId', userId);
        ui.showToast('Welcome to Demo Mode! 🎉', 'success');
        showAppScreen();
        loadProducts();
    } catch (err) {
        ui.showToast('Demo login failed: ' + err.message, 'error');
    }
}

function logout() {
    localStorage.removeItem('token');
    localStorage.removeItem('userId');
    token = null;
    userId = null;
    document.getElementById('authEmail').value = '';
    document.getElementById('authPassword').value = '';
    showAuthScreen();
    ui.showToast('Signed out successfully', 'success');
}

// Screen switching
function showAuthScreen() {
    document.getElementById('authScreen').classList.add('active');
    document.getElementById('appScreen').classList.remove('active');
    document.getElementById('navLogout').style.display = 'none';
    document.getElementById('cartLink').style.display = 'none';
    document.getElementById('userInfo').textContent = '';
    document.getElementById('topBar').style.display = 'block';
    document.getElementById('footer').style.display = 'none';
}

function showAppScreen() {
    document.getElementById('authScreen').classList.remove('active');
    document.getElementById('appScreen').classList.add('active');
    document.getElementById('navLogout').style.display = 'block';
    document.getElementById('cartLink').style.display = 'flex';
    document.getElementById('topBar').style.display = 'block';
    document.getElementById('footer').style.display = 'block';

    var userEmail = localStorage.getItem('userEmail') || ('User ' + userId.slice(0, 6));
    document.getElementById('userInfo').textContent = userEmail;
}

function goHome() {
    switchTab('shop');
}

function switchTab(tab) {
    document.querySelectorAll('.tab-content').forEach(function(el) { el.classList.remove('active'); });
    document.getElementById(tab + 'Tab').classList.add('active');

    if (tab === 'orders') loadOrders();
    else if (tab === 'cart') loadCart();
    else if (tab === 'shop') loadProducts();
}

// Products
async function loadProducts() {
    try {
        var products = await api.getProducts();
        products.forEach(function(p) { productsCache[p.id] = p; });
        document.getElementById('productsList').innerHTML = ui.renderProducts(products, productsCache);
    } catch (err) {
        console.error('Error loading products:', err);
        document.getElementById('productsList').innerHTML =
            '<div class="loading" style="grid-column: 1/-1; color: #721c24;">Failed to load products. Please try again.</div>';
    }
}

async function addToCart(productId) {
    var qty = parseInt(document.getElementById('qty-' + productId).value);
    var product = productsCache[productId];

    try {
        await api.addCartItem(userId, productId, qty, product.price);
        ui.showToast('Added ' + product.name + ' to cart!', 'success');
        loadCart();
    } catch (err) {
        ui.showToast(err.message, 'error');
    }
}

// Cart
async function loadCart() {
    try {
        var data = await api.getCart(userId);
        var items = data.items || [];
        ui.updateCartBadge(items.length);

        if (items.length === 0) {
            document.getElementById('cartContent').innerHTML = ui.renderEmptyCart();
            return;
        }

        document.getElementById('cartContent').innerHTML = ui.renderCart(items, productsCache);
    } catch (err) {
        console.error('Error loading cart:', err);
    }
}

async function removeFromCart(itemId) {
    try {
        await api.removeCartItem(userId, itemId);
        ui.showToast('Item removed from cart', 'success');
        loadCart();
    } catch (err) {
        ui.showToast(err.message, 'error');
    }
}

async function clearCart() {
    if (!confirm('Are you sure you want to empty your cart?')) return;
    try {
        await api.clearCart(userId);
        ui.showToast('Cart cleared', 'success');
        loadCart();
    } catch (err) {
        ui.showToast(err.message, 'error');
    }
}

async function checkout(total) {
    if (!confirm('Confirm your order for $' + total.toFixed(2) + '?')) return;

    try {
        var cartData = await api.getCart(userId);
        var order = await api.createOrder(userId, cartData.items);
        ui.showToast('Order placed! ID: ' + order.id.slice(0, 8), 'success');

        // Clear cart silently
        await api.clearCart(userId);
        loadCart();
        setTimeout(function() { switchTab('orders'); }, 1000);
    } catch (err) {
        ui.showToast(err.message, 'error');
    }
}

// Orders
async function loadOrders() {
    try {
        var orders = await api.getOrders(userId);

        if (orders.length === 0) {
            document.getElementById('ordersList').innerHTML = ui.renderEmptyOrders();
            return;
        }

        document.getElementById('ordersList').innerHTML = ui.renderOrders(orders, productsCache);
    } catch (err) {
        console.error('Error loading orders:', err);
        document.getElementById('ordersList').innerHTML =
            '<div class="cart-section" style="background: #f8d7da; color: #721c24;">Failed to load orders. Please try again.</div>';
    }
}

async function checkPaymentStatus(orderId) {
    try {
        var payment = await api.getPaymentStatus(orderId);
        if (!payment) {
            ui.showToast('Payment is still processing...', '');
            return;
        }
        var statusEmoji = payment.status === 'completed' ? '✅' : payment.status === 'failed' ? '❌' : '⏳';
        ui.showToast(statusEmoji + ' Payment ' + payment.status + ': $' + payment.amount.toFixed(2), payment.status === 'completed' ? 'success' : '');
        loadOrders();
    } catch (err) {
        ui.showToast('Payment status not available yet', '');
    }
}

// Initialize
window.addEventListener('load', function() {
    if (token && userId) {
        showAppScreen();
        loadProducts();
        loadCart();
    } else {
        showAuthScreen();
    }
});
