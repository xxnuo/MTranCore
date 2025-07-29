"use strict";

const Translator = require("../js/translator");
const { LogMem } = require("../js/mem");
const { performance } = require("perf_hooks");

// 设置环境变量以启用调试日志
process.env.MTRAN_LOG_LEVEL = "Info";
process.env.MTRAN_WORKER_INIT_TIMEOUT = "60000"; // 降低超时时间到1分钟
process.env.MTRAN_TRANSLATION_TIMEOUT = "30000"; // 设置翻译超时为30秒

/**
 * 生成测试文本
 */
function generateTestTexts(count = 1000) {
  const baseTexts = [
    "Hello world, this is a test message for translation.",
    "Machine translation has made significant progress in recent years.",
    "Natural language processing is a fascinating field of study.",
    "The quick brown fox jumps over the lazy dog.",
    "Artificial intelligence is transforming the way we work and live.",
    "Deep learning models have revolutionized computer vision and NLP.",
    "Open source software enables collaboration and innovation.",
    "Cloud computing provides scalable and flexible infrastructure.",
    "Data science combines statistics, programming, and domain expertise.",
    "The internet has connected people around the world like never before.",
  ];

  const texts = [];
  for (let i = 0; i < count; i++) {
    const baseText = baseTexts[i % baseTexts.length];
    texts.push(`[${i + 1}] ${baseText}`);
  }

  return texts;
}

/**
 * 监控系统资源使用情况
 */
function monitorSystemResources(label) {
  const memUsage = process.memoryUsage();
  const cpuUsage = process.cpuUsage();
  
  // 转换为MB
  const heapUsedMB = (memUsage.heapUsed / 1024 / 1024).toFixed(2);
  const rssMB = (memUsage.rss / 1024 / 1024).toFixed(2);
  const externalMB = (memUsage.external / 1024 / 1024).toFixed(2);
  
  // CPU 使用时间（微秒）
  const userCPU = (cpuUsage.user / 1000).toFixed(2); // 转换为毫秒
  const systemCPU = (cpuUsage.system / 1000).toFixed(2);
  
  console.log(`[${label}] 资源状态:`);
  console.log(`  堆内存: ${heapUsedMB}MB, RSS: ${rssMB}MB, 外部: ${externalMB}MB`);
  console.log(`  CPU时间: 用户 ${userCPU}ms, 系统 ${systemCPU}ms`);
  
  return {
    heapUsedMB: parseFloat(heapUsedMB),
    rssMB: parseFloat(rssMB),
    externalMB: parseFloat(externalMB),
    userCPU: parseFloat(userCPU),
    systemCPU: parseFloat(systemCPU)
  };
}

/**
 * 监控内存使用情况
 */
function monitorMemory(label) {
  const memUsage = process.memoryUsage();
  const heapUsedMB = (memUsage.heapUsed / 1024 / 1024).toFixed(2);
  const rssMB = (memUsage.rss / 1024 / 1024).toFixed(2);
  console.log(`[${label}] 堆内存: ${heapUsedMB}MB, RSS: ${rssMB}MB`);
  return parseFloat(heapUsedMB);
}

/**
 * 带超时的翻译函数 - 改进版
 */
async function translateWithTimeout(engine, texts, timeout = 60000) {
  // 增加默认超时时间到60秒
  return new Promise(async (resolve, reject) => {
    const timer = setTimeout(() => {
      reject(new Error(`翻译超时 (${timeout}ms) - 可能 Worker 进程卡住`));
    }, timeout);

    try {
      const startTime = performance.now();
      
      // 使用 Promise.race 来确保即使引擎内部卡住也能超时
      const result = await Promise.race([
        engine.translate(texts),
        new Promise((_, timeoutReject) => 
          setTimeout(() => timeoutReject(new Error('内部翻译超时')), timeout - 1000)
        )
      ]);
      
      const duration = (performance.now() - startTime) / 1000;
      clearTimeout(timer);

      // 如果翻译时间超过10秒，记录警告
      if (duration > 10) {
        console.log(
          `   ⚠️  翻译耗时较长: ${duration.toFixed(2)}秒 (${
            texts.length
          }个文本)`
        );
      }

      resolve(result);
    } catch (error) {
      clearTimeout(timer);
      reject(error);
    }
  });
}

