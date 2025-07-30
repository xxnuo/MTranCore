"use strict";

const fs = require("fs").promises;
const path = require("path");

const Config = require("./config");
const { MALL } = require("./lang");
const DownloaderClass = require("./downloader");
const Logger = require("./logger");

// 模块名称常量
const MODULE_NAME = "Models";

// 单例下载器实例
const downloader = new DownloaderClass();

// 常量定义
const CACHE_DIR = Config.DATA_DIR;
const MODELS_JSON_URL = Config.MODELS_JSON_URL;
const MODELS_BASE_URL = Config.MODELS_BASE_URL;
const MODELS_DIR = path.join(CACHE_DIR, "models");
const MODELS_JSON_PATH = path.join(CACHE_DIR, "models.json");
const MODELS_FLAGS_PATH = path.join(CACHE_DIR, "flags.json");

// 全局状态
let modelsData = null;
let modelsFlags = {
  downloaded: [],
};

/**
 * 检查文件是否存在
 * @param {string} filePath - 文件路径
 * @returns {Promise<boolean>} 文件是否存在
 */
async function fileExists(filePath) {
  try {
    await fs.access(filePath);
    return true;
  } catch {
    return false;
  }
}

/**
 * 确保目录存在，如果不存在则创建
 * @param {string} dirPath - 目录路径
 * @returns {Promise<boolean>} 目录是否已存在（true表示已存在，false表示新创建）
 */
async function ensureDirectory(dirPath) {
  try {
    if (await fileExists(dirPath)) {
      return true;
    }
    await fs.mkdir(dirPath, { recursive: true });
    Logger.debug(MODULE_NAME, `创建目录: ${dirPath}`);
    return false;
  } catch (error) {
    Logger.error(MODULE_NAME, `创建目录失败 ${dirPath}:`, error.message);
    throw error;
  }
}

/**
 * 加载模型标志文件
 * @returns {Promise<Object>} 模型标志对象
 */
async function loadModelFlags() {
  try {
    if (await fileExists(MODELS_FLAGS_PATH)) {
      Logger.debug(MODULE_NAME, "加载标志文件:", MODELS_FLAGS_PATH);
      const jsonContent = await fs.readFile(MODELS_FLAGS_PATH, "utf8");
      const flags = JSON.parse(jsonContent);

      // 验证数据格式
      if (!flags.downloaded || !Array.isArray(flags.downloaded)) {
        Logger.warn(MODULE_NAME, "标志文件格式无效，使用默认值");
        return { downloaded: [] };
      }

      Logger.debug(
        MODULE_NAME,
        `已加载 ${flags.downloaded.length} 个已下载标志`
      );
      return flags;
    }
    Logger.debug(MODULE_NAME, "标志文件不存在，使用默认值");
    return { downloaded: [] };
  } catch (error) {
    Logger.error(MODULE_NAME, "解析标志文件失败:", error.message);
    return { downloaded: [] };
  }
}

/**
 * 保存模型标志到文件
 * @returns {Promise<void>}
 */
async function saveModelFlags() {
  try {
    await fs.writeFile(
      MODELS_FLAGS_PATH,
      JSON.stringify(modelsFlags, null, 2),
      "utf8"
    );
    Logger.debug(MODULE_NAME, "标志文件保存成功:", MODELS_FLAGS_PATH);
  } catch (error) {
    Logger.error(MODULE_NAME, "保存标志文件失败:", error.message);
    throw error;
  }
}

/**
 * 加载或下载模型索引文件
 * @param {boolean} forceRedownload - 是否强制重新下载
 * @returns {Promise<Object>} 模型数据对象
 */
async function loadModelIndex(forceRedownload = false) {
  Logger.timeStart(MODULE_NAME, "loadModelIndex");

  // 如果需要强制重新下载或文件不存在，则下载
  if (forceRedownload || !(await fileExists(MODELS_JSON_PATH))) {
    if (Config.OFFLINE) {
      const error = new Error("离线模式不支持下载模型索引文件");
      Logger.error(MODULE_NAME, error.message);
      throw error;
    }

    try {
      Logger.info(MODULE_NAME, "正在下载模型索引文件...");
      await downloader.download(MODELS_JSON_URL, MODELS_JSON_PATH);
      Logger.info(MODULE_NAME, "模型索引文件下载完成");
    } catch (error) {
      Logger.error(MODULE_NAME, "下载模型索引文件失败:", error.message);
      throw error;
    }
  }

  try {
    Logger.debug(MODULE_NAME, "解析模型索引文件:", MODELS_JSON_PATH);
    const jsonContent = await fs.readFile(MODELS_JSON_PATH, "utf8");
    const data = JSON.parse(jsonContent);

    // 验证数据格式
    if (!data.data || !Array.isArray(data.data)) {
      throw new Error("模型索引文件格式无效");
    }

    Logger.debug(
      MODULE_NAME,
      `模型索引加载完成，包含 ${data.data.length} 个模型项`
    );
    Logger.timeEnd(MODULE_NAME, "loadModelIndex");
    return data;
  } catch (error) {
    Logger.error(MODULE_NAME, "解析模型索引文件失败:", error.message);
    Logger.timeEnd(MODULE_NAME, "loadModelIndex");
    throw new Error(`解析模型索引文件失败: ${error.message}`);
  }
}

