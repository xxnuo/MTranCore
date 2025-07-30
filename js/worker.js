// 翻译引擎 Worker
// 此文件在 Worker 线程中运行，接收翻译请求并返回结果

"use strict";

const { parentPort } = require("worker_threads");
const { performance } = require("perf_hooks");
const { Engine, CleanText, InitWasm } = require("./engine");
const { MESSAGE_TYPES } = require("./message");
const Config = require("./config");

/**
 * 日志工具函数
 * @param {...any} args - 日志参数
 */
function log(...args) {
  if (Config.LOG_LEVEL === "Info" || Config.LOG_LEVEL === "Debug") {
    console.log(`[Worker-${process.pid}]:`, ...args);
  }
}

/**
 * 错误日志工具函数
 * @param {...any} args - 错误日志参数
 */
function logError(...args) {
  console.error(`[Worker-${process.pid}]:`, ...args);
}

/**
 * 调试日志工具函数
 * @param {...any} args - 调试日志参数
 */
function logDebug(...args) {
  if (Config.LOG_LEVEL === "Debug") {
    console.log(`[Worker-${process.pid}][DEBUG]:`, ...args);
  }
}

// 全局引擎引用，用于清理
let globalEngine = null;

/**
 * 清理引擎资源
 */
function cleanupEngine() {
  if (globalEngine) {
    try {
      globalEngine.destroy();
      globalEngine = null;
      logDebug("Engine cleaned up successfully");
    } catch (error) {
      logError("Error destroying engine:", error);
    }
  }
}

// 错误处理和清理
process.on("unhandledRejection", (reason, promise) => {
  logError("Unhandled Promise Rejection at:", promise, "reason:", reason);
});

process.on("uncaughtException", (error) => {
  logError("Uncaught Exception:", error);
  cleanupEngine();
  process.exit(1);
});

process.on("exit", () => {
  logDebug("Process exiting, cleaning up resources");
  cleanupEngine();
});

process.on("SIGTERM", () => {
  logDebug("Received SIGTERM, cleaning up resources");
  cleanupEngine();
  process.exit(0);
});

process.on("SIGINT", () => {
  logDebug("Received SIGINT, cleaning up resources");
  cleanupEngine();
  process.exit(0);
});

/**
 * 工作队列类，用于管理翻译任务
 */
class WorkQueue {
  // 时间预算（毫秒）
  #TIME_BUDGET = 100;
  // 立即运行的任务数
  #RUN_IMMEDIATELY_COUNT = 20;
  // 批处理大小
  #TASK_BATCH_SIZE = 5;

  /**
   * 任务队列
   * @type {Map<number, {task: Function, resolve: Function, reject: Function, timestamp: number}>}
   */
  #tasksByTranslationId = new Map();

  #isRunning = false;
  #isWorkCancelled = false;
  #runImmediately = this.#RUN_IMMEDIATELY_COUNT;

