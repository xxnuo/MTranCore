const Config = require("./config");

/**
 * 统一的日志记录器
 */
class Logger {
  static log(module, ...args) {
    if (Config.LOG_LEVEL === "Info" || Config.LOG_LEVEL === "Debug") {
      console.log(`[${module}]:`, ...args);
    }
  }

  static debug(module, ...args) {
    if (Config.LOG_LEVEL === "Debug") {
      console.log(`[DEBUG ${module}]:`, ...args);
    }
  }

  static warn(module, ...args) {
    if (Config.LOG_LEVEL === "Warn" || Config.LOG_LEVEL === "Info" || Config.LOG_LEVEL === "Debug") {
      console.warn(`[WARN ${module}]:`, ...args);
    }
  }

  static error(module, ...args) {
    console.error(`[ERROR ${module}]:`, ...args);
  }
}

module.exports = Logger; 