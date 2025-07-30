// 翻译引擎 Worker
// 此文件在 Worker 线程中运行，接收翻译请求并返回结果

"use strict";

const { parentPort } = require("worker_threads");
const { performance } = require("perf_hooks");
const { Engine, CleanText, InitWasm } = require("./engine");
const { MESSAGE_TYPES } = require("./message");
const Config = require("./config")

function log(...args) {
  if (Config.LOG_LEVEL === "Info" || Config.LOG_LEVEL === "Debug") {
    console.log("Worker:", ...args);
  }
}

// 全局engine引用，用于清理
let globalEngine = null;

// 捕获未处理的 Promise 拒绝
process.on("unhandledRejection", (reason) => {
  console.error("Unhandled Rejection:", reason);
});

// 处理worker退出时的清理
process.on("exit", () => {
  if (globalEngine) {
    try {
      globalEngine.destroy();
      globalEngine = null;
    } catch (error) {
      console.error("Error destroying engine on exit:", error);
    }
  }
});

// 处理SIGTERM信号
process.on("SIGTERM", () => {
  if (globalEngine) {
    try {
      globalEngine.destroy();
      globalEngine = null;
    } catch (error) {
      console.error("Error destroying engine on SIGTERM:", error);
    }
  }
  process.exit(0);
});

/**
 * 初始化引擎并处理消息
 */
async function handleInitializationMessage(data) {
  if (data.type !== MESSAGE_TYPES.INIT_REQUEST.type) {
    console.error(
      "Translation worker received a message before initialization"
    );
    return;
  }

  try {
    // const { sourceLanguage, targetLanguage, modelPayloads, logLevel } = data;
    const { sourceLanguage, targetLanguage, modelPayloads } = data;

    if (!sourceLanguage) {
      throw new Error('Worker initialization missing "sourceLanguage"');
    }
    if (!targetLanguage) {
      throw new Error('Worker initialization missing "targetLanguage"');
    }

    // 初始化引擎
    const startTime = performance.now();
    log(`Initializing engine for ${sourceLanguage} -> ${targetLanguage}`);

    // 加载 WASM 模块
    const bergamot = await InitWasm();

    // 创建引擎实例
    const engine = new Engine(
      sourceLanguage,
      targetLanguage,
      bergamot,
      modelPayloads
    );
    
    // 保存全局引用用于清理
    globalEngine = engine;

    const endTime = performance.now();
    log(`Engine initialized in ${(endTime - startTime) / 1000} seconds`);

    // 设置消息处理器
    handleMessages(engine);

    // 通知主线程初始化成功
    parentPort.postMessage(MESSAGE_TYPES.INIT_SUCCESS);
  } catch (error) {
    console.error("Engine initialization error:", error);
    parentPort.postMessage({
      type: MESSAGE_TYPES.INIT_ERROR.type,
      error: error?.message || "Unknown error during initialization",
    });
  }
}

/**
 * 处理来自主线程的消息
 */
function handleMessages(engine) {
  // 工作队列管理
  const workQueue = new WorkQueue();

  // 监听来自主线程的消息
  parentPort.on("message", async (data) => {
    try {
      if (data.type === MESSAGE_TYPES.INIT_REQUEST.type) {
        throw new Error("The translation engine must not be re-initialized.");
      }

      switch (data.type) {
        case MESSAGE_TYPES.TRANSLATION_REQUEST.type: {
          const { sourceText, messageId, translationId, isHTML } = data;

          try {
            // 清理文本
            const { whitespaceBefore, whitespaceAfter, cleanedSourceText } =
              CleanText(engine.sourceLanguage, sourceText);

            // 添加到工作队列
            const startTime = performance.now();

            // 执行翻译
            const result = await workQueue.runTask(translationId, async () => {
              return engine.translate([cleanedSourceText], isHTML);
            });

            const endTime = performance.now();
            const inferenceMilliseconds = endTime - startTime;

            // 恢复空白
            const targetText = whitespaceBefore + result[0] + whitespaceAfter;

            // 返回结果
            parentPort.postMessage({
              type: MESSAGE_TYPES.TRANSLATION_RESPONSE.type,
              targetText,
              inferenceMilliseconds,
              translationId,
              messageId,
            });
          } catch (error) {
            console.error("Translation error:", error);
            parentPort.postMessage({
              type: MESSAGE_TYPES.TRANSLATION_ERROR.type,
              error: {
                message: error?.message || "Unknown translation error",
                stack: error?.stack || "(no stack)",
              },
              messageId,
            });
          }
          break;
        }

        case MESSAGE_TYPES.DISCARD_TRANSLATION_QUEUE.type: {
          log("Discarding translation queue");
          await workQueue.cancelWork();
          parentPort.postMessage(MESSAGE_TYPES.TRANSLATIONS_DISCARDED);
          break;
        }

        case MESSAGE_TYPES.CANCEL_SINGLE_TRANSLATION.type: {
          const { translationId } = data;
          log(`Cancelling translation ${translationId}`);
          workQueue.cancelTask(translationId);
          break;
        }

        default:
          console.warn("Unknown message type:", data.type);
      }
    } catch (error) {
      console.error("Message handling error:", error);
    }
  });
}

