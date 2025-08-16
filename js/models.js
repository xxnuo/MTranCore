"use strict";

const fs = require("fs").promises;
const path = require("path");

const Config = require("./config");
const { MALL } = require("./lang");
const DownloaderClass = require("./downloader");
const Downloader = new DownloaderClass(); // 创建下载器实例
const { gc } = require("./utils");

const CACHE_DIR = Config.DATA_DIR;
const MODELS_JSON_URL = Config.MODELS_JSON_URL;
const MODELS_BASE_URL = Config.MODELS_BASE_URL;

const MODELS_DIR = path.join(CACHE_DIR, "models");
const MODELS_JSON_PATH = path.join(CACHE_DIR, "models.json");
const MODELS_FLAGS_PATH = path.join(CACHE_DIR, "flags.json");

let MODELS_DATA = null;
let MODELS_FLAGS = {
  downloaded: [],
  versions: {}, // 存储已下载模型的版本信息
};

// 文件是否存在
async function exists(path) {
  try {
    await fs.access(path);
    return true;
  } catch {
    return false;
  }
}

// 确保目录存在
async function ensureDir(dir) {
  try {
    if (await exists(dir)) {
      return true;
    }
    await fs.mkdir(dir, { recursive: true });
    return false;
  } catch (error) {
    console.error(`Error creating directory ${dir}: ${error.message}`);
    return false;
  }
}

// 初始化
async function init() {
  await ensureDir(CACHE_DIR);
  await ensureDir(MODELS_DIR);

  try {
    // 离线模式下只尝试加载本地文件，不进行更新
    if (Config.OFFLINE) {
      console.log("Running in offline mode. Model updates disabled.");
      if (await exists(MODELS_JSON_PATH)) {
        MODELS_DATA = await loadModelJson(false);
      } else {
        console.warn(
          "No local models.json found in offline mode. Some features may be limited."
        );
        MODELS_DATA = { data: [] };
      }
    } else {
      MODELS_DATA = await loadModelJson(!Config.OFFLINE);
    }

    MODELS_FLAGS = await loadModelFlags();
  } catch (error) {
    console.error(`Error initializing models: ${error.message}`);
    // 确保即使出错也有基本结构
    MODELS_DATA = MODELS_DATA || { data: [] };
    MODELS_FLAGS = MODELS_FLAGS || { downloaded: [], versions: {} };
  }
  gc();
}

// 保存模型标志到文件
async function saveModelFlags() {
  try {
    await fs.writeFile(
      MODELS_FLAGS_PATH,
      JSON.stringify(MODELS_FLAGS, null, 2)
    );
  } catch (error) {
    console.error(`Failed to save flags.json: ${error.message}`);
  }
}

// 加载模型标志
async function loadModelFlags() {
  try {
    if (await exists(MODELS_FLAGS_PATH)) {
      const jsonContent = await fs.readFile(MODELS_FLAGS_PATH, "utf8");
      const flags = JSON.parse(jsonContent);

      // 兼容旧版本的flags结构
      if (!flags.versions) {
        flags.versions = {};
      }

      return flags;
    }
    return { downloaded: [], versions: {} };
  } catch (error) {
    console.error(`Failed to parse flags.json: ${error.message}`);
    return { downloaded: [], versions: {} };
  }
}

// 列出所有可用的语言对
function getAllModels() {
  if (!MODELS_DATA || !MODELS_DATA.data) {
    console.log("No model data available");
    return [];
  }

  // 收集所有唯一的语言对
  const languagePairs = new Set();
  MODELS_DATA.data.forEach((item) => {
    if (item.fromLang && item.toLang) {
      languagePairs.add(`${item.fromLang}_${item.toLang}`);
    }
  });

  // 获取第一个模型项的结构
  // const sampleItem = MODELS_DATA.data.length > 0 ? MODELS_DATA.data[0] : null;
  // console.log(
  //   "Sample model item structure:",
  //   JSON.stringify(sampleItem, null, 2)
  // );

  return Array.from(languagePairs);
}

