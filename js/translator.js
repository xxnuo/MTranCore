// 正常导入
const path = require("path");
const { Worker } = require("worker_threads");
const Lang = require("./lang");
const Models = require("./models");
const OpenCC = require("./opencc");
const LanguageDetector = require("./ld");
const { MESSAGE_TYPES } = require("./message");
const Config = require("./config");

// 引擎缓存超时时间（分钟）
const ENGINE_CACHE_TIMEOUT_MINUTES = 30;
const ENGINE_CACHE_TIMEOUT_MS = ENGINE_CACHE_TIMEOUT_MINUTES * 60 * 1000;

// 每个语言对的worker数量
const WORKERS_PER_LANGUAGE_PAIR = Config.WORKERS;

const workerPath = path.join(__dirname, "worker.js");

// 翻译引擎管理器
class Translator {
  static #cachedEngines = new Map();
  static #messageId = 0;
  static #pendingMessages = new Map();

  static GetSupportLanguages() {
    return Lang.MALL;
  }

  static getLanguagePairKey(fromLang, toLang) {
    return `${fromLang}_${toLang}`;
  }

  static async Preload(fromLang, toLang) {
    const languagePairKey = this.getLanguagePairKey(fromLang, toLang);
    const cachedEngine = this.#cachedEngines.get(languagePairKey);

    if (cachedEngine) {
      // 重置超时计时器
      this.keepAlive(languagePairKey);
      return {
        translate: (texts, isHTML) =>
          this.translateWithWorker(languagePairKey, texts, isHTML),
        discardTranslations: () => this.discardTranslations(languagePairKey),
      };
    }

    try {
      // 创建worker池
      const workerPool = await this.createWorkerPool(fromLang, toLang);

      // 添加到缓存并设置超时
      this.#cachedEngines.set(languagePairKey, {
        workerPool,
        nextWorkerIndex: 0, // 用于轮询选择worker的索引
        timeoutId: setTimeout(() => {
          this.removeEngine(languagePairKey);
        }, ENGINE_CACHE_TIMEOUT_MS),
      });

