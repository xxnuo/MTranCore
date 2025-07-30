const path = require("path");
const os = require("os");

// 配置类，用于获取环境变量
class Config {
  // 是否离线模式，离线模式下不进行网络请求，默认值为 false
  static OFFLINE = process.env.MTRAN_OFFLINE?.toLowerCase() === "true" || false;
  
  // 工作线程数，即每个语言模型的线程数，使用默认值 1 即可满足绝大多数场景
  static WORKERS = parseInt(process.env.MTRAN_WORKERS, 10) || 1;
  
  // 日志级别，可选值：Error、Warn、Info、Debug
  static LOG_LEVEL = process.env.MTRAN_LOG_LEVEL || "Error";
  
  // 数据目录，默认值为 ~/.cache/mtran
  static DATA_DIR =
    process.env.MTRAN_DATA_DIR || path.join(os.homedir(), ".cache", "mtran");
  
  // 内部使用，用于获取语言模型列表的 URL
  static MODELS_JSON_URL =
    process.env.MTRAN_MODELS_JSON_URL ||
    "https://firefox.settings.services.mozilla.com/v1/buckets/main-preview/collections/translations-models/records";
  
  // 内部使用，用于获取语言模型的 CDN 地址
  static MODELS_BASE_URL =
    process.env.MTRAN_MODELS_BASE_URL ||
    "https://firefox-settings-attachments.cdn.mozilla.net";
  
  // 内部使用，用于下载器设置用户代理
  static DOWNLOADER_UA =
    process.env.MTRAN_DOWNLOADER_UA ||
    "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:142.0) Gecko/20100101 Firefox/142.0a1";

  // === 内存管理配置 ===
  
  // 内存管理策略：conservative（保守）、balanced（平衡）、aggressive（积极）
  static MEMORY_STRATEGY = process.env.MTRAN_MEMORY_STRATEGY || "balanced";
  
  // 模型自动释放时间间隔（分钟），0表示禁用自动释放
  static MODEL_IDLE_TIMEOUT = parseFloat(process.env.MTRAN_MODEL_IDLE_TIMEOUT) || 30.0;
  
  // 获取内存配置（根据MEMORY_STRATEGY自动设置）
  static getMemoryConfig() {
    const strategies = {
      conservative: {
        modelIdleTimeout: 60.0,        // 1小时后释放模型
        memoryCheckInterval: 300000,   // 5分钟检查一次
        wasmCleanupInterval: 10000,    // 每10000次翻译清理
        wasmCleanupTimeout: 60.0,      // 1小时清理一次WASM
        timeoutResetThreshold: 600000, // 10分钟重置阈值
      },
      balanced: {
        modelIdleTimeout: 30.0,        // 30分钟后释放模型
        memoryCheckInterval: 60000,    // 1分钟检查一次
        wasmCleanupInterval: 5000,     // 每5000次翻译清理
        wasmCleanupTimeout: 30.0,      // 30分钟清理一次WASM
        timeoutResetThreshold: 300000, // 5分钟重置阈值
      },
      aggressive: {
        modelIdleTimeout: 10.0,        // 10分钟后释放模型
        memoryCheckInterval: 30000,    // 30秒检查一次
        wasmCleanupInterval: 1000,     // 每1000次翻译清理
        wasmCleanupTimeout: 10.0,      // 10分钟清理一次WASM
        timeoutResetThreshold: 120000, // 2分钟重置阈值
      }
    };
    
    const strategy = strategies[this.MEMORY_STRATEGY] || strategies.balanced;
    
    return {
      modelIdleTimeout: this.MODEL_IDLE_TIMEOUT > 0 ? this.MODEL_IDLE_TIMEOUT : strategy.modelIdleTimeout,
      memoryCheckInterval: strategy.memoryCheckInterval,
      wasmCleanupInterval: strategy.wasmCleanupInterval,
      wasmCleanupTimeout: strategy.wasmCleanupTimeout,
      timeoutResetThreshold: strategy.timeoutResetThreshold,
    };
  }

  // === Worker 配置 ===
  
  // Worker 初始化超时时间（毫秒），默认为 600000 毫秒（10分钟）
  static WORKER_INIT_TIMEOUT =
    parseInt(process.env.MTRAN_WORKER_INIT_TIMEOUT, 10) || 600000;

  // === 语言检测配置 ===
  
  // 语言检测最大文本长度，默认为 64 字符
  static MAX_DETECTION_LENGTH =
    parseInt(process.env.MTRAN_MAX_DETECTION_LENGTH, 10) || 64;

  // === 性能配置 ===
  
  // 是否启用性能日志，默认为 false
  static ENABLE_PERFORMANCE_LOGS =
    process.env.MTRAN_ENABLE_PERFORMANCE_LOGS?.toLowerCase() === "true" || false;
  
  // 翻译请求超时时间（毫秒），默认为 30000 毫秒（30秒）
  static TRANSLATION_TIMEOUT =
    parseInt(process.env.MTRAN_TRANSLATION_TIMEOUT, 10) || 30000;
  
  // 批量翻译最大并发数，默认为 10
  static MAX_CONCURRENT_TRANSLATIONS =
    parseInt(process.env.MTRAN_MAX_CONCURRENT_TRANSLATIONS, 10) || 10;

  // === 缓存配置 ===
  
  // 是否启用翻译结果缓存，默认为 false
  static ENABLE_TRANSLATION_CACHE =
    process.env.MTRAN_ENABLE_TRANSLATION_CACHE?.toLowerCase() === "true" || false;
  
