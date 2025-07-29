"use strict";

const Translator = require("../js/translator");
const { LogMem } = require("../js/mem");
const { performance } = require("perf_hooks");

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
    "The internet has connected people around the world like never before."
  ];

  const texts = [];
  for (let i = 0; i < count; i++) {
    const baseText = baseTexts[i % baseTexts.length];
    texts.push(`[${i + 1}] ${baseText}`);
  }
  
  return texts;
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
 * 持续翻译压力测试
 */
async function continuousStressTest(iterations = Infinity) {
  console.log("=== 持续翻译压力测试 ===\n");
  
  try {
    LogMem("测试开始");
    
    // 生成测试文本
    const testTexts = generateTestTexts(500);
    console.log(`生成了 ${testTexts.length} 个测试文本\n`);
    
    const initialMemory = monitorMemory("生成文本后");
    
    // 预加载翻译引擎
    console.log("预加载翻译引擎...");
    const engine = await Translator.Preload("en", "zh-Hans");
    const afterLoadMemory = monitorMemory("引擎加载后");
    
    // 分批翻译参数
    const batchSize = 50;
    const totalBatches = Math.ceil(testTexts.length / batchSize);
    
    console.log(`\n开始持续翻译压测，批次大小: ${batchSize}`);
    console.log("每10个批次显示一次内存状态\n");
    
    const startTime = performance.now();
    let totalTranslated = 0;
    let peakMemory = afterLoadMemory;
    let iterationCount = 0;
    let lastReportTime = startTime;
    
    // 持续运行直到达到指定迭代次数
    while (iterationCount < iterations) {
      iterationCount++;
      console.log(`\n=== 开始第 ${iterationCount} 轮翻译 ===`);
      
      for (let batchIndex = 0; batchIndex < totalBatches; batchIndex++) {
        const batchStart = batchIndex * batchSize;
        const batchEnd = Math.min(batchStart + batchSize, testTexts.length);
        const batch = testTexts.slice(batchStart, batchEnd);
        
        // 执行翻译
        await engine.translate(batch);
        totalTranslated += batch.length;
        
        // 每10个批次监控一次内存
        if ((batchIndex + 1) % 10 === 0) {
          const currentMemory = monitorMemory(`轮次 ${iterationCount}, 批次 ${batchIndex + 1}/${totalBatches}`);
          peakMemory = Math.max(peakMemory, currentMemory);
          
          const currentTime = performance.now();
          const elapsedSeconds = (currentTime - startTime) / 1000;
          const recentSeconds = (currentTime - lastReportTime) / 1000;
          lastReportTime = currentTime;
          
          console.log(`总翻译数: ${totalTranslated} 文本`);
          console.log(`总运行时间: ${elapsedSeconds.toFixed(2)}秒`);
          console.log(`最近速度: ${(batch.length * 10 / recentSeconds).toFixed(2)} 文本/秒`);
          console.log(`总体速度: ${(totalTranslated / elapsedSeconds).toFixed(2)} 文本/秒`);
          console.log(`峰值内存: ${peakMemory}MB (增长: ${(peakMemory - initialMemory).toFixed(2)}MB)`);
          console.log("---");
        }
      }
      
      // 每轮结束显示统计信息
      const roundEndMemory = monitorMemory(`第 ${iterationCount} 轮结束`);
      console.log(`第 ${iterationCount} 轮完成，本轮翻译 ${testTexts.length} 文本`);
      console.log(`总计已翻译: ${totalTranslated} 文本`);
      
      // 每5轮进行一次完整内存报告
      if (iterationCount % 5 === 0) {
        LogMem(`完成 ${iterationCount} 轮翻译`);
        
        // 手动触发垃圾回收（如果可用）
        if (global.gc) {
          global.gc();
          LogMem("垃圾回收后");
        }
        
        // 短暂暂停，让系统有机会释放资源
        console.log("暂停10秒...");
        await new Promise(resolve => setTimeout(resolve, 10000));
        monitorMemory("暂停后");
      }
    }
    
  } catch (error) {
    console.error("测试过程中发生错误:", error);
  } finally {
    // 关闭翻译器
    console.log("\n关闭翻译器...");
    await Translator.Shutdown();
    LogMem("翻译器关闭后");
    
    // 最后一次垃圾回收
    if (global.gc) {
      global.gc();
      LogMem("最终垃圾回收后");
    }
  }
}

/**
 * 高频短文本翻译测试
 */
async function highFrequencyShortTextTest(duration = 3600000) { // 默认运行1小时
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
      "Welcome"
    ];
    
    LogMem("高频测试开始");
    
    const engine = await Translator.Preload("en", "zh-Hans");
    monitorMemory("引擎加载后");
    
    console.log("\n开始高频短文本翻译测试...");
    
    const startTime = performance.now();
    let totalTranslated = 0;
    let lastReportTime = startTime;
    let reportInterval = 1000; // 每1000次翻译报告一次
    
    // 持续翻译直到达到指定时间
    while (performance.now() - startTime < duration) {
      // 随机选择1-3个短文本进行翻译
      const batchSize = Math.floor(Math.random() * 3) + 1;
      const batch = [];
      
      for (let i = 0; i < batchSize; i++) {
        const randomIndex = Math.floor(Math.random() * shortTexts.length);
        batch.push(shortTexts[randomIndex]);
      }
      
      await engine.translate(batch);
      totalTranslated += batch.length;
      
      // 定期报告状态
      if (totalTranslated % reportInterval === 0) {
        const currentTime = performance.now();
        const elapsedSeconds = (currentTime - startTime) / 1000;
        const recentSeconds = (currentTime - lastReportTime) / 1000;
        lastReportTime = currentTime;
        
        const currentMemory = monitorMemory(`已翻译 ${totalTranslated} 文本`);
        console.log(`运行时间: ${(elapsedSeconds / 60).toFixed(2)} 分钟`);
        console.log(`最近速度: ${(reportInterval / recentSeconds).toFixed(2)} 文本/秒`);
        console.log(`总体速度: ${(totalTranslated / elapsedSeconds).toFixed(2)} 文本/秒`);
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
    console.log(`平均速度: ${(totalTranslated / (totalMinutes * 60)).toFixed(2)} 文本/秒`);
    
    LogMem("高频测试结束");
    
  } catch (error) {
    console.error("高频测试失败:", error);
  }
}

// 运行测试
async function runTests() {
  try {
    // 运行持续翻译压力测试，无限循环直到手动停止
    await continuousStressTest();
    
    // 如果上面的测试结束了（通常是因为设置了有限的迭代次数），则运行高频测试
    await highFrequencyShortTextTest();
    
    console.log("\n=== 所有测试完成 ===");
    console.log("如果内存在长时间运行后保持稳定，说明内存管理正常！");
    
  } catch (error) {
    console.error("测试运行失败:", error);
  }
}

// 如果直接运行此文件，则执行测试
if (require.main === module) {
  console.log("提示：使用 --expose-gc 参数运行可以观察垃圾回收效果");
  console.log("例如: node --expose-gc example/autorelease2.example.js");
  console.log("按 Ctrl+C 可以随时停止测试\n");
  
  runTests().catch(console.error);
}

module.exports = {
  continuousStressTest,
  highFrequencyShortTextTest,
  generateTestTexts,
  monitorMemory
}; 