/**
 * 验证语言代码
 * @param {string} langCode - 语言代码
 * @param {string} paramName - 参数名称（用于错误信息）
 * @throws {Error} 当语言代码无效时抛出错误
 */
function validateLanguageCode(langCode, paramName = "语言代码") {
  if (!langCode || typeof langCode !== "string") {
    const error = new Error(`${paramName}不能为空且必须是字符串`);
    Logger.error(MODULE_NAME, error.message);
    throw error;
  }

  if (!MALL.includes(langCode)) {
    const error = new Error(`不支持的${paramName}: ${langCode}`);
    Logger.error(MODULE_NAME, error.message);
    throw error;
  }

  Logger.debug(MODULE_NAME, `语言代码验证通过: ${langCode}`);
}

/**
 * 获取指定语言对的模型文件信息
 * @param {string} fromLang - 源语言代码
 * @param {string} toLang - 目标语言代码
 * @returns {Object} 模型文件信息对象
 */
function getModelFileInfo(fromLang, toLang) {
  if (!modelsData?.data) {
    const error = new Error("模型数据无效或未初始化");
    Logger.error(MODULE_NAME, error.message);
    throw error;
  }

  Logger.debug(MODULE_NAME, `查找模型文件信息: ${fromLang} -> ${toLang}`);

  const modelFiles = {
    model: null,
    lex: null,
    vocab: null,
    srcvocab: null,
    trgvocab: null,
  };

  // 查找匹配的模型项
  const matchingItems = modelsData.data.filter(
    (item) => item.fromLang === fromLang && item.toLang === toLang
  );

  if (matchingItems.length === 0) {
    const error = new Error(`未找到语言对 ${fromLang}_${toLang} 的模型文件`);
    Logger.error(MODULE_NAME, error.message);
    throw error;
  }

  Logger.debug(MODULE_NAME, `找到 ${matchingItems.length} 个匹配的模型项`);

  // 收集模型文件信息
  matchingItems.forEach((item) => {
    const { fileType, attachment } = item;
    if (fileType && attachment?.location) {
      modelFiles[fileType] = item;
      Logger.debug(MODULE_NAME, `找到模型文件: ${fileType} - ${item.name}`);
    }
  });

  // 验证是否找到了必要的文件
  const hasValidFiles = Object.values(modelFiles).some((file) => file !== null);
  if (!hasValidFiles) {
    const error = new Error(`语言对 ${fromLang}_${toLang} 没有有效的模型文件`);
    Logger.error(MODULE_NAME, error.message);
    throw error;
  }

  const validFileCount = Object.values(modelFiles).filter(
    (file) => file !== null
  ).length;
  Logger.debug(
    MODULE_NAME,
    `模型文件信息获取完成，共 ${validFileCount} 个文件`
  );
  return modelFiles;
}

/**
 * 验证文件完整性
 * @param {string} filePath - 文件路径
 * @param {string} expectedHash - 预期哈希值
 * @returns {Promise<boolean>} 文件是否有效
 */
async function validateFileIntegrity(filePath, expectedHash) {
  if (!(await fileExists(filePath))) {
    Logger.debug(MODULE_NAME, `文件不存在: ${filePath}`);
    return false;
  }

  try {
    Logger.debug(MODULE_NAME, `验证文件完整性: ${path.basename(filePath)}`);
    const isValid = await downloader.verifyChecksum(
      filePath,
      expectedHash,
      "sha256"
    );
    Logger.debugIf(
      !isValid,
      MODULE_NAME,
      `文件完整性验证失败: ${path.basename(filePath)}`
    );
    return isValid;
  } catch (error) {
    Logger.warn(
      MODULE_NAME,
      `文件完整性验证失败 ${path.basename(filePath)}:`,
      error.message
    );
    return false;
  }
}

