// 发起 API 请求
async function makeRequest(url, method = 'GET', body = null) {
    const options = {
        method,
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
    };
    if (body) {
        options.body = JSON.stringify(body);
    }
    const response = await fetch(url, options);
    return await response.json();
}

// 检查身份验证
async function checkAuth(redirectTo, requiredRole = null) {
    const userId = localStorage.getItem('user_id');
    const role = localStorage.getItem('role');
    if (!userId) {
        location.href = '/login';
        return null;
    }
    if (requiredRole && role !== requiredRole) {
        showMessage('error-message', '需要管理员权限！');
        setTimeout(() => location.href = redirectTo, 1000);
        return null;
    }
    return { userId, role };
}

// 显示消息（成功或错误）
function showMessage(elementId, message, isError = true) {
    const errorDiv = document.getElementById(elementId);
    if (errorDiv) {
        errorDiv.textContent = message;
        errorDiv.classList.remove('hidden');
        errorDiv.classList.toggle('text-red-600', isError);
        errorDiv.classList.toggle('text-green-600', !isError);
        // 自动隐藏消息
        setTimeout(() => {
            errorDiv.classList.add('hidden');
            errorDiv.textContent = '';
        }, 3000);
    } else {
        console.error('消息显示失败：未找到元素', elementId);
    }
}