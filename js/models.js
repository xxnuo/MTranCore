"use strict";

const fs = require("fs").promises;
const path = require("path");

const Config = require("./config");
const { MALL } = require("./lang");
const DownloaderClass = require("./downloader");
const Downloader = new DownloaderClass(); // 创建下载器实例

const CACHE_DIR = Config.DATA_DIR;
const MODELS_JSON_URL = Config.MODELS_JSON_URL;
const MODELS_BASE_URL = Config.MODELS_BASE_URL;

const MODELS_DIR = path.join(CACHE_DIR, "models");
const MODELS_JSON_PATH = path.join(CACHE_DIR, "models.json");
const MODELS_FLAGS_PATH = path.join(CACHE_DIR, "flags.json");

let MODELS_DATA = null;
let MODELS_FLAGS = {
  downloaded: [],
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
  MODELS_DATA = await loadModelJson();
  MODELS_FLAGS = await loadModelFlags();
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
      return JSON.parse(jsonContent);
    }
    return { downloaded: [] };
  } catch (error) {
    console.error(`Failed to parse flags.json: ${error.message}`);
    return { downloaded: [] };
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
  if (forceRedownload || !(await exists(MODELS_JSON_PATH))) {
    try {
      if (Config.OFFLINE) {
        throw new Error(
          "Offline mode is not supported for loading models.json"
        );
      }
      // 同步下载，确保等待下载完成
      await Downloader.download(MODELS_JSON_URL, MODELS_JSON_PATH);
    } catch (error) {
      console.error("Failed to redownload models.json:", error.message);
    }
  }

  try {
    const jsonContent = await fs.readFile(MODELS_JSON_PATH, "utf8");
    return JSON.parse(jsonContent);
  } catch (error) {
    throw new Error(`Failed to parse models.json: ${error.message}`);
  }
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

  const matchingItems = MODELS_DATA.data.filter(
    (item) => item.fromLang === fromLang && item.toLang === toLang
  );

  // console.log(`Debug - Found ${matchingItems.length} matching items`);

  matchingItems.forEach((item) => {
    const fileType = item.fileType;
    if (fileType && item.attachment && item.attachment.location) {
      modelFiles[fileType] = item;
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

  if (forceUpdate) {
    MODELS_DATA = await loadModelJson(true);
  }

  if (!MODELS_DATA || !MODELS_DATA.data) {
    throw new Error("Invalid models data");
  }

  try {
    // 使用下划线作为分隔符，避免与语言代码中的破折号混淆
    const langPair = `${fromLang}_${toLang}`;

    // 如果已下载且不需要强制更新，直接从本地获取
    if (MODELS_FLAGS.downloaded.includes(langPair) && !forceUpdate) {
      // console.log(`Using cached model for ${fromLang}_${toLang}`);
      const modelFiles = getModelInfo(fromLang, toLang);
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

    const modelFiles = getModelInfo(fromLang, toLang);
    if (!Object.values(modelFiles).some((file) => file !== null)) {
      throw new Error(
        `No model files found for language pair: ${fromLang}_${toLang}`
      );
    }

    const downloadTasks = [];
    const downloadedFiles = {};
    let allFilesExist = true; // 标记所有文件是否都已存在

    for (const [fileType, fileInfo] of Object.entries(modelFiles)) {
      if (fileInfo && fileInfo.attachment && fileInfo.attachment.location) {
        const fileName = fileInfo.name;
        const filePath = path.join(MODELS_DIR, fileName);
        const expectedHash = fileInfo.attachment.hash;

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

        const needsDownload = forceUpdate || !(await isValid());

        if (needsDownload) {
          // 检查是否处于离线模式
          if (Config.OFFLINE) {
            throw new Error(
              `Cannot download model in offline mode: ${langPair}`
            );
          }

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
          });
        } else {
          downloadedFiles[fileType] = filePath;
        }
      }
    }

    // 如果所有文件都已存在且有效，但没有在flags中标记，则添加到flags中
    if (allFilesExist && !MODELS_FLAGS.downloaded.includes(langPair)) {
      // console.log(
      //   `All model files for ${fromLang}_${toLang} already exist, marking as downloaded`
      // );
      MODELS_FLAGS.downloaded.push(langPair);
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
          downloadedFiles[fileType] = downloadTasks[index].destination;
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
  return payload;
}

/**
 * 获取所有已下载的语言对
 */
async function getDownloadedModels() {
  return [...MODELS_FLAGS.downloaded];
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
};
