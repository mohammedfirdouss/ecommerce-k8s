window.ui = (function() {

    function escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    function hashCode(str) {
        let hash = 0;
        for (let i = 0; i < str.length; i++) {
            hash = str.charCodeAt(i) + ((hash << 5) - hash);
        }
        return hash;
    }

    function getProductImage(product) {
        const colors = [
            ['#667eea', '#764ba2'],
            ['#f093fb', '#f5576c'],
            ['#4facfe', '#00f2fe'],
            ['#43e97b', '#38f9d7'],
            ['#fa709a', '#fee140'],
            ['#a8edea', '#fed6e3'],
            ['#ff9a9e', '#fecfef'],
            ['#ffecd2', '#fcb69f']
        ];
        const idx = Math.abs(hashCode(product.id)) % colors.length;
        return colors[idx];
    }

    function getProductEmoji(product) {
        const name = (product.name || '').toLowerCase();
        if (name.includes('phone') || name.includes('laptop') || name.includes('computer')) return '📱';
        if (name.includes('shirt') || name.includes('dress') || name.includes('jacket')) return '👕';
        if (name.includes('shoe') || name.includes('sneaker') || name.includes('boot')) return '👟';
        if (name.includes('watch') || name.includes('ring') || name.includes('necklace')) return '⌚';
        if (name.includes('chair') || name.includes('table') || name.includes('lamp')) return '🏠';
        if (name.includes('makeup') || name.includes('perfume') || name.includes('cream')) return '💄';
        if (name.includes('ball') || name.includes('gym') || name.includes('fitness')) return '⚽';
        return '📦';
    }

    function formatStatus(status) {
        const statusMap = {
            'pending': '⏳ Pending',
            'completed': '✅ Completed',
            'failed': '❌ Failed'
        };
        return statusMap[status.toLowerCase()] || status;
    }

    function updateCartBadge(count) {
        const badge = document.getElementById('cartBadge');
        if (count > 0) {
            badge.textContent = count;
            badge.style.display = 'flex';
        } else {
            badge.style.display = 'none';
        }
    }

    function showToast(message, type) {
        type = type || '';
        const toast = document.getElementById('toast');
        toast.textContent = message;
        toast.className = 'toast show ' + type;
        setTimeout(function() {
            toast.classList.remove('show');
        }, 3000);
    }

    function showAlert(elementId, message, type) {
        const alert = document.getElementById(elementId);
        alert.textContent = message;
        alert.className = 'alert ' + type;
    }

    function renderProducts(products, productsCache) {
        products.forEach(function(p) { productsCache[p.id] = p; });

        if (products.length === 0) {
            return '<div class="loading" style="grid-column: 1/-1;">No products available. Add some products to get started!</div>';
        }

        return products.map(function(p) {
            var colors = getProductImage(p);
            return '<div class="product-card">' +
                '<div class="product-image-container" style="background: linear-gradient(135deg, ' + colors[0] + ' 0%, ' + colors[1] + ' 100%);">' +
                    '<div class="product-image" style="display: flex; align-items: center; justify-content: center; font-size: 64px; color: white;">' +
                        getProductEmoji(p) +
                    '</div>' +
                    '<div class="product-image-overlay"></div>' +
                '</div>' +
                '<div class="product-info">' +
                    '<div class="product-name">' + escapeHtml(p.name) + '</div>' +
                    '<div class="product-description">' + escapeHtml(p.description || 'Premium quality product') + '</div>' +
                    '<div class="product-price">$' + p.price.toFixed(2) + '</div>' +
                    '<div class="product-stock">' + (p.stock > 0 ? p.stock + ' in stock' : 'Out of stock') + '</div>' +
                    '<div class="product-actions">' +
                        '<div class="quantity-select">' +
                            '<select id="qty-' + p.id + '">' +
                                [1,2,3,4,5].map(function(n) { return '<option value="' + n + '">' + n + '</option>'; }).join('') +
                            '</select>' +
                        '</div>' +
                        '<button class="cymbal-button-primary" onclick="addToCart(\'' + p.id + '\')"' + (p.stock <= 0 ? ' disabled style="opacity:0.5"' : '') + '>' +
                            'Add to Cart' +
                        '</button>' +
                    '</div>' +
                '</div>' +
            '</div>';
        }).join('');
    }

    function renderCart(items, productsCache) {
        var subtotal = items.reduce(function(sum, item) { return sum + (item.price * item.quantity); }, 0);
        var shipping = subtotal > 50 ? 0 : 5.99;
        var total = subtotal + shipping;

        return '<div class="cart-layout">' +
            '<div class="cart-section">' +
                '<div class="cart-header">' +
                    '<h3>Cart (' + items.length + ')</h3>' +
                    '<div style="display: flex; gap: 12px;">' +
                        '<button class="cymbal-button-secondary cymbal-button-small" onclick="clearCart()">Empty Cart</button>' +
                        '<button class="cymbal-button-primary cymbal-button-small" onclick="switchTab(\'shop\')">Continue Shopping</button>' +
                    '</div>' +
                '</div>' +
                items.map(function(item) {
                    var product = productsCache[item.product_id] || { name: item.product_id };
                    var colors = getProductImage({ id: item.product_id });
                    return '<div class="cart-item">' +
                        '<div class="cart-item-image" style="background: linear-gradient(135deg, ' + colors[0] + ' 0%, ' + colors[1] + ' 100%); color: white;">' +
                            getProductEmoji(product) +
                        '</div>' +
                        '<div class="cart-item-details">' +
                            '<div class="cart-item-name">' + escapeHtml(product.name || item.product_id) + '</div>' +
                            '<div class="cart-item-sku">SKU #' + item.product_id.slice(0, 8) + '</div>' +
                            '<div class="cart-item-qty">Quantity: ' + item.quantity + '</div>' +
                            '<div class="cart-item-actions">' +
                                '<button class="cymbal-button-secondary cymbal-button-small" onclick="removeFromCart(\'' + item.id + '\')">Remove</button>' +
                            '</div>' +
                        '</div>' +
                        '<div class="cart-item-price">$' + (item.price * item.quantity).toFixed(2) + '</div>' +
                    '</div>';
                }).join('') +
            '</div>' +
            '<div class="cart-summary-section">' +
                '<h3 style="margin-bottom: 24px;">Order Summary</h3>' +
                '<div class="summary-row"><span>Subtotal</span><span>$' + subtotal.toFixed(2) + '</span></div>' +
                '<div class="summary-row"><span>Shipping</span><span>' + (shipping === 0 ? 'FREE' : '$' + shipping.toFixed(2)) + '</span></div>' +
                '<div class="summary-row total"><span>Total</span><span>$' + total.toFixed(2) + '</span></div>' +
                '<button class="cymbal-button-primary" onclick="checkout(' + total + ')" style="width: 100%; margin-top: 24px; padding: 14px;">Place Order</button>' +
                '<p style="text-align: center; font-size: 13px; color: #605f64; margin-top: 16px;">🔒 Secure checkout</p>' +
            '</div>' +
        '</div>';
    }

    function renderEmptyCart() {
        return '<div class="cart-section">' +
            '<div class="cart-empty">' +
                '<div style="font-size: 64px; margin-bottom: 16px;">🛒</div>' +
                '<h4>Your shopping cart is empty!</h4>' +
                '<p>Items you add to your shopping cart will appear here.</p>' +
                '<button class="cymbal-button-primary" onclick="switchTab(\'shop\')" style="margin-top: 24px;">Continue Shopping</button>' +
            '</div>' +
        '</div>';
    }

    function renderOrders(orders, productsCache) {
        return orders.map(function(order) {
            return '<div class="order-card">' +
                '<div class="order-header">' +
                    '<div>' +
                        '<div class="order-id">Order #' + order.id.slice(0, 8).toUpperCase() + '</div>' +
                        '<div class="order-date">' + new Date(order.created_at).toLocaleDateString('en-US', {
                            year: 'numeric',
                            month: 'long',
                            day: 'numeric',
                            hour: '2-digit',
                            minute: '2-digit'
                        }) + '</div>' +
                    '</div>' +
                    '<span class="order-status status-' + order.status.toLowerCase() + '">' + formatStatus(order.status) + '</span>' +
                '</div>' +
                '<div class="order-body">' +
                    '<ul class="order-items">' +
                        (order.items || []).map(function(item) {
                            var product = productsCache[item.product_id] || { name: item.product_id };
                            return '<li class="order-item">' +
                                '<span>' + escapeHtml(product.name || item.product_id) + ' × ' + item.quantity + '</span>' +
                                '<span>$' + (item.price * item.quantity).toFixed(2) + '</span>' +
                            '</li>';
                        }).join('') +
                    '</ul>' +
                    '<div class="order-total"><span>Total</span><span>$' + order.total.toFixed(2) + '</span></div>' +
                    '<div class="order-actions">' +
                        '<button class="cymbal-button-secondary cymbal-button-small" onclick="checkPaymentStatus(\'' + order.id + '\')">Check Payment Status</button>' +
                    '</div>' +
                '</div>' +
            '</div>';
        }).join('');
    }

    function renderEmptyOrders() {
        return '<div class="cart-section">' +
            '<div class="cart-empty">' +
                '<div style="font-size: 64px; margin-bottom: 16px;">📋</div>' +
                '<h4>No orders yet</h4>' +
                '<p>When you place orders, they will appear here.</p>' +
                '<button class="cymbal-button-primary" onclick="switchTab(\'shop\')" style="margin-top: 24px;">Start Shopping</button>' +
            '</div>' +
        '</div>';
    }

    return {
        escapeHtml,
        hashCode,
        getProductImage,
        getProductEmoji,
        formatStatus,
        updateCartBadge,
        showToast,
        showAlert,
        renderProducts,
        renderCart,
        renderEmptyCart,
        renderOrders,
        renderEmptyOrders
    };
})();
