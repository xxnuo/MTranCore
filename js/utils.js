/**
 * 通用工具函数集合
 */

/**
 * 延迟执行函数
 * @param {number} ms - 延迟时间（毫秒）
 * @returns {Promise} Promise对象
 */
function delay(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

/**
 * 安全的JSON解析
 * @param {string} jsonString - JSON字符串
 * @param {any} [defaultValue=null] - 解析失败时的默认值
 * @returns {any} 解析结果或默认值
 */
function safeJsonParse(jsonString, defaultValue = null) {
  try {
    return JSON.parse(jsonString);
  } catch (error) {
    return defaultValue;
  }
}

/**
 * 安全的JSON字符串化
 * @param {any} obj - 要序列化的对象
 * @param {string} [defaultValue="{}"] - 序列化失败时的默认值
 * @returns {string} JSON字符串或默认值
 */
function safeJsonStringify(obj, defaultValue = "{}") {
  try {
    return JSON.stringify(obj);
  } catch (error) {
    return defaultValue;
  }
}

/**
 * 防抖函数
 * @param {Function} func - 要防抖的函数
 * @param {number} wait - 等待时间（毫秒）
 * @returns {Function} 防抖后的函数
 */
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

/**
 * 节流函数
 * @param {Function} func - 要节流的函数
 * @param {number} limit - 时间间隔（毫秒）
 * @returns {Function} 节流后的函数
 */
function throttle(func, limit) {
  let inThrottle;
  return function executedFunction(...args) {
    if (!inThrottle) {
      func.apply(this, args);
      inThrottle = true;
      setTimeout(() => (inThrottle = false), limit);
    }
  };
}

/**
 * 重试执行函数
 * @param {Function} fn - 要执行的异步函数
 * @param {number} [maxRetries=3] - 最大重试次数
 * @param {number} [retryDelay=1000] - 重试间隔（毫秒）
 * @returns {Promise} Promise对象
 */
async function retry(fn, maxRetries = 3, retryDelay = 1000) {
  let lastError;

  for (let i = 0; i <= maxRetries; i++) {
    try {
      return await fn();
    } catch (error) {
      lastError = error;

      if (i === maxRetries) {
        throw lastError;
      }

      if (retryDelay > 0) {
        await delay(retryDelay);
      }
    }
  }
}

/**
 * 深度克隆对象
 * @param {any} obj - 要克隆的对象
 * @returns {any} 克隆后的对象
 */
function deepClone(obj) {
  if (obj === null || typeof obj !== "object") {
    return obj;
  }

  if (obj instanceof Date) {
    return new Date(obj.getTime());
  }

  if (obj instanceof Array) {
    return obj.map((item) => deepClone(item));
  }

  if (typeof obj === "object") {
    const clonedObj = {};
    for (const key in obj) {
      if (obj.hasOwnProperty(key)) {
        clonedObj[key] = deepClone(obj[key]);
      }
    }
    return clonedObj;
  }

  return obj;
}

/**
 * 检查值是否为空（null、undefined、空字符串、空数组、空对象）
 * @param {any} value - 要检查的值
 * @returns {boolean} 是否为空
 */
function isEmpty(value) {
  if (value === null || value === undefined) {
    return true;
  }

  if (typeof value === "string" && value.trim() === "") {
    return true;
  }

  if (Array.isArray(value) && value.length === 0) {
    return true;
  }

  if (typeof value === "object" && Object.keys(value).length === 0) {
    return true;
  }

  return false;
}

/**
 * 生成随机字符串
 * @param {number} [length=8] - 字符串长度
 * @param {string} [charset='ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789'] - 字符集
 * @returns {string} 随机字符串
 */
function randomString(
  length = 8,
  charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
) {
  let result = "";
  for (let i = 0; i < length; i++) {
    result += charset.charAt(Math.floor(Math.random() * charset.length));
  }
  return result;
}

/**
 * 格式化文件大小
 * @param {number} bytes - 字节数
 * @param {number} [decimals=2] - 小数位数
 * @returns {string} 格式化后的文件大小
 */
function formatFileSize(bytes, decimals = 2) {
  if (bytes === 0) return "0 Bytes";

  const k = 1024;
  const dm = decimals < 0 ? 0 : decimals;
  const sizes = ["Bytes", "KB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"];

  const i = Math.floor(Math.log(bytes) / Math.log(k));

  return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + " " + sizes[i];
}

/**
 * 创建超时Promise
 * @param {number} ms - 超时时间（毫秒）
 * @param {string} [message='Operation timed out'] - 超时错误消息
 * @returns {Promise} 超时Promise
 */
function timeout(ms, message = "Operation timed out") {
  return new Promise((_, reject) => {
    setTimeout(() => reject(new Error(message)), ms);
  });
}

/**
 * 带超时的Promise包装器
 * @param {Promise} promise - 原始Promise
 * @param {number} ms - 超时时间（毫秒）
 * @param {string} [message='Operation timed out'] - 超时错误消息
 * @returns {Promise} 带超时的Promise
 */
function withTimeout(promise, ms, message = "Operation timed out") {
  return Promise.race([promise, timeout(ms, message)]);
}

/**
 * 数组分块
 * @param {Array} array - 要分块的数组
 * @param {number} size - 每块的大小
 * @returns {Array} 分块后的二维数组
 */
function chunk(array, size) {
  const chunks = [];
  for (let i = 0; i < array.length; i += size) {
    chunks.push(array.slice(i, i + size));
  }
  return chunks;
}

/**
 * 数组去重
 * @param {Array} array - 要去重的数组
 * @param {Function} [keyFn] - 获取唯一键的函数
 * @returns {Array} 去重后的数组
 */
function unique(array, keyFn) {
  if (!keyFn) {
    return [...new Set(array)];
  }

  const seen = new Set();
  return array.filter((item) => {
    const key = keyFn(item);
    if (seen.has(key)) {
      return false;
    }
    seen.add(key);
    return true;
  });
}

/**
 * 安全地获取嵌套对象属性
 * @param {Object} obj - 对象
 * @param {string} path - 属性路径（如 'a.b.c'）
 * @param {any} [defaultValue] - 默认值
 * @returns {any} 属性值或默认值
 */
function get(obj, path, defaultValue) {
  const keys = path.split(".");
  let result = obj;

  for (const key of keys) {
    if (result === null || result === undefined || !(key in result)) {
      return defaultValue;
    }
    result = result[key];
  }

  return result;
}

/**
 * 安全地设置嵌套对象属性
 * @param {Object} obj - 对象
 * @param {string} path - 属性路径（如 'a.b.c'）
 * @param {any} value - 要设置的值
 * @returns {Object} 修改后的对象
 */
function set(obj, path, value) {
  const keys = path.split(".");
  let current = obj;

  for (let i = 0; i < keys.length - 1; i++) {
    const key = keys[i];
    if (
      !(key in current) ||
      typeof current[key] !== "object" ||
      current[key] === null
    ) {
      current[key] = {};
    }
    current = current[key];
  }

  current[keys[keys.length - 1]] = value;
  return obj;
}

module.exports = {
  delay,
  safeJsonParse,
  safeJsonStringify,
  debounce,
  throttle,
  retry,
  deepClone,
  isEmpty,
  randomString,
  formatFileSize,
  timeout,
  withTimeout,
  chunk,
  unique,
  get,
  set,
};
