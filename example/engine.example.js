"use strict";

const fs = require("fs").promises;
const path = require("path");
const { performance } = require("perf_hooks");
const { Engine, CleanText, InitWasm } = require("../js/engine");
const models = require("../js/models");

/**
 * 创建翻译引擎
 * @param {string} fromLang 源语言
 * @param {string} toLang 目标语言
 * @returns {Promise<Engine>} 翻译引擎实例
 */
async function createEngine(fromLang, toLang) {
  console.log(`创建翻译引擎: ${fromLang} -> ${toLang}`);

  // 确保模型已初始化
  await models.init();

  // 构造语言模型文件对象
  const translationModelPayload = await models.getModel(fromLang, toLang);
  console.log("模型文件已加载到内存");

  // 加载Bergamot WASM模块
  const bergamot = await InitWasm();
  console.log("Bergamot WASM模块已初始化");

  // 使用静态工厂方法创建引擎实例
  return new Engine(fromLang, toLang, bergamot, [translationModelPayload]);
}

/**
 * 测试翻译
 * @param {Engine} engine 翻译引擎
 * @param {string} text 要翻译的文本
 * @param {boolean} isHTML 是否为HTML内容
 */
async function testTranslation(engine, text, isHTML = false) {
  console.log(`\n测试翻译文本: "${text}"`);

  try {
    const startTime = performance.now();

    // 清理文本
    const { whitespaceBefore, whitespaceAfter, cleanedSourceText } = CleanText(
      engine.sourceLanguage,
      text
    );

    const result = await engine.translate(cleanedSourceText, isHTML);
    const totalTime = performance.now() - startTime;

    // 恢复原始格式
    const formattedResult = whitespaceBefore + result + whitespaceAfter;

    console.log(`翻译结果: "${formattedResult}"`);
    console.log(`翻译耗时: ${totalTime.toFixed(2)}ms`);

    return formattedResult;
  } catch (error) {
    console.error(`翻译失败: ${error.message}`);
    throw error;
  }
}

/**
 * 测试批量翻译
 * @param {Engine} engine 翻译引擎
 * @param {string[]} texts 要翻译的文本数组
 * @param {boolean} isHTML 是否为HTML内容
 */
async function testBatchTranslation(engine, texts, isHTML = false) {
  console.log(`\n测试批量翻译 ${texts.length} 个文本:`);
  texts.forEach((text, index) => {
    console.log(`${index + 1}. "${text}"`);
  });

  try {
    const startTime = performance.now();

    // 清理所有文本
    const cleanedTexts = texts.map((text) => {
      const { cleanedSourceText } = CleanText(engine.sourceLanguage, text);
      return cleanedSourceText;
    });

    const results = await engine.translate(cleanedTexts, isHTML);
    const totalTime = performance.now() - startTime;

    console.log("\n批量翻译结果:");
    results.forEach((result, index) => {
      console.log(`${index + 1}. "${result}"`);
    });

    console.log(`翻译耗时: ${totalTime.toFixed(2)}ms`);

    return results;
  } catch (error) {
    console.error(`批量翻译失败: ${error.message}`);
    throw error;
  }
}

/**
 * 运行测试
 */
async function runTest() {
  try {
    const engine1 = await createEngine("en", "zh-Hans");
    for (let i = 0; i < 3; i++) {
      await testTranslation(engine1, "Warm up");
      await testTranslation(engine1, "Hello world");
      await testTranslation(
        engine1,
        "Open Source Alternative to NotebookLM / Perplexity / Glean, connected to external sources such as search engines (Tavily, Linkup), Slack, Linear, Notion, YouTube, GitHub, Discord and more."
      );
    }

    // 测试英文到中文的批量翻译
    await testBatchTranslation(engine1, [
      "Hello world",
      "How are you?",
      "Good morning",
      "Open Source Alternative to NotebookLM",
    ]);

    const engine2 = await createEngine("zh-Hans", "en");
    for (let i = 0; i < 3; i++) {
      await testTranslation(engine2, "你好，世界");
      await testTranslation(
        engine2,
        "WASM VM 的性能无法与编译为 x86 的应用程序相提并论：一个方面是调用WASM模块的开销。传递数据来往。这并不像人们想象的那样多。如果你经常来回调用，这可能是一个需要优化的领域。这也包括有多少数据被传递过去。因为这被复制了。"
      );
    }

    // 测试中文到英文的批量翻译
    await testBatchTranslation(engine2, [
      "你好，世界",
      "今天天气怎么样？",
      "我喜欢编程",
      "WASM VM 的性能无法与编译为 x86 的应用程序相提并论",
    ]);

    // 输出模型存储位置
    console.log(`\n所有模型文件存储在: ${models.MODELS_DIR}`);

    // 列出已下载的模型
    const downloadedModels = await models.getDownloadedModels();
    console.log("已下载的模型:", downloadedModels);

    let whitespaceBefore = "";
    let whitespaceAfter = "";
    let cleanedSourceText = "";
    ({ whitespaceBefore, whitespaceAfter, cleanedSourceText } = CleanText(
      "en",
      "  Hello world"
    ));
    console.log(whitespaceBefore, whitespaceAfter, cleanedSourceText);

    ({ whitespaceBefore, whitespaceAfter, cleanedSourceText } = CleanText(
      "zh",
      "“红尘客栈风似刀 骤雨落宿命敲” ··· “任武林谁领风骚我却只为你折腰”“过荒村野桥寻世外古道远离人间尘嚣”"
    ));
    console.log(whitespaceBefore, whitespaceAfter, cleanedSourceText);

    console.log("\n测试完成");
  } catch (error) {
    console.error("测试过程中发生错误:", error);
  }
}

runTest();
