"use strict";

/**
 * @typedef {import("../translations").Bergamot} Bergamot
 * @typedef {import("../translations").TranslationModelPayload} TranslationModelPayload
 * @typedef {import("../translations").LanguageTranslationModelFiles} LanguageTranslationModelFiles
 */

const fs = require("fs");
const path = require("path");
const Lang = require("./lang");
const loadBergamot = require("./bergamot");
const Config = require("./config");
const Logger = require("./logger");

const WASM_PATH = path.join(__dirname, "./bergamot.wasm");
const WASM_BINARY = fs.readFileSync(WASM_PATH);

// 每种文件类型的内存对齐要求，文件类型字符串应与模型注册表中的相同
const MODEL_FILE_ALIGNMENTS = {
  model: 256,
  lex: 64,
  vocab: 64,
  qualityModel: 64,
  srcvocab: 64,
  trgvocab: 64,
};

// 常量定义
const WHITESPACE_REGEX = /^(\s*)(.*?)(\s*)$/s; // 此正则表达式匹配文本前后的空白，以便保留它们
const FULL_WIDTH_PUNCTUATION_LANGUAGE_TAGS = ["ja", "ko", "zh", ...Lang.MCC]; // 使用全角标点符号的语言列表
const FULL_WIDTH_PUNCTUATION_REGEX = /([。！？])"/g; // 此正则表达式帮助对使用全角标点符号的语言进行句子分割
const MEMORY_CLEANUP_INTERVAL = Config.CLEANUP_INTERVAL; // 每 x 次翻译清理一次内存
const MEMORY_CLEANUP_TIME_THRESHOLD = Config.CLEANUP_TIME_THRESHOLD * 60 * 1000; // y 分钟内存清理时间阈值

/**
 * 在将文本发送到翻译引擎之前，执行必要的清理步骤
 *
 * @param {string} sourceLanguage - 源语言的BCP-47语言标签
 * @param {string} sourceText - 需要清理的源文本
 * @returns {{ whitespaceBefore: string, whitespaceAfter: string, cleanedSourceText: string }}
 */
function CleanText(sourceLanguage, sourceText) {
  // 文本开头或结尾的空白可能会混淆翻译，但会影响最终结果的呈现
  const result = WHITESPACE_REGEX.exec(sourceText);
  if (!result) {
    throw new Error("Empty whitespace regex should always return a result.");
  }

  const whitespaceBefore = result[1];
  const whitespaceAfter = result[3];
  let cleanedSourceText = result[2];

  // 删除所有软连字符，因为它们会破坏分词
  cleanedSourceText = cleanedSourceText.replaceAll("\u00AD", "");

  if (FULL_WIDTH_PUNCTUATION_LANGUAGE_TAGS.includes(sourceLanguage)) {
    // 在全角标点符号后面添加空格，当它前面有左双引号时，这样可以欺骗分段算法在那里断句
    cleanedSourceText = cleanedSourceText.replaceAll(
      FULL_WIDTH_PUNCTUATION_REGEX,
      "$1 “" // 注意这行字符串的字符不是语法错误，必须不能改变！
    );
  }

  return { whitespaceBefore, whitespaceAfter, cleanedSourceText };
}

/**
 * 初始化 Bergamot 翻译引擎
 *
 * @returns {Promise<Bergamot>}
 */
async function InitWasm() {
  return new Promise((resolve, reject) => {
    const bergamot = loadBergamot({
      INITIAL_MEMORY: 234_291_200,
      print: (...args) => Logger.log("Engine", ...args),
      onAbort: () => reject(new Error("Error loading Bergamot wasm module.")),
      onRuntimeInitialized: async () => {
        await Promise.resolve();
        Logger.debug("Engine", "Bergamot WASM module initialized successfully");
        resolve(bergamot);
      },
      wasmBinary: WASM_BINARY,
    });
  });
}

/**
 * 为 Marian 翻译服务生成配置。它需要特定的空白格式
 *
 * @param {Record<string, string>} config
 * @returns {string}
 */
function generateTextConfig(config) {
  const indent = "            ";
  let result = "\n";

  for (const [key, value] of Object.entries(config)) {
    result += `${indent}${key}: ${value}\n`;
  }

  return result + indent;
}