  // 翻译缓存最大条目数，默认为 1000
  static TRANSLATION_CACHE_SIZE =
    parseInt(process.env.MTRAN_TRANSLATION_CACHE_SIZE, 10) || 1000;
  
  // 翻译缓存过期时间（毫秒），默认为 3600000 毫秒（1小时）
  static TRANSLATION_CACHE_TTL =
    parseInt(process.env.MTRAN_TRANSLATION_CACHE_TTL, 10) || 3600000;

  // === 调试配置 ===
  
  // 是否启用详细的引擎统计信息，默认为 false
  static ENABLE_ENGINE_STATS =
    process.env.MTRAN_ENABLE_ENGINE_STATS?.toLowerCase() === "true" || false;
  
  // 统计信息输出间隔（毫秒），默认为 300000 毫秒（5分钟）
  static STATS_OUTPUT_INTERVAL =
    parseInt(process.env.MTRAN_STATS_OUTPUT_INTERVAL, 10) || 300000;

  // === 网络配置 ===
  
  // 下载超时时间（毫秒），默认为 300000 毫秒（5分钟）
  static DOWNLOAD_TIMEOUT =
    parseInt(process.env.MTRAN_DOWNLOAD_TIMEOUT, 10) || 300000;
  
  // 下载重试次数，默认为 3 次
  static DOWNLOAD_RETRY_COUNT =
    parseInt(process.env.MTRAN_DOWNLOAD_RETRY_COUNT, 10) || 3;
  
  // 下载重试延迟（毫秒），默认为 1000 毫秒（1秒）
  static DOWNLOAD_RETRY_DELAY =
    parseInt(process.env.MTRAN_DOWNLOAD_RETRY_DELAY, 10) || 1000;

  /**
   * 获取所有配置项的摘要信息
   * @returns {Object} 配置摘要
   */
  static getSummary() {
    const memoryConfig = this.getMemoryConfig();
    
    return {
      // 基础配置
      offline: this.OFFLINE,
      workers: this.WORKERS,
      logLevel: this.LOG_LEVEL,
      dataDir: this.DATA_DIR,
      
      // 内存管理
      memoryStrategy: this.MEMORY_STRATEGY,
      modelIdleTimeout: this.MODEL_IDLE_TIMEOUT,
      memoryConfig: memoryConfig,
      
      // Worker 配置
      workerInitTimeout: this.WORKER_INIT_TIMEOUT,
      
      // 语言检测
      maxDetectionLength: this.MAX_DETECTION_LENGTH,
      
      // 性能配置
      enablePerformanceLogs: this.ENABLE_PERFORMANCE_LOGS,
      translationTimeout: this.TRANSLATION_TIMEOUT,
      maxConcurrentTranslations: this.MAX_CONCURRENT_TRANSLATIONS,
      
      // 缓存配置
      enableTranslationCache: this.ENABLE_TRANSLATION_CACHE,
      translationCacheSize: this.TRANSLATION_CACHE_SIZE,
      translationCacheTTL: this.TRANSLATION_CACHE_TTL,
      
      // 调试配置
      enableEngineStats: this.ENABLE_ENGINE_STATS,
      statsOutputInterval: this.STATS_OUTPUT_INTERVAL,
      
      // 网络配置
      downloadTimeout: this.DOWNLOAD_TIMEOUT,
      downloadRetryCount: this.DOWNLOAD_RETRY_COUNT,
      downloadRetryDelay: this.DOWNLOAD_RETRY_DELAY,
    };
  }

  /**
   * 验证配置的合理性
   * @returns {Array<string>} 验证错误信息数组
   */
  static validate() {
    const errors = [];
    
    if (this.WORKERS < 1 || this.WORKERS > 32) {
      errors.push("MTRAN_WORKERS must be between 1 and 32");
    }
    
    const validStrategies = ["conservative", "balanced", "aggressive"];
    if (!validStrategies.includes(this.MEMORY_STRATEGY)) {
      errors.push(`MTRAN_MEMORY_STRATEGY must be one of: ${validStrategies.join(", ")}`);
    }
    
    if (this.MODEL_IDLE_TIMEOUT < 0) {
      errors.push("MTRAN_MODEL_IDLE_TIMEOUT must be greater than or equal to 0");
    }
    
    if (this.WORKER_INIT_TIMEOUT < 30000) {
      errors.push("MTRAN_WORKER_INIT_TIMEOUT must be at least 30000ms");
    }
    
    if (this.MAX_DETECTION_LENGTH < 10 || this.MAX_DETECTION_LENGTH > 1000) {
      errors.push("MTRAN_MAX_DETECTION_LENGTH must be between 10 and 1000");
    }
    
    if (this.TRANSLATION_TIMEOUT < 1000) {
      errors.push("MTRAN_TRANSLATION_TIMEOUT must be at least 1000ms");
    }
    
    if (this.MAX_CONCURRENT_TRANSLATIONS < 1 || this.MAX_CONCURRENT_TRANSLATIONS > 100) {
      errors.push("MTRAN_MAX_CONCURRENT_TRANSLATIONS must be between 1 and 100");
    }
    
    const validLogLevels = ["Error", "Warn", "Info", "Debug"];
    if (!validLogLevels.includes(this.LOG_LEVEL)) {
      errors.push(`MTRAN_LOG_LEVEL must be one of: ${validLogLevels.join(", ")}`);
    }
    
    return errors;
  }
}

module.exports = Config;