/**
 * 准备下载任务
 * @param {Object} modelFiles - 模型文件信息
 * @param {boolean} forceUpdate - 是否强制更新
 * @returns {Promise<{downloadTasks: Array, existingFiles: Object}>} 下载任务和已存在文件
 */
async function prepareDownloadTasks(modelFiles, forceUpdate = false) {
  Logger.timeStart(MODULE_NAME, "prepareDownloadTasks");

  const downloadTasks = [];
  const existingFiles = {};

  for (const [fileType, fileInfo] of Object.entries(modelFiles)) {
    if (!fileInfo?.attachment?.location) {
      continue;
    }

    const fileName = fileInfo.name;
    const filePath = path.join(MODELS_DIR, fileName);
    const expectedHash = fileInfo.attachment.hash;

    Logger.debug(MODULE_NAME, `检查文件: ${fileName}`);

    // 检查文件是否需要下载
    const needsDownload =
      forceUpdate || !(await validateFileIntegrity(filePath, expectedHash));

    if (needsDownload) {
      if (Config.OFFLINE) {
        const error = new Error(`离线模式无法下载模型文件: ${fileName}`);
        Logger.error(MODULE_NAME, error.message);
        throw error;
      }

      const downloadUrl = `${MODELS_BASE_URL}/${fileInfo.attachment.location}`;
      downloadTasks.push({
        url: downloadUrl,
        destination: filePath,
        options: {
          checksum: expectedHash,
          algorithm: "sha256",
        },
        fileType,
        fileName,
      });
      Logger.debug(MODULE_NAME, `添加下载任务: ${fileName}`);
    } else {
      existingFiles[fileType] = filePath;
      Logger.debug(MODULE_NAME, `使用现有文件: ${fileName}`);
    }
  }

  Logger.debug(
    MODULE_NAME,
    `下载任务准备完成: ${downloadTasks.length} 个待下载，${
      Object.keys(existingFiles).length
    } 个已存在`
  );
  Logger.timeEnd(MODULE_NAME, "prepareDownloadTasks");
  return { downloadTasks, existingFiles };
}

/**
 * 处理下载结果
 * @param {Array} downloadResults - 下载结果数组
 * @param {Array} downloadTasks - 下载任务数组
 * @returns {Object} 成功下载的文件路径对象
 */
function processDownloadResults(downloadResults, downloadTasks) {
  const downloadedFiles = {};
  let allSuccessful = true;

  downloadResults.forEach((result, index) => {
    const task = downloadTasks[index];

    if (result.status === "success") {
      downloadedFiles[task.fileType] = task.destination;
      Logger.info(MODULE_NAME, `成功下载: ${task.fileName}`);
    } else {
      allSuccessful = false;
      Logger.error(MODULE_NAME, `下载失败 ${task.fileName}:`, result.error);
    }
  });

  Logger.debug(
    MODULE_NAME,
    `下载结果处理完成: ${Object.keys(downloadedFiles).length}/${
      downloadTasks.length
    } 成功`
  );
  return { downloadedFiles, allSuccessful };
}

/**
 * 更新下载标志
 * @param {string} langPair - 语言对标识
 * @param {boolean} success - 是否成功
 */
async function updateDownloadFlag(langPair, success) {
  if (success && !modelsFlags.downloaded.includes(langPair)) {
    modelsFlags.downloaded.push(langPair);
    await saveModelFlags();
    Logger.info(MODULE_NAME, `已标记语言对 ${langPair} 为已下载`);
  }
}

/**
 * 初始化模型管理器
 * @returns {Promise<void>}
 */
async function initialize() {
  Logger.timeStart(MODULE_NAME, "initialize");
  Logger.info(MODULE_NAME, "开始初始化模型管理器...");

  try {
    await ensureDirectory(CACHE_DIR);
    await ensureDirectory(MODELS_DIR);

    modelsData = await loadModelIndex();
    modelsFlags = await loadModelFlags();

    Logger.info(MODULE_NAME, "模型管理器初始化完成");
    Logger.timeEnd(MODULE_NAME, "initialize");
  } catch (error) {
    Logger.error(MODULE_NAME, "模型管理器初始化失败:", error.message);
    Logger.timeEnd(MODULE_NAME, "initialize");
    throw error;
  }
}

/**
 * 获取所有可用的语言对
 * @returns {Array<string>} 语言对数组
 */