/**
 * JS 对象需要转换为 wasm 对象以配置翻译引擎
 *
 * @param {Bergamot} bergamot
 * @param {string[]} sourceTexts
 * @param {boolean} isHTML
 * @returns {{ messages: Bergamot["VectorString"], options: Bergamot["VectorResponseOptions"] }}
 */
function getTranslationArgs(bergamot, sourceTexts, isHTML) {
  const messages = new bergamot.VectorString();
  const options = new bergamot.VectorResponseOptions();

  // 预先创建一次响应选项对象，避免在循环中重复创建
  const responseOption = {
    qualityScores: false,
    alignment: true,
    html: isHTML,
  };

  // 添加所有非空文本
  for (const sourceText of sourceTexts) {
    if (sourceText) {
      try {
        // 清理文本，确保不含无效字符
        const cleanedText = String(sourceText).trim();
        if (cleanedText.length > 0) {
          messages.push_back(cleanedText);
          options.push_back(responseOption);
        }
      } catch (error) {
        Logger.error("Engine", "Error processing text for translation:", error);
        // 跳过此文本
      }
    }
  }

  return { messages, options };
}

/**
 * 构建单个翻译模型
 *
 * @param {Bergamot} bergamot
 * @param {TranslationModelPayload} translationModelPayload
 * @returns {Bergamot["TranslationModel"]}
 */
function constructSingleTranslationModel(bergamot, translationModelPayload) {
  const { sourceLanguage, targetLanguage, languageModelFiles } =
    translationModelPayload;

  const { model, lex, vocab, qualityModel, srcvocab, trgvocab } =
    allocateModelMemory(bergamot, languageModelFiles);

  // 设置词汇表列表，可以是单个 vocab 模型，或者是 srcvocab 和 trgvocab 对
  const vocabList = new bergamot.AlignedMemoryList();

  if (vocab) {
    vocabList.push_back(vocab);
  } else if (srcvocab && trgvocab) {
    vocabList.push_back(srcvocab);
    vocabList.push_back(trgvocab);
  } else {
    throw new Error(
      `Model ${sourceLanguage}->${targetLanguage} not found. Check the model download success and the model file name is correct.`
    );
  }

  const gemmPrecision = languageModelFiles.model.record.name.endsWith(
    "intgemm8.bin"
  )
    ? "int8shiftAll"
    : "int8shiftAlphaAll";

  const config = generateTextConfig({
    "beam-size": "1",
    normalize: "1.0",
    "word-penalty": "0",
    "max-length-break": "128",
    "mini-batch-words": "1024",
    workspace: "128",
    "max-length-factor": "2.0",
    "skip-cost": (!qualityModel).toString(),
    "cpu-threads": "0",
    quiet: "true",
    "quiet-translation": "true",
    "gemm-precision": gemmPrecision,
    alignment: "soft",
  });

  Logger.debug(
    "Engine",
    `Creating translation model: ${sourceLanguage} -> ${targetLanguage}`
  );

  return new bergamot.TranslationModel(
    sourceLanguage,
    targetLanguage,
    config,
    model,
    lex ?? null,
    vocabList,
    qualityModel ?? null
  );
}

/**
 * 模型必须放置在 Bergamot wasm 模块可以访问的对齐内存中。
 * 此函数将模型二进制数据复制到这个内存空间
 *
 * @param {Bergamot} bergamot
 * @param {LanguageTranslationModelFiles} languageModelFiles
 * @returns {LanguageTranslationModelFilesAligned}
 */
function allocateModelMemory(bergamot, languageModelFiles) {
  /** @type {LanguageTranslationModelFilesAligned} */
  const results = {};

  for (const [fileType, file] of Object.entries(languageModelFiles)) {
    const alignment = MODEL_FILE_ALIGNMENTS[fileType];
    if (!alignment) {
      throw new Error(`Unknown file type: "${fileType}"`);
    }

    const alignedMemory = new bergamot.AlignedMemory(
      file.buffer.byteLength,
      alignment
    );

    alignedMemory.getByteArrayView().set(new Uint8Array(file.buffer));
    results[fileType] = alignedMemory;
  }

  return results;
}

/**
 * Engine为每个语言对创建一次。初始化过程将语言缓冲区的 ArrayBuffers，从 JS 管理的 ArrayBuffers 复制到wasm堆的对齐内部内存中
 */
