// 正常导入
const path = require("path");
const { Worker } = require("worker_threads");
const Lang = require("./lang");
const Models = require("./models");
const OpenCC = require("./opencc");
const LanguageDetector = require("./ld");
const { MESSAGE_TYPES } = require("./message");
const Config = require("./config");
const Logger = require("./logger");

// 配置常量
const ENGINE_CACHE_ENABLE = Config.AUTO_RELEASE;
const ENGINE_CACHE_TIMEOUT_MINUTES = Config.RELEASE_INTERVAL;
const ENGINE_CACHE_TIMEOUT_MS =
  ENGINE_CACHE_ENABLE && ENGINE_CACHE_TIMEOUT_MINUTES > 0
    ? Math.round(ENGINE_CACHE_TIMEOUT_MINUTES * 60 * 1000)
    : Infinity;
const WORKERS_PER_LANGUAGE_PAIR = Config.WORKERS;
const MEMORY_CHECK_INTERVAL_MS = Config.MEMORY_CHECK_INTERVAL;
const TIMEOUT_RESET_THRESHOLD_MS = Config.TIMEOUT_RESET_THRESHOLD;
const WORKER_INIT_TIMEOUT_MS = Config.WORKER_INIT_TIMEOUT;
const MAX_DETECTION_LENGTH = Config.MAX_DETECTION_LENGTH;

const workerPath = path.join(__dirname, "worker.js");

/**
 * 翻译引擎管理器
 * 负责管理翻译引擎的生命周期、Worker池、内存释放等
 */
class Translator {
  // 私有静态属性
  static #cachedEngines = new Map();
  static #messageId = 0;
  static #pendingMessages = new Map();
  static #memoryReleaseTimer = null;
  static #models = null;

  /**
   * 获取支持的语言列表
   * @returns {Array} 支持的语言代码数组
   */
  static GetSupportLanguages() {
    return Lang.MALL;
  }

  /**
   * 生成语言对键名
   * @param {string} fromLang - 源语言
   * @param {string} toLang - 目标语言
   * @returns {string} 语言对键名
   */
  static getLanguagePairKey(fromLang, toLang) {
    return `${fromLang}_${toLang}`;
  }

  /**
   * 预加载翻译引擎
   * @param {string} fromLang - 源语言
   * @param {string} toLang - 目标语言
   * @returns {Object} 翻译引擎对象
   */
  static async Preload(fromLang, toLang) {
    const languagePairKey = this.getLanguagePairKey(fromLang, toLang);
    const cachedEngine = this.#cachedEngines.get(languagePairKey);

    if (cachedEngine) {
      Logger.debug("Translator", `使用缓存的翻译引擎: ${languagePairKey}`);
      this.keepAlive(languagePairKey);
      return this.createEngineInterface(languagePairKey);
    }

    try {
      Logger.info("Translator", `创建新的翻译引擎: ${languagePairKey}`);
      
      // 创建worker池
      const workerPool = await this.createWorkerPool(fromLang, toLang);

      // 添加到缓存
      const engineData = {
        workerPool,
        nextWorkerIndex: 0,
        timeoutId: null,
        lastUsedTime: Date.now(),
        useCount: 0,
      };

      // 设置超时计时器
      this.setEngineTimeout(engineData, languagePairKey);
      this.#cachedEngines.set(languagePairKey, engineData);

      // 启动内存释放定时器
      this.startMemoryReleaseTimer();

      Logger.info("Translator", `翻译引擎创建成功: ${languagePairKey}`);
      return this.createEngineInterface(languagePairKey);
    } catch (error) {
      Logger.error("Translator", `创建翻译引擎失败: ${languagePairKey}`, error.message);
      throw new Error(`Create translation engine failed: ${error.message}`);
    }
  }

  /**
   * 创建引擎接口对象
   * @param {string} languagePairKey - 语言对键名
   * @returns {Object} 引擎接口
   */
  static createEngineInterface(languagePairKey) {
    return {
      translate: (texts, isHTML) => this.translateWithWorker(languagePairKey, texts, isHTML),
      discardTranslations: () => this.discardTranslations(languagePairKey),
    };
  }

