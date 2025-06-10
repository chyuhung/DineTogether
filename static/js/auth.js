async function checkAuth(redirectTo = '/login', requiredRole = null) {
    try {
        const response = await fetch('/api/user', { credentials: 'include' });
        const result = await response.json();
        if (result.error) {
            alert('请先登录！');
            location.href = redirectTo;
            return null;
        }
        if (requiredRole && result.role !== requiredRole) {
            alert(`需要${requiredRole}权限！`);
            location.href = '/';
            return null;
        }
        return result; // 返回用户信息 { id, username, role }
    } catch (error) {
        alert('网络错误，请稍后重试！');
        location.href = redirectTo;
        return null;
    }
}