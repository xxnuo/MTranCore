const Translator = require("../js/translator");
const Config = require("../js/config");

/**
 * 内存管理策略演示
 * 展示如何使用简化的内存管理配置
 */

async function demonstrateMemoryStrategies() {
  console.log("=== MTranCore 内存管理策略演示 ===\n");

  // 演示三种内存管理策略
  const strategies = ["conservative", "balanced", "aggressive"];
  
  for (const strategy of strategies) {
    console.log(`--- ${strategy.toUpperCase()} 策略 ---`);
    
    // 设置策略
    process.env.MTRAN_MEMORY_STRATEGY = strategy;
    
    // 获取配置
    const memoryConfig = Config.getMemoryConfig();
    console.log("配置详情：");
    console.log(`  模型释放时间: ${memoryConfig.modelIdleTimeout} 分钟`);
    console.log(`  内存检查间隔: ${memoryConfig.memoryCheckInterval / 1000} 秒`);
    console.log(`  WASM清理间隔: ${memoryConfig.wasmCleanupInterval} 次翻译`);
    console.log(`  WASM清理超时: ${memoryConfig.wasmCleanupTimeout} 分钟`);
    console.log(`  超时重置阈值: ${memoryConfig.timeoutResetThreshold / 1000} 秒`);
    console.log("");
  }

  // 演示自定义配置
  console.log("--- 自定义配置演示 ---");
  process.env.MTRAN_MEMORY_STRATEGY = "balanced";
  process.env.MTRAN_MODEL_IDLE_TIMEOUT = "15"; // 15分钟后释放模型
  
  const customConfig = Config.getMemoryConfig();
  console.log("自定义配置：");
  console.log(`  策略: balanced`);
  console.log(`  自定义模型释放时间: ${customConfig.modelIdleTimeout} 分钟`);
  console.log(`  其他配置继承自策略: balanced`);
  console.log("");

  // 演示禁用自动释放
  console.log("--- 禁用自动释放演示 ---");
  process.env.MTRAN_MODEL_IDLE_TIMEOUT = "0";
  
  const disabledConfig = Config.getMemoryConfig();
  console.log("禁用自动释放：");
  console.log(`  模型释放时间: ${disabledConfig.modelIdleTimeout} (禁用)`);
  console.log("  模型将保持在内存中直到程序结束");
  console.log("");

  // 清理环境变量
  delete process.env.MTRAN_MEMORY_STRATEGY;
  delete process.env.MTRAN_MODEL_IDLE_TIMEOUT;
}

async function demonstrateUsage() {
  console.log("=== 实际使用演示 ===\n");
  
  // 设置为积极策略，适合高频使用
  process.env.MTRAN_MEMORY_STRATEGY = "aggressive";
  process.env.MTRAN_MODEL_IDLE_TIMEOUT = "5"; // 5分钟释放
  
  console.log("当前配置: aggressive 策略，5分钟释放模型");
  
  try {
    // 执行翻译测试
    const result = await Translator.Translate("Hello World", "en", "zh");
    console.log(`翻译结果: ${result}`);
    
    // 显示配置摘要
    const summary = Config.getSummary();
    console.log("\n当前生效的内存配置:");
    console.log(JSON.stringify(summary.memoryConfig, null, 2));
    
  } catch (error) {
    console.error("翻译失败:", error.message);
  }
  
  // 清理
  delete process.env.MTRAN_MEMORY_STRATEGY;
  delete process.env.MTRAN_MODEL_IDLE_TIMEOUT;
}

// 仅演示配置，不执行实际翻译
async function demonstrateConfigOnly() {
  console.log("=== 配置演示（仅配置，不执行翻译）===\n");
  
  // 演示不同策略的配置
  const strategies = ["conservative", "balanced", "aggressive"];
  
  for (const strategy of strategies) {
    console.log(`--- ${strategy.toUpperCase()} 策略配置 ---`);
    process.env.MTRAN_MEMORY_STRATEGY = strategy;
    
    const config = Config.getMemoryConfig();
    const summary = Config.getSummary();
    
    console.log("配置摘要:");
    console.log(`  内存策略: ${summary.memoryStrategy}`);
    console.log(`  模型空闲超时: ${summary.modelIdleTimeout} 分钟`);
    console.log("详细配置:");
    console.log(JSON.stringify(config, null, 2));
    console.log("");
  }
  
  // 清理
  delete process.env.MTRAN_MEMORY_STRATEGY;
}

// 运行演示
async function main() {
  await demonstrateMemoryStrategies();
  await demonstrateConfigOnly();
  
  console.log("=== 推荐使用方式 ===");
  console.log("1. 大多数情况下使用默认的 balanced 策略");
  console.log("2. 内存受限时使用 conservative 策略");
  console.log("3. 高频使用且内存充足时使用 aggressive 策略");
  console.log("4. 需要精确控制时，设置 MTRAN_MODEL_IDLE_TIMEOUT");
  console.log("5. 服务器应用建议禁用自动释放 (MTRAN_MODEL_IDLE_TIMEOUT=0)");
}

if (require.main === module) {
  main().catch(console.error);
}

module.exports = {
  demonstrateMemoryStrategies,
  demonstrateUsage,
  demonstrateConfigOnly
}; 