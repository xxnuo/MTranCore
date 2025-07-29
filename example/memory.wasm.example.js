"use strict";

const Translator = require("../js/translator");
const { LogMem } = require("../js/mem");
const { performance } = require("perf_hooks");

/**
 * 生成测试文本
 */
function generateTestTexts(count = 50) {
  const baseTexts = [
    "Hello world, this is a comprehensive test for WebAssembly memory management.",
    "Machine translation technology has evolved significantly with deep learning advances.",
    "Natural language processing requires careful attention to memory optimization.",
    "The quick brown fox jumps over the lazy dog in multiple languages.",
    "Artificial intelligence systems must handle memory efficiently for production use."
  ];

  const texts = [];
  for (let i = 0; i < count; i++) {
    const baseText = baseTexts[i % baseTexts.length];
    texts.push(`[${i + 1}] ${baseText} Additional content to increase memory usage.`);
  }
  
  return texts;
}

/**
 * 详细监控内存使用情况
 */
function detailedMemoryMonitor(label) {
  const memUsage = process.memoryUsage();
  const heapUsedMB = (memUsage.heapUsed / 1024 / 1024).toFixed(2);
  const heapTotalMB = (memUsage.heapTotal / 1024 / 1024).toFixed(2);
  const rssMB = (memUsage.rss / 1024 / 1024).toFixed(2);
  const externalMB = (memUsage.external / 1024 / 1024).toFixed(2);
  
  console.log(`[${label}]`);
  console.log(`  堆已用: ${heapUsedMB}MB | 堆总计: ${heapTotalMB}MB`);
  console.log(`  RSS: ${rssMB}MB | 外部: ${externalMB}MB`);
  
  return {
    heapUsed: parseFloat(heapUsedMB),
    heapTotal: parseFloat(heapTotalMB),
    rss: parseFloat(rssMB),
    external: parseFloat(externalMB)
  };
}

/**
 * WebAssembly底层内存管理测试
 */