  /**
   * 设置引擎超时计时器
   * @param {Object} engineData - 引擎数据
   * @param {string} languagePairKey - 语言对键名
   */
  static setEngineTimeout(engineData, languagePairKey) {
    if (ENGINE_CACHE_ENABLE && ENGINE_CACHE_TIMEOUT_MINUTES > 0) {
      engineData.timeoutId = setTimeout(() => {
        Logger.info("Translator", `引擎超时自动释放: ${languagePairKey}`);
        this.removeEngine(languagePairKey);
      }, ENGINE_CACHE_TIMEOUT_MS);
    }
  }

  /**
   * 使用Worker进行翻译
   * @param {string} languagePairKey - 语言对键名
   * @param {string|Array} texts - 待翻译文本
   * @param {boolean} isHTML - 是否为HTML格式
   * @returns {string|Array} 翻译结果
   */
  static async translateWithWorker(languagePairKey, texts, isHTML) {
    const cachedEngine = this.#cachedEngines.get(languagePairKey);
    if (!cachedEngine) {
      throw new Error(`Translation engine not found: ${languagePairKey}`);
    }

    this.keepAlive(languagePairKey);
    this.updateEngineStats(cachedEngine);

    const { workerPool } = cachedEngine;
    const isTextArray = Array.isArray(texts);
    const sourceTexts = isTextArray ? texts : [texts];

    Logger.debug("Translator", `开始翻译任务，文本数量: ${sourceTexts.length}`);

    // 并行处理翻译任务
    const results = await Promise.all(
      sourceTexts.map((sourceText) => this.translateSingleText(sourceText, cachedEngine, isHTML))
    );

    Logger.debug("Translator", `翻译任务完成，结果数量: ${results.length}`);
    return isTextArray ? results : results[0];
  }

