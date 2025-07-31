const Translator = require("../js/translator");

/**
 * 简化内存管理配置演示
 * 展示如何使用新的简化内存管理配置
 */

// 重新加载Config模块的辅助函数
function reloadConfig() {
  delete require.cache[require.resolve("../js/config")];
  return require("../js/config");
}

async function demonstrateSimplifiedMemoryConfig() {
  console.log("=== MTranCore 简化内存管理配置演示 ===\n");

  // 演示默认配置
  console.log("--- 默认配置 ---");
  const Config = reloadConfig();
  const defaultSummary = Config.getSummary();
  console.log("默认配置：");
  console.log(`  模型空闲超时: ${defaultSummary.modelIdleTimeout} 分钟`);
  console.log(`  Worker重启任务数: ${defaultSummary.workerRestartTaskCount} 次`);
  console.log(`  内存检查间隔: ${defaultSummary.memoryCheckInterval / 1000} 秒`);
  console.log("");

  // 演示自定义配置
  console.log("--- 自定义配置演示 ---");
  process.env.MTRAN_MODEL_IDLE_TIMEOUT = "15"; // 15分钟后释放模型
  process.env.MTRAN_WORKER_RESTART_TASK_COUNT = "500"; // 500次任务后重启Worker
  process.env.MTRAN_MEMORY_CHECK_INTERVAL = "30000"; // 30秒检查一次
  
  const CustomConfig = reloadConfig();
  const customSummary = CustomConfig.getSummary();
  console.log("自定义配置：");
  console.log(`  模型空闲超时: ${customSummary.modelIdleTimeout} 分钟`);
  console.log(`  Worker重启任务数: ${customSummary.workerRestartTaskCount} 次`);
  console.log(`  内存检查间隔: ${customSummary.memoryCheckInterval / 1000} 秒`);
  console.log("");

  // 演示禁用自动释放
  console.log("--- 禁用自动释放演示 ---");
  process.env.MTRAN_MODEL_IDLE_TIMEOUT = "0";
  
  const DisabledConfig = reloadConfig();
  const disabledSummary = DisabledConfig.getSummary();
  console.log("禁用自动释放：");
  console.log(`  模型空闲超时: ${disabledSummary.modelIdleTimeout} 分钟 (禁用)`);
  console.log("  模型将保持在内存中直到程序结束");
  console.log("");

  // 清理环境变量
  delete process.env.MTRAN_MODEL_IDLE_TIMEOUT;
  delete process.env.MTRAN_WORKER_RESTART_TASK_COUNT;
  delete process.env.MTRAN_MEMORY_CHECK_INTERVAL;
}

async function demonstrateUsage() {
  console.log("=== 实际使用演示 ===\n");
  
  // 设置适合高频使用的配置
  process.env.MTRAN_MODEL_IDLE_TIMEOUT = "10"; // 10分钟释放
  process.env.MTRAN_WORKER_RESTART_TASK_COUNT = "2000"; // 2000次任务后重启
  
  console.log("当前配置: 10分钟释放模型，2000次任务后重启Worker");
  
  try {
    // 执行翻译测试
    const result = await Translator.Translate("Hello World", "en", "zh");
    console.log(`翻译结果: ${result}`);
    
    // 显示配置摘要
    const Config = reloadConfig();
    const summary = Config.getSummary();
    console.log("\n当前生效的内存配置:");
    console.log(`  模型空闲超时: ${summary.modelIdleTimeout} 分钟`);
    console.log(`  Worker重启任务数: ${summary.workerRestartTaskCount} 次`);
    console.log(`  内存检查间隔: ${summary.memoryCheckInterval / 1000} 秒`);
    
  } catch (error) {
    console.error("翻译失败:", error.message);
  }
  
  // 清理
  delete process.env.MTRAN_MODEL_IDLE_TIMEOUT;
  delete process.env.MTRAN_WORKER_RESTART_TASK_COUNT;
}

// 仅演示配置，不执行实际翻译
async function demonstrateConfigOnly() {
  console.log("=== 不同场景配置演示 ===\n");
  
  const scenarios = [
    {
      name: "内存受限环境",
      config: {
        MTRAN_MODEL_IDLE_TIMEOUT: "5",
        MTRAN_WORKER_RESTART_TASK_COUNT: "500",
        MTRAN_MEMORY_CHECK_INTERVAL: "30000"
      }
    },
    {
      name: "高频使用环境",
      config: {
        MTRAN_MODEL_IDLE_TIMEOUT: "60",
        MTRAN_WORKER_RESTART_TASK_COUNT: "5000",
        MTRAN_MEMORY_CHECK_INTERVAL: "120000"
      }
    },
    {
      name: "服务器环境",
      config: {
        MTRAN_MODEL_IDLE_TIMEOUT: "0",
        MTRAN_WORKER_RESTART_TASK_COUNT: "10000",
        MTRAN_MEMORY_CHECK_INTERVAL: "300000"
      }
    }
  ];
  
  for (const scenario of scenarios) {
    console.log(`--- ${scenario.name} ---`);
    
    // 设置配置
    for (const [key, value] of Object.entries(scenario.config)) {
      process.env[key] = value;
    }
    
    const Config = reloadConfig();
    const summary = Config.getSummary();
    console.log("配置:");
    console.log(`  模型空闲超时: ${summary.modelIdleTimeout} 分钟`);
    console.log(`  Worker重启任务数: ${summary.workerRestartTaskCount} 次`);
    console.log(`  内存检查间隔: ${summary.memoryCheckInterval / 1000} 秒`);
    console.log("");
    
    // 清理配置
    for (const key of Object.keys(scenario.config)) {
      delete process.env[key];
    }
  }
}

// 运行演示
async function main() {
  await demonstrateSimplifiedMemoryConfig();
  await demonstrateConfigOnly();
  
  console.log("=== 推荐使用方式 ===");
  console.log("1. 大多数情况下使用默认配置即可");
  console.log("2. 内存受限时减少模型空闲超时时间和Worker重启任务数");
  console.log("3. 高频使用时增加模型空闲超时时间和Worker重启任务数");
  console.log("4. 服务器应用建议禁用自动释放 (MTRAN_MODEL_IDLE_TIMEOUT=0)");
  console.log("5. 根据需要调整内存检查间隔");
}

if (require.main === module) {
  main().catch(console.error);
}

module.exports = {
  demonstrateSimplifiedMemoryConfig,
  demonstrateUsage,
  demonstrateConfigOnly
}; 