/**
 * GitLab Merge Alert - 公共JavaScript功能库
 * 提供跨页面共享的通用功能函数
 */

// ==================== 通用工具函数 ====================

/**
 * 格式化日期字符串为本地化格式
 * @param {string} dateString - ISO日期字符串
 * @returns {string} 格式化后的日期字符串
 */
function formatDate(dateString) {
    if (!dateString) return '-';
    return new Date(dateString).toLocaleString('zh-CN');
}

/**
 * 格式化手机号，中间部分用*代替
 * @param {string} phone - 手机号码
 * @returns {string} 格式化后的手机号
 */
function formatPhone(phone) {
    if (!phone || phone.length < 7) return phone;
    const start = phone.substring(0, 3);
    const end = phone.substring(phone.length - 4);
    const middle = '*'.repeat(phone.length - 7);
    return start + middle + end;
}

/**
 * 从邮箱中提取用户名部分
 * @param {string} email - 邮箱地址
 * @returns {string} 用户名部分
 */
function extractNameFromEmail(email) {
    if (!email || email === 'REDACTED' || email.trim() === '') {
        return '-';
    }
    
    if (email.includes('@')) {
        const username = email.split('@')[0];
        return username === 'REDACTED' ? '-' : username;
    }
    
    return email === 'REDACTED' ? '-' : email;
}

// ==================== Bootstrap Tooltip管理 ====================

/**
 * 初始化或重新初始化所有的Bootstrap tooltips
 * 销毁现有实例并创建新的，确保动态内容的tooltip正常工作
 */
function initTooltips() {
    // 销毁旧的tooltip实例
    document.querySelectorAll('[data-bs-toggle="tooltip"]').forEach(el => {
        const existingTooltip = bootstrap.Tooltip.getInstance(el);
        if (existingTooltip) {
            existingTooltip.dispose();
        }
    });
    
    // 创建新的tooltip实例
    const tooltipTriggerList = [].slice.call(document.querySelectorAll('[data-bs-toggle="tooltip"]'));
    tooltipTriggerList.map(function (tooltipTriggerEl) {
        return new bootstrap.Tooltip(tooltipTriggerEl);
    });
}

// ==================== 导航功能 ====================

/**
 * 高亮当前页面的导航链接
 * 在页面加载时自动执行
 */
function highlightCurrentNavigation() {
    document.addEventListener('DOMContentLoaded', function() {
        const currentPath = window.location.pathname;
        const navLinks = document.querySelectorAll('.nav-link');
        navLinks.forEach(link => {
            if (link.getAttribute('href') === currentPath) {
                link.classList.add('active');
            }
        });
    });
}

// ==================== Vue.js 通用Mixin ====================

/**
 * Vue.js 应用的通用mixin，提供跨组件共享的方法
 */
const CommonMixin = {
    methods: {
        // 日期格式化
        formatDate: formatDate,
        
        // 手机号格式化
        formatPhone: formatPhone,
        
        // 邮箱用户名提取
        extractNameFromEmail: extractNameFromEmail,
        
        // 初始化tooltips
        initTooltips: initTooltips,
        
        /**
         * 切换分组的展开/收起状态
         * @param {string} groupName - 分组名称
         */
        toggleGroup(groupName) {
            this.collapsedGroups[groupName] = !this.collapsedGroups[groupName];
            this.$forceUpdate();
        },
        
        /**
         * 切换项目的展开/收起状态
         * @param {string} projectName - 项目名称
         */
        toggleProject(projectName) {
            this.collapsedProjects[projectName] = !this.collapsedProjects[projectName];
            this.$forceUpdate();
        }
    },
    
    mounted() {
        // 页面挂载后初始化tooltips
        this.$nextTick(() => {
            this.initTooltips();
        });
    },
    
    updated() {
        // 数据更新后重新初始化tooltips
        this.$nextTick(() => {
            this.initTooltips();
        });
    }
};

// ==================== 错误处理 ====================

/**
 * 统一的错误处理函数
 * @param {Error} error - 错误对象
 * @param {string} operation - 操作描述
 */
function handleError(error, operation = '操作') {
    console.error(`${operation}失败:`, error);
    const message = error.response?.data?.message || 
                   error.response?.data?.error || 
                   error.message || 
                   '未知错误';
    alert(`${operation}失败: ${message}`);
}

// ==================== API请求辅助函数 ====================

/**
 * 创建标准的axios请求配置
 * @param {Object} options - 请求选项
 * @returns {Object} axios配置对象
 */
function createApiConfig(options = {}) {
    return {
        timeout: 10000,
        ...options,
        headers: {
            'Content-Type': 'application/json',
            ...options.headers
        }
    };
}

// ==================== 自动执行 ====================

// 页面加载时自动初始化导航高亮
highlightCurrentNavigation();

// 页面加载完成后初始化tooltips
document.addEventListener('DOMContentLoaded', function() {
    initTooltips();
});

// ==================== 导出（如果支持模块系统） ====================
if (typeof module !== 'undefined' && module.exports) {
    module.exports = {
        formatDate,
        formatPhone,
        extractNameFromEmail,
        initTooltips,
        highlightCurrentNavigation,
        CommonMixin,
        handleError,
        createApiConfig
    };
}