      return {
        translate: (texts, isHTML) =>
          this.translateWithWorker(languagePairKey, texts, isHTML),
        discardTranslations: () => this.discardTranslations(languagePairKey),
      };
    } catch (error) {
      throw new Error(`Create translation engine failed: ${error.message}`);
    }
  }

  static async translateWithWorker(languagePairKey, texts, isHTML) {
    const cachedEngine = this.#cachedEngines.get(languagePairKey);
    if (!cachedEngine) {
      throw new Error(`Translation engine not found: ${languagePairKey}`);
    }

    this.keepAlive(languagePairKey);

    const { workerPool, nextWorkerIndex } = cachedEngine;
    const isTextArray = Array.isArray(texts);
    const sourceTexts = isTextArray ? texts : [texts];

    // 创建翻译任务
    const results = [];
    for (const sourceText of sourceTexts) {
      if (sourceText && sourceText.trim()) {
        const messageId = this.#messageId++;
        const translationId = Date.now() + Math.random();

        // 使用轮询方式选择worker
        const workerIndex = cachedEngine.nextWorkerIndex;
        const worker = workerPool[workerIndex];

        // 更新下一个worker索引（循环使用worker池）
        cachedEngine.nextWorkerIndex = (workerIndex + 1) % workerPool.length;

        const result = await new Promise((resolve, reject) => {
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

        results.push(result);
      } else {
        results.push("");
      }
    }

    return isTextArray ? results : results[0];
  }

  static async createWorkerPool(fromLang, toLang) {
    // 加载模型
    const needMiddle = fromLang !== "en" && toLang !== "en";
    await this.loadModels(fromLang, toLang, needMiddle);

    // 准备模型数据
    let modelPayloads = [];
    if (needMiddle) {
      modelPayloads = [
        this.models[`${fromLang}_en`],
        this.models[`en_${toLang}`],
      ];
    } else {
      modelPayloads = [this.models[`${fromLang}_${toLang}`]];
    }

    // 创建worker池
    const workerPool = [];

    for (let i = 0; i < WORKERS_PER_LANGUAGE_PAIR; i++) {
      const worker = await this.createWorker(fromLang, toLang, modelPayloads);
      workerPool.push(worker);
    }

    return workerPool;
  }

  static async createWorker(fromLang, toLang, modelPayloads) {
    // 创建worker
    const worker = new Worker(workerPath);

    // 设置消息处理器
    worker.on("message", (data) => {
      this.handleWorkerMessage(data, fromLang, toLang);
    });

    worker.on("error", (error) => {
      console.error(`Worker error (${fromLang}->${toLang}):`, error);
      this.removeEngine(this.getLanguagePairKey(fromLang, toLang));

      // 拒绝所有待处理的消息
      for (const [messageId, { reject }] of this.#pendingMessages.entries()) {
        reject(new Error(`Worker error: ${error.message}`));
        this.#pendingMessages.delete(messageId);
      }
    });

    // 等待 worker 初始化
    await new Promise((resolve, reject) => {
      const timeout = setTimeout(() => {
        reject(new Error("Worker initialization timeout"));
      }, 600000); // 600秒超时

      worker.once("message", (data) => {
        if (data.type === MESSAGE_TYPES.WORKER_READY.type) {
          clearTimeout(timeout);

          // 发送初始化消息
          worker.postMessage({
            type: MESSAGE_TYPES.INIT_REQUEST.type,
            sourceLanguage: fromLang,
            targetLanguage:
              fromLang !== "en" && toLang !== "en" ? "en" : toLang,
            modelPayloads,
            // logLevel: "Info",
          });

          // 等待初始化完成
          worker.once("message", (initData) => {
            if (initData.type === MESSAGE_TYPES.INIT_SUCCESS.type) {
              resolve();
            } else if (initData.type === MESSAGE_TYPES.INIT_ERROR.type) {
              reject(
                new Error(`Worker initialization failed: ${initData.error}`)
              );
            }
          });
        }
      });
    });

    return worker;
  }

  static handleWorkerMessage(data, fromLang, toLang) {
    const languagePairKey = this.getLanguagePairKey(fromLang, toLang);

    switch (data.type) {
      case MESSAGE_TYPES.TRANSLATION_RESPONSE.type: {
        const { messageId, targetText } = data;
        const pendingMessage = this.#pendingMessages.get(messageId);

        if (pendingMessage) {
          pendingMessage.resolve(targetText);
          this.#pendingMessages.delete(messageId);
        }

        // 重置超时
        this.keepAlive(languagePairKey);
        break;
      }

      case MESSAGE_TYPES.TRANSLATION_ERROR.type: {
        const { messageId, error } = data;
        const pendingMessage = this.#pendingMessages.get(messageId);

        if (pendingMessage) {
          pendingMessage.reject(new Error(error.message));
          this.#pendingMessages.delete(messageId);
        }
        break;
      }

      case MESSAGE_TYPES.TRANSLATIONS_DISCARDED.type: {
        // 拒绝所有待处理的消息
        for (const [messageId, { reject }] of this.#pendingMessages.entries()) {
          reject(new Error("Translations were discarded"));
          this.#pendingMessages.delete(messageId);
        }
        break;
      }
    }
  }

  static discardTranslations(languagePairKey) {
    const cachedEngine = this.#cachedEngines.get(languagePairKey);
    if (!cachedEngine) return;

    cachedEngine.workerPool.forEach((worker) => {
      worker.postMessage({
        type: MESSAGE_TYPES.DISCARD_TRANSLATION_QUEUE.type,
      });
    });
  }

  static keepAlive(languagePairKey) {
    const cachedEngine = this.#cachedEngines.get(languagePairKey);
    if (!cachedEngine) return;

    // 清除现有的超时计时器
    if (cachedEngine.timeoutId) {
      clearTimeout(cachedEngine.timeoutId);
    }

    // 设置新的超时计时器
    cachedEngine.timeoutId = setTimeout(() => {
      this.removeEngine(languagePairKey);
    }, ENGINE_CACHE_TIMEOUT_MS);

    this.#cachedEngines.set(languagePairKey, cachedEngine);
  }

  static removeEngine(languagePairKey) {
    const cachedEngine = this.#cachedEngines.get(languagePairKey);

    if (cachedEngine) {
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
  }

  static async Shutdown() {
    // 拒绝所有待处理的消息
    for (const [messageId, { reject }] of this.#pendingMessages.entries()) {
      reject(new Error("All engines are being closed"));
      this.#pendingMessages.delete(messageId);
    }

    // 关闭所有引擎
    for (const [
      languagePairKey,
      cachedEngine,
    ] of this.#cachedEngines.entries()) {
      if (cachedEngine.timeoutId) {
        clearTimeout(cachedEngine.timeoutId);
      }

      if (cachedEngine.workerPool) {
        cachedEngine.workerPool.forEach((worker) => {
          worker.terminate();
        });
      }
    }

    // 清空缓存
    this.#cachedEngines.clear();
  }

  static models = null;

  static async loadModels(fromLang, toLang, needMiddle = false) {
    if (!this.models) {
      await Models.init();
      this.models = {};
    }

    // 加载已下载模型
    const downloadedModels = await Models.getDownloadedModels();
    for (const model of downloadedModels) {
      if (!this.models[model]) {
        const [_fromLang, _toLang] = model.split("_");
        const payload = await Models.getModel(_fromLang, _toLang);
        this.models[model] = payload;
      }
    }

    if (!downloadedModels.includes(`${fromLang}_${toLang}`)) {
      if (needMiddle) {
        const payload1 = await Models.getModel(fromLang, "en");
        const payload2 = await Models.getModel("en", toLang);
        this.models[`${fromLang}_en`] = payload1;
        this.models[`en_${toLang}`] = payload2;
      } else {
        const payload = await Models.getModel(fromLang, toLang);
        this.models[`${fromLang}_${toLang}`] = payload;
      }
    }
  }

  static async DetectLang(text) {
    try {
      // 确保输入是字符串
      if (typeof text !== "string" || !text.trim()) {
        return "en"; // 如果输入无效或为空，默认返回英语
      }

      // 因为只需要输出单一代码，限制检测文本长度，避免处理过大的文本
      const maxDetectionLength = 64;
      const textToDetect =
        text.length > maxDetectionLength
          ? text.substring(0, maxDetectionLength)
          : text;

      // 调用语言检测器
      const result = await LanguageDetector.detect(textToDetect);

      // 验证检测结果
      if (!result || !result.language || result.language === "un") {
        return "en"; // 如果检测失败或不确定，默认返回英语
      }

      if (Lang.LD.includes(result.language)) {
        return Lang.LD2M[result.language];
      }
      return result.language;
    } catch (error) {
      console.error("Language detection failed:", error);
      return "en"; // 检测失败时默认返回英语
    }
  }

  // 翻译，text 可以是单个文本，也可以是文本数组
  // 如果 text 是数组，则返回数组，否则返回单个文本
  // 如果 text 是数组且 fromLang 为 auto，则全部文本的原语言会使用第一个文本检测出来的语言
  static async Translate(text, fromLang, toLang, isHTML = false) {
    const isTextArray = Array.isArray(text);
    let texts = isTextArray ? text : [text];

    // 如果文本为空，直接返回
    if (!texts || texts.length < 1) {
      return isTextArray ? [] : "";
    }

    // 转换后的语言代码
    let _fromLang = fromLang;
    let _toLang = toLang;

    // 自动检测语言
    if (fromLang === "auto") {
      _fromLang = await this.DetectLang(texts[0]);
    }

    // 检查语言代码是否有效
    if (!Lang.MALL.includes(_fromLang)) {
      throw new Error("Invalid from language code");
    }
    if (!Lang.MALL.includes(_toLang)) {
      throw new Error("Invalid to language code");
    }

    // 检查语言代码是否为别名
    const aliasFromLang = Lang.MALIAS[_fromLang];
    if (aliasFromLang) {
      _fromLang = aliasFromLang;
    }
    const aliasToLang = Lang.MALIAS[_toLang];
    if (aliasToLang) {
      _toLang = aliasToLang;
    }

    // 如果语言相同，直接返回
    if (_fromLang === _toLang) {
      return text;
    }

    // 翻译前后处理
    let needPreProcess = false;
    let preProcessType = null;
    let needPostProcess = false;
    let postProcessType = null;

    if (Lang.MC2ZH.includes(_fromLang)) {
      // 源语言为繁体中文，转换为简体中文拿去翻译
      needPreProcess = true;
      preProcessType = Lang.CC2ZH[_fromLang];
      _fromLang = "zh-Hans";
    }

    if (Lang.MC2ZH.includes(_toLang)) {
      // 目标语言为繁体中文，标记需要后处理
      needPostProcess = true;
      postProcessType = Lang.ZH2CC[_toLang];
      _toLang = "zh-Hans";
    }

    // 检查是否为纯简繁转换
    let pureCC = false;
    // let pureCCType = null;
    // 繁繁转换
    let pureCCComplex = false;
    let pureCCComplexType1 = null;
    let pureCCComplexType2 = null;
    if (Lang.MCC.includes(_fromLang) && Lang.MCC.includes(_toLang)) {
      // // 纯简繁转换
      pureCC = true;
      // if (needPreProcess && !needPostProcess) {
      //   // 繁体转简体
      //   pureCCType = Lang.CC2ZH[_fromLang];
      // } else if (!needPreProcess && needPostProcess) {
      //   // 简体转繁体
      //   pureCCType = Lang.ZH2CC[_toLang];
      // }
      if (needPreProcess && needPostProcess) {
        // 繁繁转换
        pureCCComplex = true;
        pureCCComplexType1 = Lang.CC2ZH[_fromLang];
        pureCCComplexType2 = Lang.ZH2CC[_toLang];
      }
    }
    // 参数已处理就绪，开始翻译

    // 获取引擎
    let engine = null;
    if (pureCC && _fromLang !== "zh-Hans") {
      if (pureCCComplex) {
        // 使用Promise.all并行处理数组
        return Promise.all(
          texts.map(async (item) => {
            const _text = await OpenCC.convert(item, pureCCComplexType1);
            return OpenCC.convert(_text, pureCCComplexType2);
          })
        );
      } else {
        // 使用Promise.all并行处理数组
        texts = await Promise.all(
          texts.map(async (item) => {
            return await OpenCC.convert(item, preProcessType);
          })
        );
      }
    } else {
      // 使用引擎管理器获取或创建引擎
      if (_fromLang !== _toLang) {
        engine = await this.Preload(_fromLang, _toLang);
      }
    }

    // 预处理
    if (needPreProcess) {
      for (let i = 0; i < texts.length; i++) {
        texts[i] = await OpenCC.convert(texts[i], preProcessType);
      }
    }
    // 翻译，自带了批量翻译和单文本翻译
    if (!pureCC && _fromLang !== _toLang) {
      texts = await engine.translate(texts, isHTML);
    }
    // 后处理
    if (needPostProcess) {
      for (let i = 0; i < texts.length; i++) {
        texts[i] = await OpenCC.convert(texts[i], postProcessType);
      }
    }
    // 返回
    return isTextArray ? texts : texts[0];
  }
}

module.exports = Translator;