  /**
   * 运行任务并返回结果
   * @template T
   * @param {number} translationId 翻译ID
   * @param {() => Promise<T>} task 任务函数
   * @returns {Promise<T>} 任务结果
   */
  runTask(translationId, task) {
    // 前N个任务立即运行，避免初始延迟
    if (this.#runImmediately > 0) {
      this.#runImmediately--;
      logDebug(`Running task ${translationId} immediately`);
      return Promise.resolve().then(() => task());
    }

    return new Promise((resolve, reject) => {
      const timestamp = performance.now();
      this.#tasksByTranslationId.set(translationId, {
        task,
        resolve,
        reject,
        timestamp,
      });

      logDebug(
        `Queued task ${translationId}, queue size: ${
          this.#tasksByTranslationId.size
        }`
      );

      // 启动队列处理
      this.#run().catch((error) => {
        logError("Queue processing error:", error);
        reject(error);
      });
    });
  }

  /**
   * 取消特定任务
   * @param {number} translationId 翻译ID
   */
  cancelTask(translationId) {
    const wasPresent = this.#tasksByTranslationId.delete(translationId);
    if (wasPresent) {
      logDebug(`Cancelled task ${translationId}`);
    }
  }

  /**
   * 获取队列统计信息
   * @returns {{size: number, oldestTaskAge: number}}
   */
  getStats() {
    const size = this.#tasksByTranslationId.size;
    let oldestTaskAge = 0;

    if (size > 0) {
      const now = performance.now();
      const timestamps = Array.from(this.#tasksByTranslationId.values()).map(
        (t) => t.timestamp
      );
      const oldestTimestamp = Math.min(...timestamps);
      oldestTaskAge = now - oldestTimestamp;
    }

    return { size, oldestTaskAge };
  }

  /**
   * 内部运行函数
   */
  async #run() {
    if (this.#isRunning) {
      return;
    }

    this.#isRunning = true;
    logDebug("Queue processing started");

    let lastTimeout = null;
    let tasksInBatch = 0;
    let batchCount = 0;
    let totalProcessed = 0;

    try {
      while (this.#tasksByTranslationId.size > 0) {
        if (this.#isWorkCancelled) {
          logDebug("Queue processing cancelled");
          break;
        }

        const now = performance.now();

        // 时间预算和批处理管理
        if (lastTimeout === null) {
          lastTimeout = now;
          // 让其他工作进入队列
          await this.#yield();
        } else if (
          now - lastTimeout > this.#TIME_BUDGET ||
          batchCount >= this.#TASK_BATCH_SIZE
        ) {
          // 达到时间预算或批处理大小限制，让出控制权
          await this.#yield();
          logDebug(
            `Processed batch: ${tasksInBatch} tasks, queue size: ${
              this.#tasksByTranslationId.size
            }`
          );
          lastTimeout = performance.now();
          batchCount = 0;
        }

        // 再次检查取消状态
        if (this.#isWorkCancelled || this.#tasksByTranslationId.size === 0) {
          break;
        }

        tasksInBatch++;
        batchCount++;
        totalProcessed++;

        // FIFO 队列处理
        const [translationId, taskData] = this.#tasksByTranslationId
          .entries()
          .next().value;
        const { task, resolve, reject } = taskData;
        this.#tasksByTranslationId.delete(translationId);

        try {
          logDebug(`Processing task ${translationId}`);
          const result = await task();

          // 检查取消状态
          if (this.#isWorkCancelled) {
            logDebug(`Task ${translationId} completed but queue was cancelled`);
            break;
          }

          resolve(result);
          logDebug(`Task ${translationId} completed successfully`);
        } catch (error) {
          logError(`Task ${translationId} failed:`, error);
          reject(error);
        }
      }
    } catch (error) {
      logError("Queue processing error:", error);
    } finally {
      logDebug(
        `Queue processing finished. Total processed: ${totalProcessed}, remaining: ${
          this.#tasksByTranslationId.size
        }`
      );
      this.#isRunning = false;
    }
  }

  /**
   * 让出控制权给事件循环
   */
  async #yield() {
    return new Promise((resolve) => setImmediate(resolve));
  }

  /**
   * 取消所有工作
   */
  async cancelWork() {
    logDebug(
      `Cancelling all work, queue size: ${this.#tasksByTranslationId.size}`
    );
    this.#isWorkCancelled = true;

    // 拒绝所有待处理的任务
    const tasks = Array.from(this.#tasksByTranslationId.values());
    this.#tasksByTranslationId.clear();

    for (const { reject } of tasks) {
      reject(new Error("Translation cancelled"));
    }

    // 等待当前批处理完成
    await this.#yield();
    this.#isWorkCancelled = false;

    logDebug("All work cancelled");
  }
}

/**
 * 翻译 Worker 类，封装所有 Worker 逻辑
 */
class TranslationWorker {
  constructor() {
    this.engine = null;
    this.workQueue = new WorkQueue();
    this.isInitialized = false;
    this.setupMessageHandlers();
  }

  /**
   * 设置消息处理器
   */
  setupMessageHandlers() {
    parentPort.on("message", async (data) => {
      try {
        await this.handleMessage(data);
      } catch (error) {
        logError("Message handling error:", error);
        this.sendErrorResponse(error, data.messageId);
      }
    });
  }

  /**
   * 处理来自主线程的消息
   * @param {Object} data - 消息数据
   */
  async handleMessage(data) {
    if (!data || typeof data.type !== "string") {
      throw new Error("Invalid message format");
    }

    logDebug(`Received message: ${data.type}`);

    switch (data.type) {
      case MESSAGE_TYPES.INIT_REQUEST.type:
        await this.handleInitialization(data);
        break;

      case MESSAGE_TYPES.TRANSLATION_REQUEST.type:
        if (!this.isInitialized) {
          throw new Error("Worker not initialized");
        }
        await this.handleTranslationRequest(data);
        break;

      case MESSAGE_TYPES.DISCARD_TRANSLATION_QUEUE.type:
        await this.handleDiscardQueue();
        break;

      case MESSAGE_TYPES.CANCEL_SINGLE_TRANSLATION.type:
        this.handleCancelTranslation(data);
        break;

      default:
        logError("Unknown message type:", data.type);
    }
  }

