window.api = (function() {
    const API_URL = '';

    async function register(email, password) {
        const res = await fetch(`${API_URL}/api/auth/register`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ email, password })
        });
        const data = await res.json();
        if (!res.ok) throw new Error(data.error || 'Registration failed');
        return data;
    }

    async function loginUser(email, password) {
        const res = await fetch(`${API_URL}/api/auth/login`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ email, password })
        });
        const data = await res.json();
        if (!res.ok) throw new Error(data.error || 'Login failed');
        return data;
    }

    async function getProducts() {
        const res = await fetch(`${API_URL}/api/products/`);
        return await res.json() || [];
    }

    async function getCart(userId) {
        const res = await fetch(`${API_URL}/api/cart/`, {
            headers: { 'X-User-ID': userId }
        });
        return await res.json();
    }

    async function addCartItem(userId, productId, quantity, price) {
        const res = await fetch(`${API_URL}/api/cart/items`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-User-ID': userId
            },
            body: JSON.stringify({ product_id: productId, quantity: quantity, price: price })
        });
        if (!res.ok) throw new Error('Failed to add to cart');
        return await res.json();
    }

    async function removeCartItem(userId, itemId) {
        const res = await fetch(`${API_URL}/api/cart/items/${itemId}`, {
            method: 'DELETE',
            headers: { 'X-User-ID': userId }
        });
        if (!res.ok) throw new Error('Failed to remove item');
        return res;
    }

    async function clearCart(userId) {
        const res = await fetch(`${API_URL}/api/cart/`, {
            method: 'DELETE',
            headers: { 'X-User-ID': userId }
        });
        if (!res.ok) throw new Error('Failed to clear cart');
        return res;
    }

    async function createOrder(userId, items) {
        const res = await fetch(`${API_URL}/api/orders/`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-User-ID': userId
            },
            body: JSON.stringify({ items: items })
        });
        if (!res.ok) throw new Error('Failed to create order');
        return await res.json();
    }

    async function getOrders(userId) {
        const res = await fetch(`${API_URL}/api/orders/`, {
            headers: { 'X-User-ID': userId }
        });
        return await res.json() || [];
    }

    async function getPaymentStatus(orderId) {
        const res = await fetch(`${API_URL}/api/payments/${orderId}`);
        if (!res.ok) return null;
        return await res.json();
    }

    return {
        register,
        loginUser,
        getProducts,
        getCart,
        addCartItem,
        removeCartItem,
        clearCart,
        createOrder,
        getOrders,
        getPaymentStatus
    };
})();