/**
 * 检测 Worker 是否响应
 */
async function checkWorkerHealth(engine) {
  console.log("   🔍 检查 Worker 健康状态...");
  
  try {
    // 发送一个简单的测试翻译
    const testStart = performance.now();
    const result = await Promise.race([
      engine.translate(["test"]),
      new Promise((_, reject) => 
        setTimeout(() => reject(new Error('Worker 健康检查超时')), 5000)
      )
    ]);
    const testDuration = (performance.now() - testStart) / 1000;
    
    console.log(`   ✓ Worker 响应正常，耗时: ${testDuration.toFixed(2)}秒`);
    return true;
  } catch (error) {
    console.log(`   ✗ Worker 健康检查失败: ${error.message}`);
    return false;
  }
}

/**
 * 重启翻译引擎
 */
async function restartEngine(fromLang, toLang) {
  console.log("   🔄 尝试重启翻译引擎...");
  
  try {
    // 关闭当前引擎
    await Translator.Shutdown();
    console.log("   ✓ 旧引擎已关闭");
    
    // 等待一段时间让资源释放
    await new Promise(resolve => setTimeout(resolve, 2000));
    
    // 重新加载引擎
    const newEngine = await Promise.race([
      Translator.Preload(fromLang, toLang),
      new Promise((_, reject) => 
        setTimeout(() => reject(new Error('引擎重启超时')), 60000)
      )
    ]);
    
    console.log("   ✓ 新引擎加载成功");
    return newEngine;
  } catch (error) {
    console.log(`   ✗ 引擎重启失败: ${error.message}`);
    throw error;
  }
}

/**
 * 检查翻译引擎是否可用
 */
async function checkEngineAvailability() {
  console.log("=== 检查翻译引擎可用性 ===\n");

  try {
    // 检查支持的语言
    const supportedLanguages = Translator.GetSupportLanguages();
    console.log(
      "支持的语言:",
      supportedLanguages.slice(0, 10),
      "...(共",
      supportedLanguages.length,
      "种)"
    );

    // 检查是否支持 en -> zh-Hans
    if (
      !supportedLanguages.includes("en") ||
      !supportedLanguages.includes("zh-Hans")
    ) {
      throw new Error("不支持英文到中文的翻译");
    }

    console.log("✓ 语言支持检查通过\n");
    return true;
  } catch (error) {
    console.error("✗ 引擎可用性检查失败:", error.message);
    return false;
  }
}

/**
 * 持续翻译压力测试
 */