function getAllAvailableModels() {
  if (!modelsData?.data) {
    Logger.warn(MODULE_NAME, "模型数据不可用");
    return [];
  }

  // 收集所有唯一的语言对
  const languagePairs = new Set();

  modelsData.data.forEach((item) => {
    if (item.fromLang && item.toLang) {
      languagePairs.add(`${item.fromLang}_${item.toLang}`);
    }
  });

  const result = Array.from(languagePairs).sort();
  Logger.debug(MODULE_NAME, `找到 ${result.length} 个可用的语言对`);
  return result;
}

/**
 * 获取指定语言对的模型文件路径，如果本地没有则下载
 * @param {string} fromLang - 源语言代码
 * @param {string} toLang - 目标语言代码
 * @param {boolean} forceUpdate - 是否强制更新
 * @returns {Promise<Object>} 模型文件路径对象
 */
async function getModelPaths(fromLang, toLang, forceUpdate = false) {
  Logger.timeStart(MODULE_NAME, `getModelPaths_${fromLang}_${toLang}`);
  Logger.info(
    MODULE_NAME,
    `获取模型文件路径: ${fromLang} -> ${toLang}${
      forceUpdate ? " (强制更新)" : ""
    }`
  );

  // 验证输入参数
  validateLanguageCode(fromLang, "源语言代码");
  validateLanguageCode(toLang, "目标语言代码");

  const langPair = `${fromLang}_${toLang}`;

  try {
    // 如果需要强制更新，重新加载模型索引
    if (forceUpdate) {
      Logger.debug(MODULE_NAME, "强制更新，重新加载模型索引");
      modelsData = await loadModelIndex(true);
    }

    // 检查缓存
    if (modelsFlags.downloaded.includes(langPair) && !forceUpdate) {
      Logger.debug(MODULE_NAME, `检查缓存文件: ${langPair}`);
      const modelFiles = getModelFileInfo(fromLang, toLang);
      const cachedFiles = {};

      // 验证缓存文件是否仍然有效
      for (const [fileType, fileInfo] of Object.entries(modelFiles)) {
        if (fileInfo?.name) {
          const filePath = path.join(MODELS_DIR, fileInfo.name);
          if (await fileExists(filePath)) {
            cachedFiles[fileType] = filePath;
          }
        }
      }

      if (Object.keys(cachedFiles).length > 0) {
        Logger.info(MODULE_NAME, `使用缓存的模型文件: ${langPair}`);
        Logger.timeEnd(MODULE_NAME, `getModelPaths_${fromLang}_${toLang}`);
        return cachedFiles;
      } else {
        Logger.warn(MODULE_NAME, `缓存文件不完整，重新下载: ${langPair}`);
      }
    }

    // 获取模型文件信息
    const modelFiles = getModelFileInfo(fromLang, toLang);

    // 准备下载任务
    const { downloadTasks, existingFiles } = await prepareDownloadTasks(
      modelFiles,
      forceUpdate
    );

    let allFiles = { ...existingFiles };

    // 执行下载任务
    if (downloadTasks.length > 0) {
      Logger.info(
        MODULE_NAME,
        `开始下载 ${downloadTasks.length} 个模型文件...`
      );

      const downloadResults = await downloader.batchDownload(downloadTasks);
      const { downloadedFiles, allSuccessful } = processDownloadResults(
        downloadResults,
        downloadTasks
      );

      allFiles = { ...allFiles, ...downloadedFiles };

      // 更新下载标志
      await updateDownloadFlag(
        langPair,
        allSuccessful &&
          Object.keys(downloadedFiles).length === downloadTasks.length
      );
    } else {
      // 所有文件都已存在，更新标志
      Logger.debug(MODULE_NAME, `所有文件都已存在: ${langPair}`);
      await updateDownloadFlag(langPair, true);
    }

    // 验证是否有可用文件
    if (Object.keys(allFiles).length === 0) {
      const error = new Error(`语言对 ${langPair} 没有可用的模型文件`);
      Logger.error(MODULE_NAME, error.message);
      throw error;
    }

    Logger.info(
      MODULE_NAME,
      `模型文件准备完成: ${langPair} (${Object.keys(allFiles).length} 个文件)`
    );
    Logger.timeEnd(MODULE_NAME, `getModelPaths_${fromLang}_${toLang}`);
    return allFiles;
  } catch (error) {
    Logger.error(MODULE_NAME, `获取模型文件失败 ${langPair}:`, error.message);
    Logger.timeEnd(MODULE_NAME, `getModelPaths_${fromLang}_${toLang}`);
    throw error;
  }
}

/**
 * 获取模型数据（包含文件内容）
 * @param {string} fromLang - 源语言代码
 * @param {string} toLang - 目标语言代码
 * @param {boolean} forceUpdate - 是否强制更新
 * @returns {Promise<Object>} 包含模型数据的负载对象
 */
