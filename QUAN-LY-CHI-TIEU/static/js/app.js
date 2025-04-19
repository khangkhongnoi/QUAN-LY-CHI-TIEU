// Format tiền tệ
function formatMoney(amount) {
    return new Intl.NumberFormat('vi-VN', {
        style: 'currency',
        currency: 'VND'
    }).format(amount).replace('₫', 'VND');
}

// Load dữ liệu
async function loadData() {
    try {
        // Load dữ liệu song song
        const [summaryRes, expensesRes] = await Promise.all([
            fetch('/summary'),
            fetch('/expenses')
        ]);

        // Xử lý thống kê
        const summary = await summaryRes.json();
        document.getElementById('dailyTotal').textContent = formatMoney(summary.daily);
        document.getElementById('monthlyTotal').textContent = formatMoney(summary.monthly);

        // Xử lý danh sách chi tiêu
        const expenses = await expensesRes.json();
        const tbody = document.getElementById('expensesList');
        tbody.innerHTML = expenses.map(expense => `
            <tr>
                <td>${expense.Category}</td>
                <td class="text-nowrap">${formatMoney(expense.Amount)}</td>
                <td>${new Date(expense.CreatedAt).toLocaleDateString('vi-VN')} ${new Date(expense.CreatedAt).toLocaleTimeString('vi-VN', {hour: '2-digit', minute:'2-digit'})}</td>
            </tr>
        `).join('');

    } catch (error) {
        console.error('Lỗi khi tải dữ liệu:', error);
    }
}

// Load categories với xử lý lỗi
async function loadCategories() {
    try {
        const response = await fetch('/categories');
        if (!response.ok) throw new Error('Network response was not ok');

        const categories = await response.json();
        const select = document.getElementById('category');

        select.innerHTML = `
            <option value="">Chọn danh mục</option>
            ${categories.map(cat => `
                <option value="${cat.ID}">${cat.Name}</option>
            `).join('')}
            <option value="5">Đồ ăn khác</option>
        `;

        // Xử lý hiển thị ô nhập mới
        select.addEventListener('change', () => {
            document.getElementById('newCategorySection').style.display =
                select.value === '5' ? 'block' : 'none';
        });

    } catch (error) {
        console.error('Lỗi khi tải danh mục:', error);
    }
}