async function testWasmMemoryManagement() {
  console.log("=== WebAssembly底层内存管理测试 ===\n");
  
  // 启用详细日志
  process.env.MTRAN_LOG_LEVEL = "Debug";
  
  try {
    LogMem("测试开始");
    
    const testTexts = generateTestTexts(30);
    console.log(`生成了 ${testTexts.length} 个测试文本\n`);
    
    const initialMemory = detailedMemoryMonitor("初始状态");
    
    // 预加载翻译引擎
    console.log("预加载翻译引擎...");
    const engine = await Translator.Preload("en", "zh-Hans");
    const afterLoadMemory = detailedMemoryMonitor("引擎加载后");
    
    console.log(`\n开始WebAssembly内存管理测试`);
    console.log("每500次翻译显示详细内存状态\n");
    
    const startTime = performance.now();
    let totalTranslated = 0;
    let memoryReadings = [];
    let maxRssGrowth = 0;
    let memoryLeakDetected = false;
    
    // 进行密集翻译测试，观察内存管理效果
    for (let round = 0; round < 200; round++) { // 总共6000次翻译
      // 每轮翻译30个文本
      await engine.translate(testTexts);
      totalTranslated += testTexts.length;
      
      // 每500次翻译详细监控内存
      if (totalTranslated % 500 === 0) {
        const currentMemory = detailedMemoryMonitor(`已翻译 ${totalTranslated} 次`);
        
        memoryReadings.push({
          count: totalTranslated,
          ...currentMemory
        });
        
        // 分析内存变化
        if (memoryReadings.length >= 2) {
          const prev = memoryReadings[memoryReadings.length - 2];
          const current = memoryReadings[memoryReadings.length - 1];
          
          const rssDiff = current.rss - prev.rss;
          const heapDiff = current.heapUsed - prev.heapUsed;
          const externalDiff = current.external - prev.external;
          
          console.log(`  变化: RSS${rssDiff >= 0 ? '+' : ''}${rssDiff.toFixed(2)}MB | 堆${heapDiff >= 0 ? '+' : ''}${heapDiff.toFixed(2)}MB | 外部${externalDiff >= 0 ? '+' : ''}${externalDiff.toFixed(2)}MB`);
          
          // 检测内存释放
          if (rssDiff < -5) {
            console.log(`  ✅ 检测到内存释放: RSS减少${Math.abs(rssDiff).toFixed(2)}MB`);
          } else if (rssDiff > 20) {
            console.log(`  ⚠️  内存增长较大: RSS增加${rssDiff.toFixed(2)}MB`);
            maxRssGrowth = Math.max(maxRssGrowth, rssDiff);
          }
          
          // 检测潜在内存泄漏
          if (rssDiff > 15 && memoryReadings.length >= 4) {
            const recent4 = memoryReadings.slice(-4);
            const allIncreasing = recent4.every((reading, i) => 
              i === 0 || reading.rss > recent4[i-1].rss
            );
            if (allIncreasing) {
              memoryLeakDetected = true;
              console.log(`  🚨 可能的内存泄漏: 连续4次测量RSS都在增长`);
            }
          }
        }
        
        console.log("---\n");
        
        // 手动触发垃圾回收（如果可用）
        if (global.gc && totalTranslated % 1000 === 0) {
          global.gc();
          console.log("  🔄 执行了垃圾回收");
        }
      }
    }
    
    const endTime = performance.now();
    const totalTime = endTime - startTime;
    const finalMemory = detailedMemoryMonitor("翻译完成");
    
    console.log("\n=== 详细测试结果分析 ===");
    console.log(`总翻译数: ${totalTranslated}`);
    console.log(`总耗时: ${(totalTime/1000).toFixed(2)}秒`);
    console.log(`平均速度: ${(totalTranslated / (totalTime / 1000)).toFixed(2)} 文本/秒`);
    console.log(`最大RSS增长: ${maxRssGrowth.toFixed(2)}MB`);
    
    // 内存增长分析
    if (memoryReadings.length >= 3) {
      const firstReading = memoryReadings[0];
      const lastReading = memoryReadings[memoryReadings.length - 1];
      
      const rssGrowthRate = (lastReading.rss - firstReading.rss) / (lastReading.count - firstReading.count);
      const heapGrowthRate = (lastReading.heapUsed - firstReading.heapUsed) / (lastReading.count - firstReading.count);
      const externalGrowthRate = (lastReading.external - firstReading.external) / (lastReading.count - firstReading.count);
      
      console.log(`\n内存增长率分析:`);
      console.log(`  RSS增长率: ${(rssGrowthRate * 1000).toFixed(4)}MB/千次翻译`);
      console.log(`  堆增长率: ${(heapGrowthRate * 1000).toFixed(4)}MB/千次翻译`);
      console.log(`  外部增长率: ${(externalGrowthRate * 1000).toFixed(4)}MB/千次翻译`);
      
      // 评估修复效果
      if (rssGrowthRate < 0.005) { // 小于5KB/千次翻译
        console.log(`\n✅ WebAssembly内存管理优秀！RSS增长率极低。`);
      } else if (rssGrowthRate < 0.02) { // 小于20KB/千次翻译
        console.log(`\n🟡 WebAssembly内存管理良好，但仍有改进空间。`);
      } else {
        console.log(`\n❌ WebAssembly内存泄漏仍然存在，需要进一步优化。`);
      }
      
      if (memoryLeakDetected) {
        console.log(`🚨 检测到潜在内存泄漏模式！`);
      }
    }
    
    // 等待并观察内存释放
    console.log(`\n等待60秒观察内存自然释放...`);
    for (let i = 1; i <= 6; i++) {
      await new Promise(resolve => setTimeout(resolve, 10000)); // 等待10秒
      const waitMemory = detailedMemoryMonitor(`等待 ${i * 10}秒后`);
      
      if (i === 6) {
        const memoryReleased = finalMemory.rss - waitMemory.rss;
        if (memoryReleased > 0) {
          console.log(`✅ 等待期间释放了 ${memoryReleased.toFixed(2)}MB RSS内存`);
        } else {
          console.log(`⚠️  等待期间内存未释放，可能存在持久泄漏`);
        }
      }
    }
    
  } catch (error) {
    console.error("测试过程中发生错误:", error);
  } finally {
    // 关闭翻译器
    console.log("\n关闭翻译器并强制清理...");
    await Translator.Shutdown();
    LogMem("翻译器关闭后");
    
    // 多次垃圾回收
    if (global.gc) {
      for (let i = 0; i < 3; i++) {
        global.gc();
        await new Promise(resolve => setTimeout(resolve, 100));
      }
      LogMem("多次垃圾回收后");
    }
  }
}

// 运行测试
if (require.main === module) {
  console.log("提示：使用 --expose-gc 参数运行可以观察垃圾回收效果");
  console.log("例如: node --expose-gc example/wasm-memory-management-test.js\n");
  
  testWasmMemoryManagement().catch(console.error);
}

module.exports = {
  testWasmMemoryManagement,
  generateTestTexts,
  detailedMemoryMonitor
}; 