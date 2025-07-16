const path = require("path");
const fs = require("fs");
const { performance } = require("perf_hooks");

const models = require("../js/models");

async function testModels() {
  try {
    await models.init();
    // 1. 列出所有可用的语言对
    console.log("Listing all available language pairs...");
    const availablePairs = await models.getAllModels();
    console.log("Available language pairs:", availablePairs);

    // 3. 列出已下载的语言对
    console.log("Listing downloaded language pairs...");
    const downloadedModels = await models.getDownloadedModels();
    console.log(
      "Downloaded language pairs:",
      downloadedModels,
      `... (total: ${downloadedModels.length})`
    );

    // 4. 为英语到中文下载模型
    console.log("\nDownloading models for en_zh-Hans...");
    performance.mark("start");
    const enzhModels = await models.getModel("en", "zh-Hans");
    performance.mark("end");
    performance.measure("getModel", "start", "end");
    console.log(
      `getModel time: ${performance.getEntriesByName("getModel")[0].duration}ms`
    );
    
    // 5. 验证下载的模型文件完整性
    console.log("\nValidating downloaded model files...");
    for (const [fileType, filePath] of Object.entries(enzhModels)) {
      if (filePath && fs.existsSync(filePath)) {
        // 检查文件是否存在且大小不为0
        const stats = fs.statSync(filePath);
        const isFileValid = stats.size > 0;
        console.log(
          `${fileType} file (${path.basename(filePath)}) is ${
            isFileValid ? "valid" : "invalid"
          }`
        );
      }
    }

    // 6. 测试缓存机制 - 再次获取相同模型，应该使用缓存
    console.log("\nTesting cache mechanism for en_zh-Hans...");
    console.time("Cache retrieval");
    performance.mark("start");
    const cachedModels = await models.getModel("en", "zh-Hans");
    performance.mark("end");
    performance.measure("getModel", "start", "end");
    console.log(
      `getModel time: ${performance.getEntriesByName("getModel")[0].duration}ms`
    );
    console.timeEnd("Cache retrieval");
    console.log("Cached models:", Object.keys(cachedModels));

    // 7. 为中文到英语下载模型
    console.log("\nDownloading models for zh-Hans_en...");
    const zhenModels = await models.getModel("zh-Hans", "en");
    
    console.log(`\nAll model files are stored in: ${models.MODELS_DIR}`);
  } catch (error) {
    console.error("Error in test:", error);
  } finally {
    // 确保脚本执行完成后退出
    process.exit(0);
  }
}

testModels();
