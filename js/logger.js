const Config = require("./config");

/**
 * 日志级别枚举
 */
const LOG_LEVELS = {
  ERROR: 0,
  WARN: 1,
  INFO: 2,
  DEBUG: 3,
};

/**
 * 日志颜色配置
 */
const LOG_COLORS = {
  ERROR: "\x1b[31m", // 红色
  WARN: "\x1b[33m", // 黄色
  INFO: "\x1b[36m", // 青色
  DEBUG: "\x1b[90m", // 灰色
  RESET: "\x1b[0m", // 重置颜色
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
 * 检查是否在终端环境中（支持颜色输出）
 * @returns {boolean} 是否支持颜色
 */
function supportsColor() {
  return process.stdout.isTTY && process.env.NODE_ENV !== "test";
}

/**
 * 统一的日志记录器
 */
class Logger {
  // 缓存当前日志级别，避免重复计算
  static _currentLogLevel = getLogLevelValue(Config.LOG_LEVEL);

  // 性能计时器缓存
  static _timers = new Map();

  // 是否支持颜色输出
  static _colorSupport = supportsColor();

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
   * 获取格式化的时间戳
   * @returns {string} 格式化的时间戳
   * @private
   */
  static _getTimestamp() {
    return new Date().toISOString();
  }

  /**
   * 格式化日志前缀
   * @param {string} level - 日志级别名称
   * @param {string} module - 模块名称
   * @returns {string} 格式化的前缀
   * @private
   */
  static _formatPrefix(level, module) {
    const timestamp = Logger._getTimestamp();
    const color = Logger._colorSupport ? LOG_COLORS[level] : "";
    const reset = Logger._colorSupport ? LOG_COLORS.RESET : "";

    return `${color}[${timestamp}] [${level} ${module}]:${reset}`;
  }

  /**
   * 通用日志输出方法
   * @param {string} level - 日志级别
   * @param {number} levelValue - 日志级别数值
   * @param {Function} logFn - 日志输出函数
   * @param {string} module - 模块名称
   * @param {...any} args - 日志参数
   * @private
   */
  static _log(level, levelValue, logFn, module, ...args) {
    if (Logger._shouldLog(levelValue)) {
      const prefix = Logger._formatPrefix(level, module);
      logFn(prefix, ...args);
    }
  }

  /**
   * 普通日志输出 (INFO级别)
   * @param {string} module - 模块名称
   * @param {...any} args - 日志参数
   */
  static log(module, ...args) {
    Logger._log("INFO", LOG_LEVELS.INFO, console.log, module, ...args);
  }

  /**
   * 信息日志输出 (INFO级别) - log方法的别名
   * @param {string} module - 模块名称
   * @param {...any} args - 日志参数
   */
  static info(module, ...args) {
    Logger.log(module, ...args);
  }

  /**
   * 调试日志输出 (DEBUG级别)
   * @param {string} module - 模块名称
   * @param {...any} args - 日志参数
   */
  static debug(module, ...args) {
    Logger._log("DEBUG", LOG_LEVELS.DEBUG, console.log, module, ...args);
  }

  /**
   * 警告日志输出 (WARN级别)
   * @param {string} module - 模块名称
   * @param {...any} args - 日志参数
   */
  static warn(module, ...args) {
    Logger._log("WARN", LOG_LEVELS.WARN, console.warn, module, ...args);
  }

  /**
   * 错误日志输出 (ERROR级别) - 总是输出
   * @param {string} module - 模块名称
   * @param {...any} args - 日志参数
   */
  static error(module, ...args) {
    const prefix = Logger._formatPrefix("ERROR", module);
    console.error(prefix, ...args);
  }

  /**
   * 性能测量开始
   * @param {string} module - 模块名称
   * @param {string} label - 性能标签
   */
  static timeStart(module, label) {
    if (Logger._shouldLog(LOG_LEVELS.DEBUG)) {
      const key = `${module}:${label}`;
      Logger._timers.set(key, process.hrtime.bigint());
      Logger.debug(module, `开始计时: ${label}`);
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
      const startTime = Logger._timers.get(key);

      if (startTime) {
        const endTime = process.hrtime.bigint();
        const duration = Number(endTime - startTime) / 1000000; // 转换为毫秒
        Logger._timers.delete(key);
        Logger.debug(module, `计时结束: ${label} - ${duration.toFixed(2)}ms`);
      } else {
        Logger.warn(module, `未找到计时器: ${label}`);
      }
    }
  }

  /**
   * 条件日志输出
   * @param {boolean} condition - 条件
   * @param {string} module - 模块名称
   * @param {...any} args - 日志参数
   */
  static logIf(condition, module, ...args) {
    if (condition) {
      Logger.log(module, ...args);
    }
  }

  /**
   * 条件调试日志输出
   * @param {boolean} condition - 条件
   * @param {string} module - 模块名称
   * @param {...any} args - 日志参数
   */
  static debugIf(condition, module, ...args) {
    if (condition) {
      Logger.debug(module, ...args);
    }
  }

  /**
   * 获取当前日志级别名称
   * @returns {string} 日志级别名称
   */
  static getCurrentLogLevel() {
    const levelNames = Object.keys(LOG_LEVELS);
    const currentLevel = Logger._currentLogLevel;
    return (
      levelNames.find((name) => LOG_LEVELS[name] === currentLevel) || "ERROR"
    );
  }

  /**
   * 设置日志级别
   * @param {string} level - 新的日志级别
   */
  static setLogLevel(level) {
    const newLevel = getLogLevelValue(level);
    if (newLevel !== undefined) {
      Logger._currentLogLevel = newLevel;
      Logger.info("Logger", `日志级别已更新为: ${level.toUpperCase()}`);
    } else {
      Logger.warn("Logger", `无效的日志级别: ${level}`);
    }
  }

  /**
   * 清理所有性能计时器
   */
  static clearTimers() {
    Logger._timers.clear();
    Logger.debug("Logger", "已清理所有性能计时器");
  }

  /**
   * 获取活跃的计时器数量
   * @returns {number} 活跃计时器数量
   */
  static getActiveTimersCount() {
    return Logger._timers.size;
  }

  /**
   * 输出分组开始标记（仅在DEBUG级别）
   * @param {string} module - 模块名称
   * @param {string} groupName - 分组名称
   */
  static groupStart(module, groupName) {
    Logger.debug(module, `开始分组: ${groupName}`);
  }

  /**
   * 输出分组结束标记（仅在DEBUG级别）
   * @param {string} module - 模块名称
   * @param {string} groupName - 分组名称
   */
  static groupEnd(module, groupName) {
    Logger.debug(module, `结束分组: ${groupName}`);
  }

  /**
   * 输出表格数据（仅在DEBUG级别）
   * @param {string} module - 模块名称
   * @param {string} title - 表格标题
   * @param {Array|Object} data - 表格数据
   */
  static table(module, title, data) {
    if (Logger._shouldLog(LOG_LEVELS.DEBUG)) {
      Logger.debug(module, `表格: ${title}`);
      console.table(data);
    }
  }
}

// 导出日志级别常量，供外部使用
Logger.LOG_LEVELS = LOG_LEVELS;

module.exports = Logger;