class Engine {
  /**
   * 创建一个新的Engine实例
   * @param {string} sourceLanguage
   * @param {string} targetLanguage
   * @param {Bergamot} bergamot
   * @param {Array<TranslationModelPayload>} translationModelPayloads
   * @private
   */
  constructor(
    sourceLanguage,
    targetLanguage,
    bergamot,
    translationModelPayloads
  ) {
    /** @type {string} */
    this.sourceLanguage = sourceLanguage;
    /** @type {string} */
    this.targetLanguage = targetLanguage;
    /** @type {Bergamot} */
    this.bergamot = bergamot;
    /** @type {Bergamot["TranslationModel"][]} */
    this.languageTranslationModels = [];
    /** @type {Object[]} */
    this.allocatedMemoryBlocks = []; // 跟踪分配的内存块
    /** @type {number} */
    this.translationCount = 0; // 翻译计数
    /** @type {number} */
    this.lastMemoryCleanup = Date.now(); // 上次内存清理时间
    /** @type {number} */
    this.initialWasmMemorySize = bergamot.HEAP8 ? bergamot.HEAP8.length : 0;

    // 初始化翻译模型
    this._initializeTranslationModels(translationModelPayloads);

    /** @type {Bergamot["BlockingService"]} */
    this.translationService = new bergamot.BlockingService({
      cacheSize: 0, // 这里禁用缓存，交由上层翻译器负责缓存
    });

    // 预创建常用的空结果对象，避免在空输入情况下创建新对象
    this.emptyBatchResult = [];
    this.emptyResult = "";

    Logger.debug(
      "Engine",
      `Engine initialized for ${sourceLanguage} -> ${targetLanguage}`
    );
  }

  /**
   * 初始化翻译模型
   * @param {Array<TranslationModelPayload>} translationModelPayloads
   * @private
   */
  _initializeTranslationModels(translationModelPayloads) {
    try {
      for (const translationModelPayload of translationModelPayloads) {
        const model = constructSingleTranslationModel(
          this.bergamot,
          translationModelPayload
        );
        this.languageTranslationModels.push(model);
      }
    } catch (error) {
      Logger.error("Engine", "Failed to initialize translation models:", error);
      throw error;
    }
  }

  /**
   * 获取当前WebAssembly内存使用情况
   * @returns {{currentSize: number, initialSize: number, growth: number, growthMB: string}}
   */
  getWasmMemoryStats() {
    const currentSize = this.bergamot.HEAP8 ? this.bergamot.HEAP8.length : 0;
    const growth = currentSize - this.initialWasmMemorySize;
    return {
      currentSize: currentSize,
      initialSize: this.initialWasmMemorySize,
      growth: growth,
      growthMB: (growth / 1024 / 1024).toFixed(2),
    };
  }

  /**
   * 强制执行WebAssembly内存清理
   */
  forceWasmMemoryCleanup() {
    try {
      // 强制执行延迟删除的对象
      if (this.bergamot.flushPendingDeletes) {
        this.bergamot.flushPendingDeletes();
      }

      // 记录清理时间
      this.lastMemoryCleanup = Date.now();

      const stats = this.getWasmMemoryStats();
      Logger.debug(
        "Engine",
        `Memory cleanup completed - WASM size: ${stats.growthMB}MB growth`
      );
    } catch (error) {
      Logger.error("Engine", "Failed to cleanup WASM memory:", error);
    }
  }

  /**
   * 检查是否需要进行内存清理
   * @returns {boolean}
   */
  shouldCleanupMemory() {
    const timeSinceLastCleanup = Date.now() - this.lastMemoryCleanup;
    const translationsSinceCleanup =
      this.translationCount % MEMORY_CLEANUP_INTERVAL;

    // 每1000次翻译或每5分钟清理一次
    return (
      translationsSinceCleanup === 0 ||
      timeSinceLastCleanup > MEMORY_CLEANUP_TIME_THRESHOLD
    );
  }

  /**
   * 执行翻译 - 单个模型
   * @param {Bergamot["TranslationModel"]} model
   * @param {Bergamot["VectorString"]} messages
   * @param {Bergamot["VectorResponseOptions"]} options
   * @returns {Promise<any>}
   * @private
   */
  async _translateWithSingleModel(model, messages, options) {
    return new Promise((resolve, reject) => {
      try {
        const result = this.translationService.translate(
          model,
          messages,
          options
        );
        resolve(result);
      } catch (error) {
        reject(
          new Error(`Translation failed: ${error.message || "Unknown error"}`)
        );
      }
    });
  }