async function continuousStressTest(iterations = Infinity) {
  console.log("=== 持续翻译压力测试 ===\n");

  try {
    // 检查引擎可用性
    const isAvailable = await checkEngineAvailability();
    if (!isAvailable) {
      console.log("引擎不可用，跳过测试");
      return;
    }

    LogMem("测试开始");

    // 生成测试文本
    const testTexts = generateTestTexts(500);
    console.log(`生成了 ${testTexts.length} 个测试文本\n`);

    const initialMemory = monitorMemory("生成文本后");

    // 预加载翻译引擎
    console.log("预加载翻译引擎...");
    console.log("注意：首次加载可能需要下载模型文件，请耐心等待...");

    let engine;
    try {
      const loadStartTime = performance.now();
      engine = await Promise.race([
        Translator.Preload("en", "zh-Hans"),
        new Promise(
          (_, reject) =>
            setTimeout(() => reject(new Error("引擎加载超时")), 120000) // 2分钟超时
        ),
      ]);
      const loadTime = (performance.now() - loadStartTime) / 1000;
      console.log(`引擎加载完成，耗时: ${loadTime.toFixed(2)}秒`);
    } catch (error) {
      console.error("引擎加载失败:", error.message);
      console.log("可能的原因:");
      console.log("1. 网络连接问题，无法下载模型文件");
      console.log("2. 模型文件损坏或缺失");
      console.log("3. Worker 初始化失败");
      console.log("\n建议:");
      console.log("1. 检查网络连接");
      console.log("2. 删除 ~/.cache/mtran 目录重新下载模型");
      console.log("3. 使用 MTRAN_LOG_LEVEL=Debug 查看详细日志");
      return;
    }

    const afterLoadMemory = monitorMemory("引擎加载后");

    // 测试基本翻译功能
    console.log("\n测试基本翻译功能...");
    try {
      const testResult = await translateWithTimeout(
        engine,
        ["Hello world"],
        10000
      );
      console.log(`测试翻译结果: "${testResult[0]}"`);
      console.log("✓ 基本翻译功能正常\n");
    } catch (error) {
      console.error("✗ 基本翻译测试失败:", error.message);
      return;
    }

    // 分批翻译参数
    const batchSize = 20; // 减少批量大小从50到20
    const totalBatches = Math.ceil(testTexts.length / batchSize);

    console.log(`开始持续翻译压测，批次大小: ${batchSize}`);
    console.log("每5个批次显示一次内存状态\n"); // 改为每5个批次

    const startTime = performance.now();
    let totalTranslated = 0;
    let peakMemory = afterLoadMemory;
    let iterationCount = 0;
    let lastReportTime = startTime;
    let consecutiveErrors = 0;

    // 持续运行直到达到指定迭代次数
    while (iterationCount < iterations) {
      iterationCount++;
      console.log(`\n=== 开始第 ${iterationCount} 轮翻译 ===`);

      let batchErrors = 0;

      for (let batchIndex = 0; batchIndex < totalBatches; batchIndex++) {
        const batchStart = batchIndex * batchSize;
        const batchEnd = Math.min(batchStart + batchSize, testTexts.length);
        const batch = testTexts.slice(batchStart, batchEnd);

        console.log(
          `   处理批次 ${batchIndex + 1}/${totalBatches} (${
            batch.length
          }个文本)...`
        );

        try {
          // 执行翻译，增加超时时间到60秒
          await translateWithTimeout(engine, batch, 60000);
          totalTranslated += batch.length;
          consecutiveErrors = 0; // 重置连续错误计数
          console.log(`   ✓ 批次 ${batchIndex + 1} 完成`);
        } catch (error) {
          batchErrors++;
          consecutiveErrors++;
          console.error(`   ✗ 批次 ${batchIndex + 1} 翻译失败:`, error.message);

          // 如果是超时错误，进行更详细的诊断
          if (error.message.includes('超时') || error.message.includes('timeout')) {
            console.log("   🔍 检测到超时错误，进行诊断...");
            
            // 检查 Worker 健康状态
            const isHealthy = await checkWorkerHealth(engine);
            if (!isHealthy) {
              console.log("   ⚠️  Worker 可能已经卡住，尝试重启引擎...");
              try {
                engine = await restartEngine("en", "zh-Hans");
                console.log("   ✓ 引擎重启成功，继续测试");
                consecutiveErrors = 0; // 重置错误计数
              } catch (restartError) {
                console.error("   ✗ 引擎重启失败，停止测试");
                return;
              }
            }
          }

          // 如果连续错误太多，停止测试
          if (consecutiveErrors >= 3) {
            console.error("连续翻译错误过多，停止测试");
            return;
          }

          // 短暂等待后继续
          console.log(`   等待2秒后重试...`);
          await new Promise((resolve) => setTimeout(resolve, 2000));
        }

        // 每5个批次监控一次内存
        if ((batchIndex + 1) % 5 === 0) {
          const currentMemory = monitorMemory(
            `轮次 ${iterationCount}, 批次 ${batchIndex + 1}/${totalBatches}`
          );
          peakMemory = Math.max(peakMemory, currentMemory);

          const currentTime = performance.now();
          const elapsedSeconds = (currentTime - startTime) / 1000;
          const recentSeconds = (currentTime - lastReportTime) / 1000;
          lastReportTime = currentTime;

          console.log(`总翻译数: ${totalTranslated} 文本`);
          console.log(`总运行时间: ${elapsedSeconds.toFixed(2)}秒`);
          console.log(
            `最近速度: ${((batch.length * 5) / recentSeconds).toFixed(
              2
            )} 文本/秒`
          );
          console.log(
            `总体速度: ${(totalTranslated / elapsedSeconds).toFixed(2)} 文本/秒`
          );
          console.log(
            `峰值内存: ${peakMemory}MB (增长: ${(
              peakMemory - initialMemory
            ).toFixed(2)}MB)`
          );
          if (batchErrors > 0) {
            console.log(`本轮批次错误: ${batchErrors}`);
          }
          console.log("---");
        }
      }

      // 每轮结束显示统计信息
      const roundEndMemory = monitorMemory(`第 ${iterationCount} 轮结束`);
      console.log(
        `第 ${iterationCount} 轮完成，本轮翻译 ${testTexts.length} 文本`
      );
      console.log(`总计已翻译: ${totalTranslated} 文本`);
      if (batchErrors > 0) {
        console.log(`本轮错误数: ${batchErrors}`);
      }

      // 每5轮进行一次完整内存报告
      if (iterationCount % 5 === 0) {
        LogMem(`完成 ${iterationCount} 轮翻译`);

        // 手动触发垃圾回收（如果可用）
        if (global.gc) {
          console.log("执行垃圾回收...");
          global.gc();
          LogMem("垃圾回收后");
        }

        // 短暂暂停，让系统有机会释放资源
        console.log("暂停10秒...");
        await new Promise((resolve) => setTimeout(resolve, 10000));
        monitorMemory("暂停后");
      }
    }
  } catch (error) {
    console.error("测试过程中发生错误:", error);
    console.error("错误堆栈:", error.stack);
  } finally {
    // 关闭翻译器
    console.log("\n关闭翻译器...");
    try {
      await Translator.Shutdown();
      LogMem("翻译器关闭后");
    } catch (error) {
      console.error("关闭翻译器时出错:", error.message);
    }

    // 最后一次垃圾回收
    if (global.gc) {
      console.log("执行最终垃圾回收...");
      global.gc();
      LogMem("最终垃圾回收后");
    }
  }
}