  /**
   * 翻译单个文本
   * @param {string} sourceText - 源文本
   * @param {Object} cachedEngine - 缓存的引擎
   * @param {boolean} isHTML - 是否为HTML格式
   * @returns {Promise<string>} 翻译结果
   */
  static async translateSingleText(sourceText, cachedEngine, isHTML) {
    if (!sourceText || !sourceText.trim()) {
      return "";
    }

    const messageId = this.#messageId++;
    const translationId = Date.now() + Math.random();
    
    // 轮询选择worker
    const worker = this.selectWorker(cachedEngine);

    return new Promise((resolve, reject) => {
      this.#pendingMessages.set(messageId, {
        resolve,
        reject,
        translationId,
      });

      worker.postMessage({
        type: MESSAGE_TYPES.TRANSLATION_REQUEST.type,
        sourceText,
        messageId,
        translationId,
        isHTML,
      });
    });
  }

  /**
   * 选择可用的Worker
   * @param {Object} cachedEngine - 缓存的引擎
   * @returns {Worker} 选中的Worker
   */
  static selectWorker(cachedEngine) {
    const { workerPool, nextWorkerIndex } = cachedEngine;
    const worker = workerPool[nextWorkerIndex];
    
    // 更新下一个worker索引（循环使用worker池）
    cachedEngine.nextWorkerIndex = (nextWorkerIndex + 1) % workerPool.length;
    
    return worker;
  }

  /**
   * 更新引擎使用统计
   * @param {Object} cachedEngine - 缓存的引擎
   */
  static updateEngineStats(cachedEngine) {
    cachedEngine.useCount++;
    cachedEngine.lastUsedTime = Date.now();
  }

  /**
   * 创建Worker池
   * @param {string} fromLang - 源语言
   * @param {string} toLang - 目标语言
   * @returns {Array} Worker池
   */
  static async createWorkerPool(fromLang, toLang) {
    Logger.debug("Translator", `开始创建Worker池: ${fromLang} -> ${toLang}`);
    
    // 加载模型
    const needMiddle = fromLang !== "en" && toLang !== "en";
    await this.loadModels(fromLang, toLang, needMiddle);

    // 准备模型数据
    const modelPayloads = this.prepareModelPayloads(fromLang, toLang, needMiddle);

    // 并行创建workers
    const workerPromises = Array.from({ length: WORKERS_PER_LANGUAGE_PAIR }, () =>
      this.createWorker(fromLang, toLang, modelPayloads)
    );

    const workerPool = await Promise.all(workerPromises);
    Logger.info("Translator", `Worker池创建完成，数量: ${workerPool.length}`);
    
    return workerPool;
  }

  /**
   * 准备模型载荷数据
   * @param {string} fromLang - 源语言
   * @param {string} toLang - 目标语言
   * @param {boolean} needMiddle - 是否需要中间语言
   * @returns {Array} 模型载荷数组
   */
  static prepareModelPayloads(fromLang, toLang, needMiddle) {
    if (needMiddle) {
      return [
        this.#models[`${fromLang}_en`],
        this.#models[`en_${toLang}`],
      ];
    } else {
      return [this.#models[`${fromLang}_${toLang}`]];
    }
  }

  /**
   * 创建单个Worker
   * @param {string} fromLang - 源语言
   * @param {string} toLang - 目标语言
   * @param {Array} modelPayloads - 模型载荷
   * @returns {Worker} 创建的Worker
   */
  static async createWorker(fromLang, toLang, modelPayloads) {
    const worker = new Worker(workerPath);

    // 设置事件监听器
    this.setupWorkerEventHandlers(worker, fromLang, toLang);

    // 初始化Worker
    await this.initializeWorker(worker, fromLang, toLang, modelPayloads);

    return worker;
  }

  /**
   * 设置Worker事件处理器
   * @param {Worker} worker - Worker实例
   * @param {string} fromLang - 源语言
   * @param {string} toLang - 目标语言
   */
  static setupWorkerEventHandlers(worker, fromLang, toLang) {
    worker.on("message", (data) => {
      this.handleWorkerMessage(data, fromLang, toLang);
    });

    worker.on("error", (error) => {
      Logger.error("Translator", `Worker错误 (${fromLang}->${toLang}):`, error.message);
      this.handleWorkerError(error, fromLang, toLang);
    });
  }

  /**
   * 处理Worker错误
   * @param {Error} error - 错误对象
   * @param {string} fromLang - 源语言
   * @param {string} toLang - 目标语言
   */
  static handleWorkerError(error, fromLang, toLang) {
    const languagePairKey = this.getLanguagePairKey(fromLang, toLang);
    this.removeEngine(languagePairKey);

    // 拒绝所有待处理的消息
    for (const [messageId, { reject }] of this.#pendingMessages.entries()) {
      reject(new Error(`Worker error: ${error.message}`));
      this.#pendingMessages.delete(messageId);
    }
  }

  /**
   * 初始化Worker
   * @param {Worker} worker - Worker实例
   * @param {string} fromLang - 源语言
   * @param {string} toLang - 目标语言
   * @param {Array} modelPayloads - 模型载荷
   */
  static async initializeWorker(worker, fromLang, toLang, modelPayloads) {
    return new Promise((resolve, reject) => {
      const timeout = setTimeout(() => {
        reject(new Error("Worker initialization timeout"));
      }, WORKER_INIT_TIMEOUT_MS);

      worker.once("message", (data) => {
        if (data.type === MESSAGE_TYPES.WORKER_READY.type) {
          clearTimeout(timeout);

          // 发送初始化消息
          worker.postMessage({
            type: MESSAGE_TYPES.INIT_REQUEST.type,
            sourceLanguage: fromLang,
            targetLanguage: fromLang !== "en" && toLang !== "en" ? "en" : toLang,
            modelPayloads,
          });

          // 等待初始化完成
          worker.once("message", (initData) => {
            if (initData.type === MESSAGE_TYPES.INIT_SUCCESS.type) {
              Logger.debug("Translator", `Worker初始化成功: ${fromLang} -> ${toLang}`);
              resolve();
            } else if (initData.type === MESSAGE_TYPES.INIT_ERROR.type) {
              Logger.error("Translator", `Worker初始化失败: ${initData.error}`);
              reject(new Error(`Worker initialization failed: ${initData.error}`));
            }
          });
        }
      });
    });
  }

  /**
   * 处理Worker消息
   * @param {Object} data - 消息数据
   * @param {string} fromLang - 源语言
   * @param {string} toLang - 目标语言
   */
  static handleWorkerMessage(data, fromLang, toLang) {
    const languagePairKey = this.getLanguagePairKey(fromLang, toLang);

    switch (data.type) {
      case MESSAGE_TYPES.TRANSLATION_RESPONSE.type:
        this.handleTranslationResponse(data, languagePairKey);
        break;

      case MESSAGE_TYPES.TRANSLATION_ERROR.type:
        this.handleTranslationError(data);
        break;

      case MESSAGE_TYPES.TRANSLATIONS_DISCARDED.type:
        this.handleTranslationsDiscarded();
        break;

      default:
        Logger.warn("Translator", `未知的Worker消息类型: ${data.type}`);
    }
  }

  /**
   * 处理翻译响应
   * @param {Object} data - 响应数据
   * @param {string} languagePairKey - 语言对键名
   */
  static handleTranslationResponse(data, languagePairKey) {
    const { messageId, targetText } = data;
    const pendingMessage = this.#pendingMessages.get(messageId);

    if (pendingMessage) {
      pendingMessage.resolve(targetText);
      this.#pendingMessages.delete(messageId);
      Logger.debug("Translator", `翻译完成，消息ID: ${messageId}`);
    }

    this.keepAlive(languagePairKey);
  }

  /**
   * 处理翻译错误
   * @param {Object} data - 错误数据
   */
  static handleTranslationError(data) {
    const { messageId, error } = data;
    const pendingMessage = this.#pendingMessages.get(messageId);

    if (pendingMessage) {
      Logger.error("Translator", `翻译错误，消息ID: ${messageId}`, error.message);
      pendingMessage.reject(new Error(error.message));
      this.#pendingMessages.delete(messageId);
    }
  }

  /**
   * 处理翻译被丢弃
   */
  static handleTranslationsDiscarded() {
    Logger.warn("Translator", "翻译队列被丢弃，拒绝所有待处理消息");
    
    for (const [messageId, { reject }] of this.#pendingMessages.entries()) {
      reject(new Error("Translations were discarded"));
      this.#pendingMessages.delete(messageId);
    }
  }

  /**
   * 丢弃翻译队列
   * @param {string} languagePairKey - 语言对键名
   */
  static discardTranslations(languagePairKey) {
    const cachedEngine = this.#cachedEngines.get(languagePairKey);
    if (!cachedEngine) return;

    Logger.info("Translator", `丢弃翻译队列: ${languagePairKey}`);
    
    cachedEngine.workerPool.forEach((worker) => {
      worker.postMessage({
        type: MESSAGE_TYPES.DISCARD_TRANSLATION_QUEUE.type,
      });
    });
  }

  /**
   * 保持引擎活跃状态
   * @param {string} languagePairKey - 语言对键名
   */
  static keepAlive(languagePairKey) {
    const cachedEngine = this.#cachedEngines.get(languagePairKey);
    if (!cachedEngine) return;

    cachedEngine.lastUsedTime = Date.now();

    // 避免频繁重置超时计时器
    const timeSinceLastReset = Date.now() - (cachedEngine.lastTimeoutReset || 0);
    const shouldResetTimeout = !cachedEngine.timeoutId || timeSinceLastReset > TIMEOUT_RESET_THRESHOLD_MS;

    if (shouldResetTimeout) {
      this.resetEngineTimeout(cachedEngine, languagePairKey);
    }

    this.#cachedEngines.set(languagePairKey, cachedEngine);
  }

  /**
   * 重置引擎超时计时器
   * @param {Object} cachedEngine - 缓存的引擎
   * @param {string} languagePairKey - 语言对键名
   */
  static resetEngineTimeout(cachedEngine, languagePairKey) {
    // 清除现有的超时计时器
    if (cachedEngine.timeoutId) {
      clearTimeout(cachedEngine.timeoutId);
    }

    // 设置新的超时计时器
    if (ENGINE_CACHE_ENABLE && ENGINE_CACHE_TIMEOUT_MINUTES > 0) {
      cachedEngine.timeoutId = setTimeout(() => {
        this.removeEngine(languagePairKey);
      }, ENGINE_CACHE_TIMEOUT_MS);

      cachedEngine.lastTimeoutReset = Date.now();
    }
  }

  /**
   * 移除引擎
   * @param {string} languagePairKey - 语言对键名
   */
  static removeEngine(languagePairKey) {
    const cachedEngine = this.#cachedEngines.get(languagePairKey);

    if (cachedEngine) {
      Logger.info("Translator", `移除翻译引擎: ${languagePairKey}`);
      
      if (cachedEngine.timeoutId) {
        clearTimeout(cachedEngine.timeoutId);
      }

      if (cachedEngine.workerPool) {
        cachedEngine.workerPool.forEach((worker) => {
          worker.terminate();
        });
      }
    }

    this.#cachedEngines.delete(languagePairKey);

    // 如果没有更多的引擎，停止内存释放定时器
    if (this.#cachedEngines.size === 0) {
      this.stopMemoryReleaseTimer();
    }
  }

  /**
   * 启动内存释放定时器
   */
  static startMemoryReleaseTimer() {
    if (this.#memoryReleaseTimer || !ENGINE_CACHE_ENABLE || ENGINE_CACHE_TIMEOUT_MINUTES <= 0) {
      return;
    }

    Logger.debug("Translator", "启动内存释放定时器");
    
    this.#memoryReleaseTimer = setInterval(() => {
      this.checkAndReleaseMemory();
    }, MEMORY_CHECK_INTERVAL_MS);
  }

  /**
   * 停止内存释放定时器
   */
  static stopMemoryReleaseTimer() {
    if (this.#memoryReleaseTimer) {
      Logger.debug("Translator", "停止内存释放定时器");
      clearInterval(this.#memoryReleaseTimer);
      this.#memoryReleaseTimer = null;
    }
  }

  /**
   * 检查并释放内存
   */
  static checkAndReleaseMemory() {
    if (!ENGINE_CACHE_ENABLE || ENGINE_CACHE_TIMEOUT_MINUTES <= 0 || this.#cachedEngines.size === 0) {
      return;
    }

    const now = Date.now();
    const timeoutMs = ENGINE_CACHE_TIMEOUT_MS;
    let releasedCount = 0;

    Logger.debug("Translator", `开始内存检查，当前引擎数量: ${this.#cachedEngines.size}`);

    // 检查需要释放的引擎
    for (const [languagePairKey, cachedEngine] of this.#cachedEngines.entries()) {
      if (now - cachedEngine.lastUsedTime >= timeoutMs) {
        Logger.info("Translator", 
          `自动释放模型: ${languagePairKey}，闲置时间: ${Config.RELEASE_INTERVAL} 分钟`
        );
        this.removeEngine(languagePairKey);
        releasedCount++;
      }
    }

    if (releasedCount > 0) {
      Logger.info("Translator", `内存检查完成，释放了 ${releasedCount} 个引擎`);
      this.releaseUnusedModelMemory();
    }
  }

  /**
   * 关闭所有引擎
   */
  static async Shutdown() {
    Logger.info("Translator", "开始关闭所有翻译引擎");
    
    // 停止内存释放定时器
    this.stopMemoryReleaseTimer();

    // 拒绝所有待处理的消息
    for (const [messageId, { reject }] of this.#pendingMessages.entries()) {
      reject(new Error("All engines are being closed"));
      this.#pendingMessages.delete(messageId);
    }

    // 关闭所有引擎
    const shutdownPromises = [];
    for (const [languagePairKey, cachedEngine] of this.#cachedEngines.entries()) {
      if (cachedEngine.timeoutId) {
        clearTimeout(cachedEngine.timeoutId);
      }

      if (cachedEngine.workerPool) {
        for (const worker of cachedEngine.workerPool) {
          shutdownPromises.push(worker.terminate());
        }
      }
    }

    await Promise.all(shutdownPromises);
    this.#cachedEngines.clear();
    
    Logger.info("Translator", "所有翻译引擎已关闭");
  }

  /**
   * 加载模型
   * @param {string} fromLang - 源语言
   * @param {string} toLang - 目标语言
   * @param {boolean} needMiddle - 是否需要中间语言
   */
  static async loadModels(fromLang, toLang, needMiddle = false) {
    if (!this.#models) {
      Logger.debug("Translator", "初始化模型管理器");
      await Models.init();
      this.#models = {};
    }

    // 加载已下载模型
    const downloadedModels = await Models.getDownloadedModels();
    await this.loadDownloadedModels(downloadedModels);

    // 加载所需模型
    if (!downloadedModels.includes(`${fromLang}_${toLang}`)) {
      await this.loadRequiredModels(fromLang, toLang, needMiddle);
    }
  }

  /**
   * 加载已下载的模型
   * @param {Array} downloadedModels - 已下载的模型列表
   */
  static async loadDownloadedModels(downloadedModels) {
    const loadPromises = downloadedModels
      .filter(model => !this.#models[model])
      .map(async (model) => {
        const [_fromLang, _toLang] = model.split("_");
        const payload = await Models.getModel(_fromLang, _toLang);
        this.#models[model] = payload;
        Logger.debug("Translator", `加载已下载模型: ${model}`);
      });

    await Promise.all(loadPromises);
  }

  /**
   * 加载所需模型
   * @param {string} fromLang - 源语言
   * @param {string} toLang - 目标语言
   * @param {boolean} needMiddle - 是否需要中间语言
   */
  static async loadRequiredModels(fromLang, toLang, needMiddle) {
    if (needMiddle) {
      Logger.info("Translator", `加载中间翻译模型: ${fromLang} -> en -> ${toLang}`);
      const [payload1, payload2] = await Promise.all([
        Models.getModel(fromLang, "en"),
        Models.getModel("en", toLang)
      ]);
      this.#models[`${fromLang}_en`] = payload1;
      this.#models[`en_${toLang}`] = payload2;
    } else {
      Logger.info("Translator", `加载直接翻译模型: ${fromLang} -> ${toLang}`);
      const payload = await Models.getModel(fromLang, toLang);
      this.#models[`${fromLang}_${toLang}`] = payload;
    }
  }

  /**
   * 释放未使用的模型内存
   */
  static releaseUnusedModelMemory() {
    if (!this.#models) return;

    // 获取当前正在使用的模型列表
    const activeModels = this.getActiveModels();

    // 释放不在活跃列表中的模型
    const modelKeys = Object.keys(this.#models);
    let releasedCount = 0;

    for (const modelKey of modelKeys) {
      if (!activeModels.has(modelKey)) {
        delete this.#models[modelKey];
        releasedCount++;
        Logger.debug("Translator", `释放模型内存: ${modelKey}`);
      }
    }

    if (releasedCount > 0) {
      Logger.info("Translator", `释放了 ${releasedCount} 个未使用的模型`);
    }
  }

  /**
   * 获取当前活跃的模型集合
   * @returns {Set} 活跃模型集合
   */
  static getActiveModels() {
    const activeModels = new Set();
    
    for (const [languagePairKey] of this.#cachedEngines.entries()) {
      const [fromLang, toLang] = languagePairKey.split("_");

      // 直接翻译模型
      activeModels.add(`${fromLang}_${toLang}`);

      // 中间翻译模型
      if (fromLang !== "en" && toLang !== "en") {
        activeModels.add(`${fromLang}_en`);
        activeModels.add(`en_${toLang}`);
      }
    }

    return activeModels;
  }

  /**
   * 检测文本语言
   * @param {string} text - 待检测文本
   * @returns {string} 检测到的语言代码
   */
  static async DetectLang(text) {
    try {
      // 输入验证
      if (typeof text !== "string" || !text.trim()) {
        Logger.warn("Translator", "语言检测输入无效，返回默认语言");
        return "en";
      }

      // 限制检测文本长度
      const textToDetect = text.length > MAX_DETECTION_LENGTH
        ? text.substring(0, MAX_DETECTION_LENGTH)
        : text;

      Logger.debug("Translator", `开始语言检测，文本长度: ${textToDetect.length}`);

      // 调用语言检测器
      const result = await LanguageDetector.detect(textToDetect);

      // 验证检测结果
      if (!result || !result.language || result.language === "un") {
        Logger.warn("Translator", "语言检测失败或不确定，返回默认语言");
        return "en";
      }

      const detectedLang = Lang.LD.includes(result.language) 
        ? Lang.LD2M[result.language] 
        : result.language;

      Logger.debug("Translator", `语言检测完成: ${detectedLang}`);
      return detectedLang;
    } catch (error) {
      Logger.error("Translator", "语言检测失败:", error.message);
      return "en";
    }
  }

  /**
   * 翻译文本
   * @param {string|Array} text - 待翻译文本或文本数组
   * @param {string} fromLang - 源语言
   * @param {string} toLang - 目标语言
   * @param {boolean} isHTML - 是否为HTML格式
   * @returns {string|Array} 翻译结果
   */
  static async Translate(text, fromLang, toLang, isHTML = false) {
    Logger.timeStart("Translator", "翻译任务");
    
    try {
      const isTextArray = Array.isArray(text);
      let texts = isTextArray ? text : [text];

      // 输入验证
      if (!texts || texts.length < 1) {
        Logger.warn("Translator", "翻译输入为空");
        return isTextArray ? [] : "";
      }

      Logger.info("Translator", `开始翻译任务: ${fromLang} -> ${toLang}, 文本数量: ${texts.length}`);

      // 语言处理
      const { processedFromLang, processedToLang } = await this.processLanguages(texts[0], fromLang, toLang);

      // 检查是否需要翻译
      if (processedFromLang === processedToLang) {
        Logger.info("Translator", "源语言与目标语言相同，直接返回");
        return text;
      }

      // 执行翻译
      const result = await this.executeTranslation(texts, processedFromLang, processedToLang, isHTML);
      
      Logger.info("Translator", "翻译任务完成");
      return isTextArray ? result : result[0];
    } catch (error) {
      Logger.error("Translator", "翻译失败:", error.message);
      throw error;
    } finally {
      Logger.timeEnd("Translator", "翻译任务");
    }
  }

  /**
   * 处理语言代码
   * @param {string} firstText - 第一个文本（用于语言检测）
   * @param {string} fromLang - 源语言
   * @param {string} toLang - 目标语言
   * @returns {Object} 处理后的语言代码
   */
  static async processLanguages(firstText, fromLang, toLang) {
    let processedFromLang = fromLang;
    let processedToLang = toLang;

    // 自动检测语言
    if (fromLang === "auto") {
      processedFromLang = await this.DetectLang(firstText);
      Logger.info("Translator", `自动检测源语言: ${processedFromLang}`);
    }

    // 验证语言代码
    this.validateLanguageCodes(processedFromLang, processedToLang);

    // 处理语言别名
    processedFromLang = Lang.MALIAS[processedFromLang] || processedFromLang;
    processedToLang = Lang.MALIAS[processedToLang] || processedToLang;

    return { processedFromLang, processedToLang };
  }

  /**
   * 验证语言代码
   * @param {string} fromLang - 源语言
   * @param {string} toLang - 目标语言
   */
  static validateLanguageCodes(fromLang, toLang) {
    if (!Lang.MALL.includes(fromLang)) {
      throw new Error(`Invalid from language code: ${fromLang}`);
    }
    if (!Lang.MALL.includes(toLang)) {
      throw new Error(`Invalid to language code: ${toLang}`);
    }
  }

  /**
   * 执行翻译
   * @param {Array} texts - 文本数组
   * @param {string} fromLang - 源语言
   * @param {string} toLang - 目标语言
   * @param {boolean} isHTML - 是否为HTML格式
   * @returns {Array} 翻译结果数组
   */
  static async executeTranslation(texts, fromLang, toLang, isHTML) {
    // 分析翻译需求
    const translationPlan = this.analyzeTranslationPlan(fromLang, toLang);
    
    // 预处理
    if (translationPlan.needPreProcess) {
      texts = await this.preprocessTexts(texts, translationPlan.preProcessType);
    }

    // 执行翻译
    if (translationPlan.pureCC && translationPlan.pureCCComplex) {
      // 纯繁体转换
      texts = await this.performComplexCCTranslation(texts, translationPlan);
    } else if (translationPlan.pureCC && translationPlan.needPreProcess) {
      // 简单繁简转换
      texts = await this.preprocessTexts(texts, translationPlan.preProcessType);
    } else if (fromLang !== toLang) {
      // 标准翻译
      const engine = await this.Preload(fromLang, toLang);
      texts = await engine.translate(texts, isHTML);
    }

    // 后处理
    if (translationPlan.needPostProcess) {
      texts = await this.postprocessTexts(texts, translationPlan.postProcessType);
    }

    return texts;
  }

  /**
   * 分析翻译计划
   * @param {string} fromLang - 源语言
   * @param {string} toLang - 目标语言
   * @returns {Object} 翻译计划
   */
  static analyzeTranslationPlan(fromLang, toLang) {
    const plan = {
      needPreProcess: false,
      preProcessType: null,
      needPostProcess: false,
      postProcessType: null,
      pureCC: false,
      pureCCComplex: false,
      pureCCComplexType1: null,
      pureCCComplexType2: null,
    };

    // 源语言预处理
    if (Lang.MC2ZH.includes(fromLang)) {
      plan.needPreProcess = true;
      plan.preProcessType = Lang.CC2ZH[fromLang];
      fromLang = "zh-Hans";
    }

    // 目标语言后处理
    if (Lang.MC2ZH.includes(toLang)) {
      plan.needPostProcess = true;
      plan.postProcessType = Lang.ZH2CC[toLang];
      toLang = "zh-Hans";
    }

    // 检查纯简繁转换
    if (Lang.MCC.includes(fromLang) && Lang.MCC.includes(toLang)) {
      plan.pureCC = true;
      
      if (plan.needPreProcess && plan.needPostProcess) {
        plan.pureCCComplex = true;
        plan.pureCCComplexType1 = plan.preProcessType;
        plan.pureCCComplexType2 = plan.postProcessType;
      }
    }

    return plan;
  }

  /**
   * 预处理文本
   * @param {Array} texts - 文本数组
   * @param {string} processType - 处理类型
   * @returns {Array} 处理后的文本数组
   */
  static async preprocessTexts(texts, processType) {
    Logger.debug("Translator", `开始预处理，类型: ${processType}`);
    return Promise.all(texts.map(text => OpenCC.convert(text, processType)));
  }

  /**
   * 后处理文本
   * @param {Array} texts - 文本数组
   * @param {string} processType - 处理类型
   * @returns {Array} 处理后的文本数组
   */
  static async postprocessTexts(texts, processType) {
    Logger.debug("Translator", `开始后处理，类型: ${processType}`);
    return Promise.all(texts.map(text => OpenCC.convert(text, processType)));
  }

  /**
   * 执行复杂的繁体转换
   * @param {Array} texts - 文本数组
   * @param {Object} plan - 翻译计划
   * @returns {Array} 转换后的文本数组
   */
  static async performComplexCCTranslation(texts, plan) {
    Logger.debug("Translator", "执行复杂繁体转换");
    return Promise.all(texts.map(async (text) => {
      const intermediateText = await OpenCC.convert(text, plan.pureCCComplexType1);
      return OpenCC.convert(intermediateText, plan.pureCCComplexType2);
    }));
  }

  /**
   * 获取引擎统计信息
   * @returns {Object} 统计信息
   */
  static getStats() {
    const stats = {
      activeEngines: this.#cachedEngines.size,
      pendingMessages: this.#pendingMessages.size,
      loadedModels: this.#models ? Object.keys(this.#models).length : 0,
      memoryTimerActive: !!this.#memoryReleaseTimer,
      engines: [],
    };

    for (const [key, engine] of this.#cachedEngines.entries()) {
      stats.engines.push({
        languagePair: key,
        workerCount: engine.workerPool?.length || 0,
        useCount: engine.useCount,
        lastUsedTime: new Date(engine.lastUsedTime).toISOString(),
        hasTimeout: !!engine.timeoutId,
      });
    }

    return stats;
  }
}

module.exports = Translator;