  /**
   * 执行翻译 - 通过中转语言
   * @param {Bergamot["TranslationModel"]} sourceModel
   * @param {Bergamot["TranslationModel"]} targetModel
   * @param {Bergamot["VectorString"]} messages
   * @param {Bergamot["VectorResponseOptions"]} options
   * @returns {Promise<any>}
   * @private
   */
  async _translateViaPivoting(sourceModel, targetModel, messages, options) {
    return new Promise((resolve, reject) => {
      try {
        const result = this.translationService.translateViaPivoting(
          sourceModel,
          targetModel,
          messages,
          options
        );
        resolve(result);
      } catch (error) {
        reject(
          new Error(
            `Pivoting translation failed: ${error.message || "Unknown error"}`
          )
        );
      }
    });
  }

  /**
   * 处理翻译结果
   * @param {any} responses
   * @returns {string[]}
   * @private
   */
  _processTranslationResults(responses) {
    const results = [];
    const responsesSize = responses.size();
    results.length = responsesSize; // 预分配数组大小

    for (let i = 0; i < responsesSize; i++) {
      try {
        results[i] = responses.get(i).getTranslatedText();
      } catch (error) {
        Logger.error(
          "Engine",
          `Failed to get translated text at index ${i}:`,
          error
        );
        results[i] = "[Translation error]";
      }
    }

    return results;
  }

  /**
   * 执行翻译的核心方法，被单文本和批量翻译共用
   *
   * @param {string|string[]} input - 输入文本或文本数组
   * @param {boolean} isHTML - 是否为HTML内容
   * @returns {Promise<string[]|string>}
   */
  async translate(input, isHTML = false) {
    // 增加翻译计数
    this.translationCount++;

    // 将单个文本转换为数组处理
    const isTextArray = Array.isArray(input);
    const sourceTexts = isTextArray ? input : [input];

    let responses;
    let messages;
    let options;

    try {
      const translationArgs = getTranslationArgs(
        this.bergamot,
        sourceTexts,
        isHTML
      );
      messages = translationArgs.messages;
      options = translationArgs.options;

      if (messages.size() === 0) {
        return isTextArray ? this.emptyBatchResult : this.emptyResult;
      }

      const modelCount = this.languageTranslationModels.length;

      if (modelCount === 1) {
        responses = await this._translateWithSingleModel(
          this.languageTranslationModels[0],
          messages,
          options
        );
      } else if (modelCount === 2) {
        responses = await this._translateViaPivoting(
          this.languageTranslationModels[0],
          this.languageTranslationModels[1],
          messages,
          options
        );
      } else {
        throw new Error("Too many models provided to the translation engine.");
      }

      // 处理翻译结果
      const results = this._processTranslationResults(responses);
      return isTextArray ? results : results[0];
    } catch (error) {
      Logger.error("Engine", "Translation process failed:", error);
      throw error;
    } finally {
      // 立即释放翻译相关的内存
      this._cleanupTranslationMemory(messages, options, responses);

      // 定期执行深度内存清理
      if (this.shouldCleanupMemory()) {
        // 使用setTimeout异步执行，避免阻塞翻译响应
        setTimeout(() => {
          this.forceWasmMemoryCleanup();
        }, 0);
      }
    }
  }

  /**
   * 清理翻译相关内存
   * @param {any} messages
   * @param {any} options
   * @param {any} responses
   * @private
   */
  _cleanupTranslationMemory(messages, options, responses) {
    try {
      messages?.delete();
      options?.delete();
      responses?.delete();
    } catch (error) {
      Logger.error("Engine", "Failed to clean up translation memory:", error);
    }
  }

  /**
   * 销毁Engine实例，释放所有资源
   */
  destroy() {
    try {
      // 删除翻译模型
      for (const model of this.languageTranslationModels) {
        if (model && typeof model.delete === "function") {
          model.delete();
        }
      }
      this.languageTranslationModels = [];

      // 删除翻译服务
      if (
        this.translationService &&
        typeof this.translationService.delete === "function"
      ) {
        this.translationService.delete();
      }

      // 强制清理所有延迟删除的对象
      this.forceWasmMemoryCleanup();

      Logger.debug(
        "Engine",
        `Engine destroyed after ${this.translationCount} translations`
      );
    } catch (error) {
      Logger.error("Engine", "Error destroying engine:", error);
    }
  }
}

module.exports = {
  Engine,
  CleanText,
  InitWasm,
};