// 重新下载模型索引并自动重载模型索引数据
async function loadModelJson(forceRedownload = false) {
  // 在离线模式下，不允许强制重新下载
  if (Config.OFFLINE) {
    forceRedownload = false;
  }

  if (forceRedownload || !(await exists(MODELS_JSON_PATH))) {
    try {
      // 同步下载，确保等待下载完成
      if (!Config.OFFLINE || forceRedownload) {
        console.log(`Downloading models.json`);
        await Downloader.download(MODELS_JSON_URL, MODELS_JSON_PATH);
      } else if (!(await exists(MODELS_JSON_PATH))) {
        // 离线模式且文件不存在时返回空数据
        return { data: [] };
      }
    } catch (error) {
      console.error("Failed to redownload models.json:", error.message);
      // 如果下载失败但本地有缓存，使用本地缓存
      if (await exists(MODELS_JSON_PATH)) {
        console.log("Using cached models.json");
      } else {
        return { data: [] };
      }
    }
  }

  try {
    const jsonContent = await fs.readFile(MODELS_JSON_PATH, "utf8");
    return JSON.parse(jsonContent);
  } catch (error) {
    throw new Error(`Failed to parse models.json: ${error.message}`);
  }
}

// 版本比较函数，用于比较模型版本
function compareVersions(v1, v2) {
  // 处理未定义的版本
  if (!v1) return -1;
  if (!v2) return 1;

  // 标准版本格式处理
  const v1Parts = v1
    .replace(/[a-zA-Z]/g, ".")
    .split(".")
    .map((p) => parseInt(p) || 0);
  const v2Parts = v2
    .replace(/[a-zA-Z]/g, ".")
    .split(".")
    .map((p) => parseInt(p) || 0);

  // 比较版本号的每个部分
  for (let i = 0; i < Math.max(v1Parts.length, v2Parts.length); i++) {
    const part1 = v1Parts[i] || 0;
    const part2 = v2Parts[i] || 0;
    if (part1 !== part2) {
      return part2 - part1; // 降序排列，使得最新版本在前
    }
  }

  return 0;
}

// 根据语言对获取模型信息
function getModelInfo(fromLang, toLang) {
  if (!MODELS_DATA || !MODELS_DATA.data) {
    throw new Error("Invalid models data");
  }

  // // 添加调试信息
  // console.log(
  //   `Debug - MODELS_DATA: ${JSON.stringify(MODELS_DATA.data.length)} items`
  // );
  // console.log(`Debug - Looking for model: ${fromLang}_${toLang}`);

  const modelFiles = {
    model: null,
    lex: null,
    vocab: null,
    srcvocab: null,
    trgvocab: null,
  };

  // 获取匹配的项目并按版本排序
  const matchingItems = MODELS_DATA.data.filter(
    (item) => item.fromLang === fromLang && item.toLang === toLang
  );

  // 按文件类型分组并按版本排序
  const groupedByType = {};
  matchingItems.forEach((item) => {
    const fileType = item.fileType;
    if (fileType && item.attachment && item.attachment.location) {
      if (!groupedByType[fileType]) {
        groupedByType[fileType] = [];
      }
      groupedByType[fileType].push(item);
    }
  });

  // 对每种文件类型选择最新版本
  Object.keys(groupedByType).forEach((fileType) => {
    if (groupedByType[fileType].length > 0) {
      // 按版本降序排序，最新版本在前
      groupedByType[fileType].sort((a, b) =>
        compareVersions(a.version, b.version)
      );
      // 使用最新版本
      modelFiles[fileType] = groupedByType[fileType][0];

      // if (groupedByType[fileType].length > 1) {
      //   console.log(
      //     `Multiple versions found for ${fromLang}_${toLang} ${fileType}, using version ${modelFiles[fileType].version}`
      //   );
      // }
    }
  });

  return modelFiles;
}

/**
 * 获取指定语言对的模型文件，如果本地没有则下载
 */