async function getModelData(fromLang, toLang, forceUpdate = false) {
  Logger.timeStart(MODULE_NAME, `getModelData_${fromLang}_${toLang}`);
  Logger.info(MODULE_NAME, `获取模型数据: ${fromLang} -> ${toLang}`);

  const modelFilePaths = await getModelPaths(fromLang, toLang, forceUpdate);
  const languageModelFiles = {};

  // 读取所有模型文件内容
  for (const [fileType, filePath] of Object.entries(modelFilePaths)) {
    try {
      Logger.debug(MODULE_NAME, `读取模型文件: ${path.basename(filePath)}`);
      const buffer = await fs.readFile(filePath);
      const modelRecord = {
        name: path.basename(filePath),
        fileType,
        fromLang,
        toLang,
      };

      languageModelFiles[fileType] = {
        buffer,
        record: modelRecord,
      };
      Logger.debug(
        MODULE_NAME,
        `模型文件读取完成: ${path.basename(filePath)} (${buffer.length} bytes)`
      );
    } catch (error) {
      Logger.error(MODULE_NAME, `读取模型文件失败 ${filePath}:`, error.message);
      Logger.timeEnd(MODULE_NAME, `getModelData_${fromLang}_${toLang}`);
      throw error;
    }
  }

  const result = {
    sourceLanguage: fromLang,
    targetLanguage: toLang,
    languageModelFiles,
  };

  Logger.info(
    MODULE_NAME,
    `模型数据获取完成: ${fromLang} -> ${toLang} (${
      Object.keys(languageModelFiles).length
    } 个文件)`
  );
  Logger.timeEnd(MODULE_NAME, `getModelData_${fromLang}_${toLang}`);
  return result;
}

/**
 * 获取所有已下载的语言对
 * @returns {Array<string>} 已下载的语言对数组
 */
function getDownloadedModels() {
  const result = [...modelsFlags.downloaded].sort();
  Logger.debug(MODULE_NAME, `已下载的语言对: ${result.length} 个`);
  return result;
}

/**
 * 清理无效的下载标志
 * @returns {Promise<void>}
 */
async function cleanupInvalidFlags() {
  Logger.timeStart(MODULE_NAME, "cleanupInvalidFlags");
  Logger.info(MODULE_NAME, "开始清理无效的下载标志...");

  const validFlags = [];
  let cleanedCount = 0;

  for (const langPair of modelsFlags.downloaded) {
    const [fromLang, toLang] = langPair.split("_");

    try {
      // 验证语言对是否仍然有效
      validateLanguageCode(fromLang);
      validateLanguageCode(toLang);

      // 检查文件是否仍然存在
      const modelFiles = getModelFileInfo(fromLang, toLang);
      let hasValidFiles = false;

      for (const [, fileInfo] of Object.entries(modelFiles)) {
        if (fileInfo?.name) {
          const filePath = path.join(MODELS_DIR, fileInfo.name);
          if (await fileExists(filePath)) {
            hasValidFiles = true;
            break;
          }
        }
      }

      if (hasValidFiles) {
        validFlags.push(langPair);
      } else {
        Logger.info(MODULE_NAME, `清理无效标志: ${langPair} (文件不存在)`);
        cleanedCount++;
      }
    } catch (error) {
      Logger.info(MODULE_NAME, `清理无效标志: ${langPair} (${error.message})`);
      cleanedCount++;
    }
  }

  if (validFlags.length !== modelsFlags.downloaded.length) {
    modelsFlags.downloaded = validFlags;
    await saveModelFlags();
    Logger.info(MODULE_NAME, `已清理 ${cleanedCount} 个无效的下载标志`);
  } else {
    Logger.debug(MODULE_NAME, "没有发现无效的下载标志");
  }

  Logger.timeEnd(MODULE_NAME, "cleanupInvalidFlags");
}

// 导出的公共接口
module.exports = {
  // 常量
  MODELS_DIR,
  MODELS_JSON_PATH,
  MODELS_FLAGS_PATH,

  // 只读访问器
  get MODELS_DATA() {
    return modelsData;
  },
  get MODELS_FLAGS() {
    return modelsFlags;
  },

  // 主要功能函数
  init: initialize,
  getModel: getModelData,
  getModelPaths,
  getAllModels: getAllAvailableModels,
  getDownloadedModels,

  // 维护功能
  cleanupInvalidFlags,

  // 工具函数（供测试使用）
  fileExists,
  ensureDirectory,
  validateLanguageCode,
};