/**
 * 工作队列类，用于管理翻译任务
 */
class WorkQueue {
  // 时间预算（毫秒）
  #TIME_BUDGET = 100;
  // 立即运行的任务数
  #RUN_IMMEDIATELY_COUNT = 20;

  /**
   * 任务队列
   * @type {Map<number, {task: Function, resolve: Function, reject: Function}>}
   */
  #tasksByTranslationId = new Map();

  #isRunning = false;
  #isWorkCancelled = false;
  #runImmediately = this.#RUN_IMMEDIATELY_COUNT;
  #taskBatchSize = 5; // 批处理大小

  /**
   * 运行任务并返回结果
   * @template T
   * @param {number} translationId 翻译ID
   * @param {() => T} task 任务函数
   * @returns {Promise<T>} 任务结果
   */
  runTask(translationId, task) {
    if (this.#runImmediately > 0) {
      // 前N个任务立即运行
      this.#runImmediately--;
      return Promise.resolve(task());
    }

    return new Promise((resolve, reject) => {
      this.#tasksByTranslationId.set(translationId, { task, resolve, reject });
      this.#run().catch((error) => console.error(error));
    });
  }

  /**
   * 取消特定任务
   * @param {number} translationId 翻译ID
   */
  cancelTask(translationId) {
    this.#tasksByTranslationId.delete(translationId);
  }

  /**
   * 内部运行函数
   */
  async #run() {
    if (this.#isRunning) {
      return;
    }

    this.#isRunning = true;

    let lastTimeout = null;
    let tasksInBatch = 0;
    let batchCount = 0;

    while (this.#tasksByTranslationId.size) {
      if (this.#isWorkCancelled) {
        break;
      }

      const now = performance.now();

      if (lastTimeout === null) {
        lastTimeout = now;
        // 让其他工作进入队列
        await new Promise((resolve) => setTimeout(resolve, 0));
      } else if (
        now - lastTimeout > this.#TIME_BUDGET ||
        batchCount >= this.#taskBatchSize
      ) {
        // 无等待超时，清除当前事件循环中的promise队列
        await new Promise((resolve) => setTimeout(resolve, 0));
        log(`Processed ${tasksInBatch} tasks in batch`);
        lastTimeout = performance.now();
        batchCount = 0;
      }

      // 每个await之间检查
      if (this.#isWorkCancelled || !this.#tasksByTranslationId.size) {
        break;
      }

      tasksInBatch++;
      batchCount++;

      // 将Map作为FIFO队列处理，取出最早的项目
      const [translationId, taskAndResolvers] = this.#tasksByTranslationId
        .entries()
        .next().value;
      const { task, resolve, reject } = taskAndResolvers;
      this.#tasksByTranslationId.delete(translationId);

      try {
        const result = await task();

        // 每个await之间检查
        if (this.#isWorkCancelled) {
          break;
        }

        // 工作完成，解析原始任务
        resolve(result);
      } catch (error) {
        reject(error);
      }
    }

    log(`Finished processing ${tasksInBatch} tasks`);
    this.#isRunning = false;
  }

  /**
   * 取消所有工作
   */
  async cancelWork() {
    this.#isWorkCancelled = true;
    this.#tasksByTranslationId.clear(); // 使用clear()代替重新分配
    await new Promise((resolve) => setTimeout(resolve, 0));
    this.#isWorkCancelled = false;
  }
}

// 监听主线程的初始化消息
parentPort.on("message", (data) => {
  if (data.type === MESSAGE_TYPES.INIT_REQUEST.type) {
    handleInitializationMessage(data);
  }
});

// 通知主线程worker已准备好接收初始化消息
parentPort.postMessage(MESSAGE_TYPES.WORKER_READY);