async function getModelPaths(fromLang, toLang, forceUpdate = false) {
  if (!MALL.includes(fromLang)) {
    throw new Error(`Invalid fromLang: ${fromLang}`);
  }
  if (!MALL.includes(toLang)) {
    throw new Error(`Invalid toLang: ${toLang}`);
  }

  // 离线模式下不允许强制更新
  if (Config.OFFLINE) {
    forceUpdate = false;
  }

  if (forceUpdate && !Config.OFFLINE) {
    MODELS_DATA = await loadModelJson(true);
  }

  if (!MODELS_DATA || !MODELS_DATA.data) {
    throw new Error("Invalid models data");
  }

  try {
    // 使用下划线作为分隔符，避免与语言代码中的破折号混淆
    const langPair = `${fromLang}_${toLang}`;

    // 获取最新的模型信息
    const modelFiles = getModelInfo(fromLang, toLang);

    // 检查是否有可用模型
    if (!Object.values(modelFiles).some((file) => file !== null)) {
      throw new Error(
        `No model files found for language pair: ${fromLang}_${toLang}`
      );
    }

    // 检查是否有版本更新
    let hasNewerVersion = false;
    let currentVersions = {};

    // 收集最新模型的版本信息
    Object.entries(modelFiles).forEach(([fileType, fileInfo]) => {
      if (fileInfo && fileInfo.version) {
        currentVersions[fileType] = fileInfo.version;

        // 检查是否有更新的版本，仅在非离线模式下进行检查
        if (!Config.OFFLINE) {
          const storedVersion = MODELS_FLAGS.versions[langPair]?.[fileType];
          if (
            storedVersion &&
            compareVersions(fileInfo.version, storedVersion) > 0
          ) {
            hasNewerVersion = true;
            console.log(
              `Newer version available for ${langPair} ${fileType}: ${fileInfo.version} (current: ${storedVersion})`
            );
          }
        }
      }
    });

    // 如果已下载且不需要强制更新且没有新版本，直接从本地获取
    if (
      MODELS_FLAGS.downloaded.includes(langPair) &&
      !forceUpdate &&
      (!hasNewerVersion || Config.OFFLINE)
    ) {
      // console.log(`Using cached model for ${fromLang}_${toLang}`);
      const downloadedFiles = {};

      for (const [fileType, fileInfo] of Object.entries(modelFiles)) {
        if (fileInfo && fileInfo.name) {
          const filePath = path.join(MODELS_DIR, fileInfo.name);
          if (await exists(filePath)) {
            downloadedFiles[fileType] = filePath;
          }
        }
      }

      if (Object.keys(downloadedFiles).length > 0) {
        return downloadedFiles;
      }
    }

    const downloadTasks = [];
    const downloadedFiles = {};
    let allFilesExist = true; // 标记所有文件是否都已存在

    for (const [fileType, fileInfo] of Object.entries(modelFiles)) {
      if (fileInfo && fileInfo.attachment && fileInfo.attachment.location) {
        const fileName = fileInfo.name;
        const filePath = path.join(MODELS_DIR, fileName);
        const expectedHash = fileInfo.attachment.hash;

        // 检查是否需要更新版本，离线模式下不检查版本更新
        const storedVersion = MODELS_FLAGS.versions[langPair]?.[fileType];
        const needsVersionUpdate =
          !Config.OFFLINE &&
          fileInfo.version &&
          (!storedVersion ||
            compareVersions(fileInfo.version, storedVersion) > 0);

        const isValid = async () => {
          if (!(await exists(filePath))) {
            return false;
          }
          try {
            return await Downloader.verifyChecksum(
              filePath,
              expectedHash,
              "sha256"
            );
          } catch (error) {
            return false;
          }
        };

        const needsDownload =
          (forceUpdate || needsVersionUpdate || !(await isValid())) &&
          !Config.OFFLINE;

        if (needsDownload) {
          console.log(
            `Updating model: ${langPair} (${fileType} to version ${
              fileInfo.version || "unknown"
            })`
          );

          allFilesExist = false; // 有文件需要下载，标记为不是所有文件都存在
          const downloadUrl = `${MODELS_BASE_URL}/${fileInfo.attachment.location}`;
          downloadTasks.push({
            url: downloadUrl,
            destination: filePath,
            options: {
              checksum: expectedHash,
              algorithm: "sha256",
            },
            fileType: fileType,
            version: fileInfo.version,
          });
        } else {
          // 检查文件是否存在
          if (await exists(filePath)) {
            downloadedFiles[fileType] = filePath;
          } else if (Config.OFFLINE) {
            console.warn(`File ${fileName} not found in offline mode.`);
          }
        }
      }
    }

    // 如果所有文件都已存在且有效，但没有在flags中标记，则添加到flags中
    if (allFilesExist && !MODELS_FLAGS.downloaded.includes(langPair)) {
      // console.log(
      //   `All model files for ${fromLang}_${toLang} already exist, marking as downloaded`
      // );
      MODELS_FLAGS.downloaded.push(langPair);

      // 更新版本信息
      if (!MODELS_FLAGS.versions[langPair]) {
        MODELS_FLAGS.versions[langPair] = {};
      }

      // 合并当前版本信息
      MODELS_FLAGS.versions[langPair] = {
        ...MODELS_FLAGS.versions[langPair],
        ...currentVersions,
      };

      await saveModelFlags();
    }

    if (downloadTasks.length > 0) {
      const downloadResults = await Downloader.batchDownload(downloadTasks);

      // 检查是否所有文件都下载成功
      let allDownloadsSuccessful = true;

      // 处理下载结果
      downloadResults.forEach((result, index) => {
        if (result.status === "success") {
          const fileType = downloadTasks[index].fileType;
          const version = downloadTasks[index].version;
          downloadedFiles[fileType] = downloadTasks[index].destination;

          // 更新版本信息
          if (version) {
            if (!MODELS_FLAGS.versions[langPair]) {
              MODELS_FLAGS.versions[langPair] = {};
            }
            MODELS_FLAGS.versions[langPair][fileType] = version;
          }
        } else {
          allDownloadsSuccessful = false;
          console.error(
            `Failed to download ${downloadTasks[index].fileType} file: ${result.error}`
          );
        }
      });

      // 只有当所有文件都下载成功时，才更新语言对记录
      if (
        allDownloadsSuccessful &&
        !MODELS_FLAGS.downloaded.includes(langPair)
      ) {
        MODELS_FLAGS.downloaded.push(langPair);
        await saveModelFlags();
      } else if (allDownloadsSuccessful) {
        // 如果只是更新了版本，也需要保存标志
        await saveModelFlags();
      }
    }

    // 如果没有找到任何文件，抛出错误
    if (Object.keys(downloadedFiles).length === 0) {
      throw new Error(`No valid model files available for ${langPair}`);
    }

    return downloadedFiles;
  } catch (error) {
    console.error(
      `Error getting models for ${fromLang}_${toLang}: ${error.message}`
    );
    throw error;
  }
}

