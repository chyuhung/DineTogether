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
    try {
        const response = await fetch(url, options);
        const contentType = response.headers.get('content-type');
        if (!response.ok) {
            let errorMessage = '请求失败，请稍后重试';
            if (contentType && contentType.includes('application/json')) {
                const result = await response.json();
                errorMessage = result.error || errorMessage;
            }
            throw new Error(errorMessage);
        }
        if (contentType && contentType.includes('application/json')) {
            return await response.json();
        } else {
            throw new Error('服务器返回了非 JSON 响应');
        }
    } catch (error) {
        throw new Error(error.message || '网络错误，请稍后重试');
    }
}

// 上传图片
async function uploadImages(files) {
    if (files.length === 0) return [];
    const formData = new FormData();
    for (const file of files) {
        formData.append('images', file);
    }
    try {
        const response = await fetch('/upload-image', {
            method: 'POST',
            body: formData,
            credentials: 'include',
        });
        const contentType = response.headers.get('content-type');
        if (!response.ok) {
            let errorMessage = '图片上传失败';
            if (contentType && contentType.includes('application/json')) {
                const result = await response.json();
                errorMessage = result.error || errorMessage;
            }
            throw new Error(errorMessage);
        }
        if (contentType && contentType.includes('application/json')) {
            const result = await response.json();
            if (result.message === '图片上传成功') {
                return result.image_urls;
            }
            throw new Error(result.error || '图片上传失败');
        } else {
            throw new Error('服务器返回了非 JSON 响应');
        }
    } catch (error) {
        throw new Error(error.message || '图片上传失败');
    }
}

// 删除图片
async function deleteImage(imageUrl) {
    try {
        const result = await makeRequest('/delete-image', 'POST', { image_url: imageUrl });
        if (result.message !== '图片删除成功') {
            throw new Error(result.error || '图片删除失败');
        }
    } catch (error) {
        throw new Error(error.message || '图片删除失败');
    }
}

// 检查身份验证
async function checkAuth(redirectTo, requiredRole = null) {
    const userId = localStorage.getItem('user_id');
    const role = localStorage.getItem('role') || 'guest';
    if (!userId && redirectTo !== '/login') {
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
        setTimeout(() => {
            errorDiv.classList.add('hidden');
            errorDiv.textContent = '';
        }, 3000);
    } else {
        console.error('消息显示失败：未找到元素', elementId);
    }
}