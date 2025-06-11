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
        alert('请先登录！');
        location.href = '/login';
        return null;
    }
    if (requiredRole && role !== requiredRole) {
        alert('需要管理员权限！');
        location.href = redirectTo;
        return null;
    }
    return { userId, role };
}

// 显示错误消息
function showError(elementId, message) {
    const errorDiv = document.getElementById(elementId);
    if (errorDiv) {
        errorDiv.textContent = message;
        errorDiv.classList.remove('hidden');
    } else {
        alert(message);
    }
}