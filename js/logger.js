const Config = require("./config");

/**
 * 日志级别枚举
 */
const LOG_LEVELS = {
  ERROR: 0,
  WARN: 1,
  INFO: 2,
  DEBUG: 3
};

/**
 * 将字符串日志级别转换为数字
 * @param {string} level - 日志级别字符串
 * @returns {number} 数字日志级别
 */
function getLogLevelValue(level) {
  const upperLevel = level?.toUpperCase();
  return LOG_LEVELS[upperLevel] ?? LOG_LEVELS.ERROR;
}

/**
 * 统一的日志记录器
 */
class Logger {
  // 缓存当前日志级别，避免重复计算
  static _currentLogLevel = getLogLevelValue(Config.LOG_LEVEL);

  /**
   * 更新日志级别缓存
   */
  static updateLogLevel() {
    Logger._currentLogLevel = getLogLevelValue(Config.LOG_LEVEL);
  }

  /**
   * 检查是否应该输出指定级别的日志
   * @param {number} level - 要检查的日志级别
   * @returns {boolean} 是否应该输出
   * @private
   */
  static _shouldLog(level) {
    return Logger._currentLogLevel >= level;
  }

  /**
   * 格式化日志前缀
   * @param {string} level - 日志级别名称
   * @param {string} module - 模块名称
   * @returns {string} 格式化的前缀
   * @private
   */
  static _formatPrefix(level, module) {
    const timestamp = new Date().toISOString();
    return `[${timestamp}] [${level} ${module}]:`;
  }

  /**
   * 普通日志输出 (INFO级别)
   * @param {string} module - 模块名称
   * @param {...any} args - 日志参数
   */
  static log(module, ...args) {
    if (Logger._shouldLog(LOG_LEVELS.INFO)) {
      console.log(Logger._formatPrefix("INFO", module), ...args);
    }
  }

  /**
   * 调试日志输出 (DEBUG级别)
   * @param {string} module - 模块名称
   * @param {...any} args - 日志参数
   */
  static debug(module, ...args) {
    if (Logger._shouldLog(LOG_LEVELS.DEBUG)) {
      console.log(Logger._formatPrefix("DEBUG", module), ...args);
    }
  }

  /**
   * 警告日志输出 (WARN级别)
   * @param {string} module - 模块名称
   * @param {...any} args - 日志参数
   */
  static warn(module, ...args) {
    if (Logger._shouldLog(LOG_LEVELS.WARN)) {
      console.warn(Logger._formatPrefix("WARN", module), ...args);
    }
  }

  /**
   * 错误日志输出 (ERROR级别) - 总是输出
   * @param {string} module - 模块名称
   * @param {...any} args - 日志参数
   */
  static error(module, ...args) {
    console.error(Logger._formatPrefix("ERROR", module), ...args);
  }

  /**
   * 性能测量开始
   * @param {string} module - 模块名称
   * @param {string} label - 性能标签
   */
  static timeStart(module, label) {
    if (Logger._shouldLog(LOG_LEVELS.DEBUG)) {
      const key = `${module}:${label}`;
      console.time(key);
    }
  }

  /**
   * 性能测量结束
   * @param {string} module - 模块名称
   * @param {string} label - 性能标签
   */
  static timeEnd(module, label) {
    if (Logger._shouldLog(LOG_LEVELS.DEBUG)) {
      const key = `${module}:${label}`;
      console.timeEnd(key);
    }
  }

  /**
   * 条件日志输出
   * @param {boolean} condition - 条件
   * @param {string} module - 模块名称
   * @param {...any} args - 日志参数
   */
  static logIf(condition, module, ...args) {
    if (condition && Logger._shouldLog(LOG_LEVELS.INFO)) {
      console.log(Logger._formatPrefix("INFO", module), ...args);
    }
  }

  /**
   * 获取当前日志级别名称
   * @returns {string} 日志级别名称
   */
  static getCurrentLogLevel() {
    const levelNames = Object.keys(LOG_LEVELS);
    const currentLevel = Logger._currentLogLevel;
    return levelNames.find(name => LOG_LEVELS[name] === currentLevel) || "ERROR";
  }
}

module.exports = Logger; 