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
    "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:142.0) Gecko/20100101 Firefox/142.0a1";
  // 是否启用模型内存自动释放功能，默认为 true
  static AUTO_RELEASE =
    process.env.MTRAN_AUTO_RELEASE?.toLowerCase() !== "false";
  // 模型内存自动释放的时间间隔（分钟），默认为 30 分钟
  static RELEASE_INTERVAL =
    parseFloat(process.env.MTRAN_RELEASE_INTERVAL) || 30.0;
}

module.exports = Config;