/**
 * 高频短文本翻译测试
 */
async function highFrequencyShortTextTest(duration = 3600000) {
  // 默认运行1小时
  console.log("\n=== 高频短文本翻译测试 ===\n");

  try {
    // 生成短文本
    const shortTexts = [
      "Hello world",
      "How are you?",
      "Good morning",
      "Thank you",
      "What's your name?",
      "Nice to meet you",
      "Goodbye",
      "See you later",
      "Have a nice day",
      "Welcome",
    ];

    LogMem("高频测试开始");

    console.log("预加载翻译引擎...");
    const engine = await Translator.Preload("en", "zh-Hans");
    monitorMemory("引擎加载后");

    console.log("\n开始高频短文本翻译测试...");

    const startTime = performance.now();
    let totalTranslated = 0;
    let lastReportTime = startTime;
    let reportInterval = 1000; // 每1000次翻译报告一次
    let consecutiveErrors = 0;

    // 持续翻译直到达到指定时间
    while (performance.now() - startTime < duration) {
      // 随机选择1-3个短文本进行翻译
      const batchSize = Math.floor(Math.random() * 3) + 1;
      const batch = [];

      for (let i = 0; i < batchSize; i++) {
        const randomIndex = Math.floor(Math.random() * shortTexts.length);
        batch.push(shortTexts[randomIndex]);
      }

      try {
        await translateWithTimeout(engine, batch, 20000); // 增加超时时间到20秒
        totalTranslated += batch.length;
        consecutiveErrors = 0;
      } catch (error) {
        consecutiveErrors++;
        console.error(`高频翻译错误:`, error.message);

        if (consecutiveErrors >= 5) {
          // 减少连续错误阈值
          console.error("连续错误过多，停止高频测试");
          break;
        }

        // 短暂等待
        await new Promise((resolve) => setTimeout(resolve, 500)); // 减少等待时间
      }

      // 定期报告状态
      if (totalTranslated % reportInterval === 0) {
        const currentTime = performance.now();
        const elapsedSeconds = (currentTime - startTime) / 1000;
        const recentSeconds = (currentTime - lastReportTime) / 1000;
        lastReportTime = currentTime;

        const currentMemory = monitorMemory(`已翻译 ${totalTranslated} 文本`);
        console.log(`运行时间: ${(elapsedSeconds / 60).toFixed(2)} 分钟`);
        console.log(
          `最近速度: ${(reportInterval / recentSeconds).toFixed(2)} 文本/秒`
        );
        console.log(
          `总体速度: ${(totalTranslated / elapsedSeconds).toFixed(2)} 文本/秒`
        );
        console.log("---");

        // 每10000次翻译进行一次完整内存报告
        if (totalTranslated % (reportInterval * 10) === 0) {
          LogMem(`完成 ${totalTranslated} 次翻译`);

          // 手动触发垃圾回收（如果可用）
          if (global.gc) {
            global.gc();
            LogMem("垃圾回收后");
          }
        }
      }
    }

    const endTime = performance.now();
    const totalMinutes = (endTime - startTime) / 60000;

    console.log(`\n高频测试完成，总运行时间: ${totalMinutes.toFixed(2)} 分钟`);
    console.log(`总翻译数: ${totalTranslated} 文本`);
    console.log(
      `平均速度: ${(totalTranslated / (totalMinutes * 60)).toFixed(2)} 文本/秒`
    );

    LogMem("高频测试结束");
  } catch (error) {
    console.error("高频测试失败:", error);
    console.error("错误堆栈:", error.stack);
  }
}