async function getModel(fromLang, toLang, forceUpdate = false) {
  const modelFiles = await getModelPaths(fromLang, toLang, forceUpdate);
  const languageModelFiles = {};
  for (const [fileType, filePath] of Object.entries(modelFiles)) {
    const modelRecord = {
      name: path.basename(filePath),
      fileType: fileType,
      fromLang: fromLang,
      toLang: toLang,
    };
    languageModelFiles[fileType] = {
      buffer: await fs.readFile(filePath),
      record: modelRecord,
    };
  }
  const payload = {
    sourceLanguage: fromLang,
    targetLanguage: toLang,
    languageModelFiles: languageModelFiles,
  };

  gc();
  return payload;
}

/**
 * 检查指定语言对的模型是否有更新
 * @param {string} fromLang 源语言
 * @param {string} toLang 目标语言
 * @returns {Promise<{hasUpdate: boolean, versions: Object}>} 更新状态和版本信息
 */
async function checkModelUpdates(fromLang, toLang) {
  if (!MALL.includes(fromLang)) {
    throw new Error(`Invalid fromLang: ${fromLang}`);
  }
  if (!MALL.includes(toLang)) {
    throw new Error(`Invalid toLang: ${toLang}`);
  }

  // 离线模式下不检查更新
  if (Config.OFFLINE) {
    return { hasUpdate: false, versions: {}, offline: true };
  }

  try {
    // 强制重新下载模型索引
    MODELS_DATA = await loadModelJson(true);

    const langPair = `${fromLang}_${toLang}`;
    const modelFiles = getModelInfo(fromLang, toLang);

    // 检查是否有可用模型
    if (!Object.values(modelFiles).some((file) => file !== null)) {
      return { hasUpdate: false, versions: {} };
    }

    // 检查是否有版本更新
    let hasNewerVersion = false;
    const currentVersions = {};
    const newVersions = {};

    // 收集最新模型的版本信息
    Object.entries(modelFiles).forEach(([fileType, fileInfo]) => {
      if (fileInfo && fileInfo.version) {
        newVersions[fileType] = fileInfo.version;

        // 检查是否有更新的版本
        const storedVersion = MODELS_FLAGS.versions[langPair]?.[fileType];
        currentVersions[fileType] = storedVersion || "none";

        if (
          !storedVersion ||
          compareVersions(fileInfo.version, storedVersion) > 0
        ) {
          hasNewerVersion = true;
        }
      }
    });

    return {
      hasUpdate: hasNewerVersion,
      versions: {
        current: currentVersions,
        latest: newVersions,
      },
    };
  } catch (error) {
    console.error(
      `Error checking updates for ${fromLang}_${toLang}: ${error.message}`
    );
    return { hasUpdate: false, versions: {} };
  }
}

/**
 * 获取所有已下载的语言对
 */
async function getDownloadedModels() {
  return [...MODELS_FLAGS.downloaded];
}

/**
 * 获取已下载模型的版本信息
 */
function getModelVersions() {
  return { ...MODELS_FLAGS.versions };
}

module.exports = {
  MODELS_DIR,
  MODELS_JSON_PATH,
  MODELS_FLAGS_PATH,
  MODELS_DATA,
  MODELS_FLAGS,
  init,
  getModel,
  getDownloadedModels,
  getAllModels,
  checkModelUpdates,
  getModelVersions,
};