  /**
   * 处理初始化请求
   * @param {Object} data - 初始化数据
   */
  async handleInitialization(data) {
    if (this.isInitialized) {
      throw new Error("Translation engine must not be re-initialized");
    }

    const { sourceLanguage, targetLanguage, modelPayloads } = data;

    // 验证必需参数
    if (!sourceLanguage) {
      throw new Error('Worker initialization missing "sourceLanguage"');
    }
    if (!targetLanguage) {
      throw new Error('Worker initialization missing "targetLanguage"');
    }
    if (!modelPayloads) {
      throw new Error('Worker initialization missing "modelPayloads"');
    }

    try {
      const startTime = performance.now();
      log(`Initializing engine: ${sourceLanguage} -> ${targetLanguage}`);

      // 加载 WASM 模块
      const bergamot = await InitWasm();

      // 创建引擎实例
      this.engine = new Engine(
        sourceLanguage,
        targetLanguage,
        bergamot,
        modelPayloads
      );

      // 保存全局引用用于清理
      globalEngine = this.engine;
      this.isInitialized = true;

      const endTime = performance.now();
      const initTime = (endTime - startTime) / 1000;
      log(`Engine initialized successfully in ${initTime.toFixed(2)} seconds`);

      // 通知主线程初始化成功
      parentPort.postMessage(MESSAGE_TYPES.INIT_SUCCESS);
    } catch (error) {
      logError("Engine initialization failed:", error);
      parentPort.postMessage({
        type: MESSAGE_TYPES.INIT_ERROR.type,
        error: error?.message || "Unknown initialization error",
        stack: Config.LOG_LEVEL === "Debug" ? error?.stack : undefined,
      });
    }
  }

  /**
   * 处理翻译请求
   * @param {Object} data - 翻译请求数据
   */
  async handleTranslationRequest(data) {
    const { sourceText, messageId, translationId, isHTML } = data;

    // 验证必需参数
    if (typeof sourceText !== "string") {
      throw new Error("Invalid sourceText");
    }
    if (typeof translationId !== "number") {
      throw new Error("Invalid translationId");
    }

    try {
      logDebug(
        `Translation request ${translationId}: "${sourceText.substring(0, 50)}${
          sourceText.length > 50 ? "..." : ""
        }"`
      );

      // 清理文本
      const { whitespaceBefore, whitespaceAfter, cleanedSourceText } =
        CleanText(this.engine.sourceLanguage, sourceText);

      const startTime = performance.now();

      // 执行翻译
      const result = await this.workQueue.runTask(translationId, async () => {
        return this.engine.translate([cleanedSourceText], isHTML);
      });

      const endTime = performance.now();
      const inferenceMilliseconds = endTime - startTime;

      // 恢复空白字符
      const targetText = whitespaceBefore + result[0] + whitespaceAfter;

      logDebug(
        `Translation ${translationId} completed in ${inferenceMilliseconds.toFixed(
          2
        )}ms`
      );

      // 返回结果
      parentPort.postMessage({
        type: MESSAGE_TYPES.TRANSLATION_RESPONSE.type,
        targetText,
        inferenceMilliseconds,
        translationId,
        messageId,
      });
    } catch (error) {
      logError(`Translation ${translationId} error:`, error);
      parentPort.postMessage({
        type: MESSAGE_TYPES.TRANSLATION_ERROR.type,
        error: {
          message: error?.message || "Unknown translation error",
          stack: Config.LOG_LEVEL === "Debug" ? error?.stack : undefined,
        },
        messageId,
        translationId,
      });
    }
  }

  /**
   * 处理丢弃队列请求
   */
  async handleDiscardQueue() {
    const stats = this.workQueue.getStats();
    log(
      `Discarding translation queue (${
        stats.size
      } tasks, oldest: ${stats.oldestTaskAge.toFixed(2)}ms)`
    );

    await this.workQueue.cancelWork();
    parentPort.postMessage(MESSAGE_TYPES.TRANSLATIONS_DISCARDED);
  }

  /**
   * 处理取消单个翻译请求
   * @param {Object} data - 取消请求数据
   */
  handleCancelTranslation(data) {
    const { translationId } = data;
    if (typeof translationId !== "number") {
      logError("Invalid translationId for cancellation");
      return;
    }

    log(`Cancelling translation ${translationId}`);
    this.workQueue.cancelTask(translationId);
  }

  /**
   * 发送错误响应
   * @param {Error} error - 错误对象
   * @param {string} [messageId] - 消息ID
   */
  sendErrorResponse(error, messageId) {
    parentPort.postMessage({
      type: MESSAGE_TYPES.TRANSLATION_ERROR.type,
      error: {
        message: error?.message || "Unknown error",
        stack: Config.LOG_LEVEL === "Debug" ? error?.stack : undefined,
      },
      messageId,
    });
  }
}

// 创建 Worker 实例
const worker = new TranslationWorker();

// 通知主线程 Worker 已准备好接收初始化消息
parentPort.postMessage(MESSAGE_TYPES.WORKER_READY);
log("Translation worker ready");