// 运行测试
async function runTests() {
  try {
    console.log("MTranCore 内存管理压力测试");
    console.log("==============================");
    console.log(`Node.js 版本: ${process.version}`);
    console.log(`平台: ${process.platform} ${process.arch}`);
    console.log(`PID: ${process.pid}`);
    console.log(`启动参数: ${process.argv.join(" ")}`);
    console.log("==============================\n");

    // 运行持续翻译压力测试，无限循环直到手动停止
    await continuousStressTest();

    // 如果上面的测试结束了（通常是因为设置了有限的迭代次数），则运行高频测试
    await highFrequencyShortTextTest();

    console.log("\n=== 所有测试完成 ===");
    console.log("如果内存在长时间运行后保持稳定，说明内存管理正常！");
  } catch (error) {
    console.error("测试运行失败:", error);
    console.error("错误堆栈:", error.stack);
  }
}

// 优雅退出处理
process.on("SIGINT", async () => {
  console.log("\n\n收到中断信号，正在优雅退出...");
  try {
    await Translator.Shutdown();
    console.log("翻译器已关闭");
  } catch (error) {
    console.error("关闭翻译器时出错:", error.message);
  }
  process.exit(0);
});

process.on("uncaughtException", (error) => {
  console.error("未捕获的异常:", error);
  console.error("错误堆栈:", error.stack);
  process.exit(1);
});

process.on("unhandledRejection", (reason, promise) => {
  console.error("未处理的 Promise 拒绝:", reason);
  console.error("Promise:", promise);
  process.exit(1);
});

// 如果直接运行此文件，则执行测试
if (require.main === module) {
  console.log("提示：使用 --expose-gc 参数运行可以观察垃圾回收效果");
  console.log("例如: node --expose-gc example/memory.gc2.example.js");
  console.log("按 Ctrl+C 可以随时停止测试\n");

  runTests().catch(console.error);
}

module.exports = {
  continuousStressTest,
  highFrequencyShortTextTest,
  generateTestTexts,
  monitorMemory,
  monitorSystemResources,
  checkEngineAvailability,
  translateWithTimeout,
  checkWorkerHealth,
  restartEngine,
};
