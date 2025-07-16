"use strict";

const { performance } = require("perf_hooks");
const Translator = require("../js/translator");

/**
 * 测试翻译
 * @param {string} text 要翻译的文本
 * @param {string} fromLang 源语言
 * @param {string} toLang 目标语言
 * @param {boolean} isHTML 是否为HTML内容
 */
async function testTranslation(
  text,
  fromLang = "auto",
  toLang,
  isHTML = false
) {
  console.log(`\n测试翻译文本: "${text}" (${fromLang} -> ${toLang})`);

  try {
    const startTime = performance.now();
    const result = await Translator.Translate(text, fromLang, toLang, isHTML);
    const totalTime = performance.now() - startTime;

    console.log(`翻译结果: "${result}"`);
    console.log(`翻译耗时: ${totalTime.toFixed(2)}ms`);

    return result;
  } catch (error) {
    console.error(`翻译失败: ${error.message}`);
    throw error;
  }
}

/**
 * 测试批量翻译
 * @param {string[]} texts 要翻译的文本数组
 * @param {string} fromLang 源语言
 * @param {string} toLang 目标语言
 * @param {boolean} isHTML 是否为HTML内容
 */
async function testBatchTranslation(
  texts,
  fromLang = "auto",
  toLang,
  isHTML = false
) {
  console.log(
    `\n测试批量翻译 ${texts.length} 个文本 (${fromLang} -> ${toLang}):`
  );
  texts.forEach((text, index) => {
    console.log(`${index + 1}. "${text}"`);
  });

  try {
    const startTime = performance.now();
    const results = await Translator.Translate(texts, fromLang, toLang, isHTML);
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
    // LogMem("测试开始");

    // 测试英文到中文翻译
    console.log("\n测试英文到中文翻译");
    for (let i = 0; i < 3; i++) {
      await testTranslation("Warm up", "en", "zh-Hans");
      await testTranslation("Hello world", "en", "zh-Hans");
      await testTranslation(
        "Open Source Alternative to NotebookLM / Perplexity / Glean, connected to external sources such as search engines (Tavily, Linkup), Slack, Linear, Notion, YouTube, GitHub, Discord and more.",
        "en",
        "zh-Hans"
      );
    }
    // LogMem("英文到中文翻译测试后");

    // 测试英文到中文的批量翻译
    await testBatchTranslation(
      [
        "Hello world",
        "How are you?",
        "Good morning",
        "Open Source Alternative to NotebookLM",
      ],
      "en",
      "zh-Hans"
    );
    // LogMem("英文到中文批量翻译后");

    // 测试中文到英文翻译
    console.log("\n测试中文到英文翻译");
    for (let i = 0; i < 3; i++) {
      await testTranslation("你好，世界", "zh-Hans", "en");
      await testTranslation(
        "WASM VM 的性能无法与编译为 x86 的应用程序相提并论：一个方面是调用WASM模块的开销。传递数据来往。这并不像人们想象的那样多。如果你经常来回调用，这可能是一个需要优化的领域。这也包括有多少数据被传递过去。因为这被复制了。",
        "zh-Hans",
        "en"
      );
    }
    // LogMem("中文到英文翻译测试后");

    // 测试中文到英文的批量翻译
    await testBatchTranslation(
      [
        "你好，世界",
        "今天天气怎么样？",
        "我喜欢编程",
        "WASM VM 的性能无法与编译为 x86 的应用程序相提并论",
      ],
      "zh-Hans",
      "en"
    );
    // LogMem("中文到英文批量翻译后");

    // 测试简繁转换
    console.log("\n测试简体到繁体转换");
    await testTranslation("简体中文转换为繁体中文测试", "zh-Hans", "zh-Hant");
    // LogMem("简体到繁体转换测试后");

    console.log("\n测试繁体到简体转换");
    await testTranslation("繁體中文轉換為簡體中文測試", "zh-Hant", "zh-Hans");
    // LogMem("繁体到简体转换测试后");

    // 测试非英语之间的翻译
    console.log("\n测试中文到日文翻译");
    await testTranslation(
      "这是一个通过英文作为中间语言的翻译测试",
      "zh-Hans",
      "ja"
    );
    // LogMem("中文到日文翻译测试后");

    // 测试自动语言检测
    console.log("\n测试自动语言检测");
    await testTranslation(
      "This is an automatic language detection test",
      "auto",
      "zh-Hans"
    );
    await testTranslation("这是一个自动语言检测测试", "auto", "en");
    // LogMem("自动语言检测测试后");

    console.log("\n测试完成");
    // LogMem("测试结束");

    // 关闭所有翻译引擎
    console.log("\n关闭所有翻译引擎");
    await Translator.Shutdown();
    console.log("所有引擎已关闭");

    // 测试用户输入翻译
    // console.log("\n测试用户输入翻译");
    // const rl = readline.createInterface({
    //   input: process.stdin,
    //   output: process.stdout,
    // });

    // console.log("请输入要翻译的文本（输入'exit'退出）：");

    // const processUserInput = () => {
    //   rl.question("> ", async (input) => {
    //     if (input.toLowerCase() === "exit") {
    //       console.log("翻译会话结束");
    //       rl.close();
    //       return;
    //     }

    //     try {
    //       await testTranslation(translator, input, "auto", "en");
    //     } catch (err) {
    //       console.error("翻译出错:", err.message);
    //     }

    //     processUserInput();
    //   });
    // };

    // processUserInput();

    // // 等待用户输入完成后再结束程序
    // await new Promise((resolve) => rl.once("close", resolve));
  } catch (error) {
    console.error("测试失败:", error);
    // 确保在出错时也关闭所有引擎
    await Translator.Shutdown();
  }
}

// 执行测试
runTest().catch(console.error);
