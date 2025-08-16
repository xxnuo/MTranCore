"use strict";

const { performance } = require("perf_hooks");
const Translator = require("../js/translator");
const Logger = require("../js/logger");

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
 * 重复执行单项翻译测试
 * @param {string} text 要翻译的文本
 * @param {string} fromLang 源语言
 * @param {string} toLang 目标语言
 * @param {number} repeatCount 重复次数
 * @param {boolean} isHTML 是否为HTML内容
 */
async function repeatTranslationTest(
  text,
  fromLang,
  toLang,
  repeatCount = 3,
  isHTML = false
) {
  Logger.debug(
    "Test",
    `开始重复测试 ${repeatCount} 次 (${fromLang} -> ${toLang})`
  );
  for (let i = 0; i < repeatCount; i++) {
    await testTranslation(text, fromLang, toLang, isHTML);
  }
}

/**
 * 运行测试
 */
async function runTest() {
  const REPEAT_COUNT = 3; // 统一的重复测试次数

  try {
    // LogMem("测试开始");

    // 测试英文到中文翻译
    console.log("\n测试英文到中文翻译");
    await testTranslation("Warm up", "en", "zh-Hans"); // 预热
    await repeatTranslationTest("Hello world", "en", "zh-Hans", REPEAT_COUNT);
    await repeatTranslationTest(
      "Open Source Alternative to NotebookLM / Perplexity / Glean, connected to external sources such as search engines (Tavily, Linkup), Slack, Linear, Notion, YouTube, GitHub, Discord and more.",
      "en",
      "zh-Hans",
      REPEAT_COUNT
    );
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
    await testTranslation("预热", "zh-Hans", "en"); // 预热
    await repeatTranslationTest("你好，世界", "zh-Hans", "en", REPEAT_COUNT);
    await repeatTranslationTest(
      "WASM VM 的性能无法与编译为 x86 的应用程序相提并论：一个方面是调用WASM模块的开销。传递数据来往。这并不像人们想象的那样多。如果你经常来回调用，这可能是一个需要优化的领域。这也包括有多少数据被传递过去。因为这被复制了。",
      "zh-Hans",
      "en",
      REPEAT_COUNT
    );
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
    await repeatTranslationTest(
      "简体中文转换为繁体中文测试",
      "zh-Hans",
      "zh-Hant",
      REPEAT_COUNT
    );
    // LogMem("简体到繁体转换测试后");

    console.log("\n测试繁体到简体转换");
    await repeatTranslationTest(
      "繁體中文轉換為簡體中文測試",
      "zh-Hant",
      "zh-Hans",
      REPEAT_COUNT
    );
    // LogMem("繁体到简体转换测试后");

    // 测试非英语之间的翻译
    console.log("\n测试中文到日文翻译");
    await repeatTranslationTest(
      "这是一个通过英文作为中间语言的翻译测试",
      "zh-Hans",
      "ja",
      REPEAT_COUNT
    );
    // LogMem("中文到日文翻译测试后");

    // 测试其他跨英文的翻译
    console.log("\n测试德语到法语翻译");
    await repeatTranslationTest(
      "Hallo, wie geht es Ihnen heute?",
      "de",
      "fr",
      REPEAT_COUNT
    );
    // LogMem("德语到法语翻译测试后");

    console.log("\n测试西班牙语到意大利语翻译");
    await repeatTranslationTest(
      "Buenos días, ¿cómo estás?",
      "es",
      "it",
      REPEAT_COUNT
    );
    // LogMem("西班牙语到意大利语翻译测试后");

    console.log("\n测试俄语到葡萄牙语翻译");
    await repeatTranslationTest(
      "Здравствуйте, как вы сегодня?",
      "ru",
      "pt",
      REPEAT_COUNT
    );
    // LogMem("俄语到葡萄牙语翻译测试后");

    // 测试跨语言批量翻译

    // 测试中文到日文批量翻译
    console.log("\n测试中文到日文批量翻译");
    const zhToJaTexts = [
      "这是一个通过英文作为中间语言的翻译测试",
      "批量翻译可以提高效率",
      "跨语言翻译需要中间语言",
    ];
    await testBatchTranslation(zhToJaTexts, "zh-Hans", "ja", false);

    // 测试德语到法语批量翻译
    console.log("\n测试德语到法语批量翻译");
    const deToFrTexts = [
      "Hallo, wie geht es Ihnen heute?",
      "Dies ist ein Batch-Übersetzungstest",
      "Mehrsprachige Übersetzung ist wichtig",
    ];
    await testBatchTranslation(deToFrTexts, "de", "fr", false);

    // 测试西班牙语到意大利语批量翻译
    console.log("\n测试西班牙语到意大利语批量翻译");
    const esToItTexts = [
      "Buenos días, ¿cómo estás?",
      "Esta es una prueba de traducción por lotes",
      "La traducción entre idiomas es útil",
    ];
    await testBatchTranslation(esToItTexts, "es", "it", false);

    // 测试俄语到葡萄牙语批量翻译
    console.log("\n测试俄语到葡萄牙语批量翻译");
    const ruToPtTexts = [
      "Здравствуйте, как вы сегодня?",
      "Это тест пакетного перевода",
      "Перевод между языками важен",
    ];
    await testBatchTranslation(ruToPtTexts, "ru", "pt", false);
    // LogMem("批量翻译测试后");

    // 测试自动语言检测
    console.log("\n测试自动语言检测");
    await repeatTranslationTest(
      "This is an automatic language detection test",
      "auto",
      "zh-Hans",
      REPEAT_COUNT
    );
    await repeatTranslationTest(
      "这是一个自动语言检测测试",
      "auto",
      "en",
      REPEAT_COUNT
    );
    // LogMem("自动语言检测测试后");

    // 测试HTML内容翻译
    console.log("\n测试HTML内容翻译");
    await repeatTranslationTest(
      "<p>This is a <strong>HTML</strong> translation test.</p>",
      "en",
      "zh-Hans",
      REPEAT_COUNT,
      true
    );
    // LogMem("HTML内容翻译测试后");

    // 测试长文本翻译
    console.log("\n测试长文本翻译");
    await repeatTranslationTest(
      "The WebAssembly System Interface (WASI) is a modular system interface for WebAssembly. " +
        "As WebAssembly continues to grow in popularity and expand into new domains, " +
        "there's an increasing need for a standard way for WebAssembly modules to interact with their environment. " +
        "WASI aims to provide a standardized set of APIs that allow WebAssembly modules to access system resources " +
        "in a secure, portable, and efficient manner.",
      "en",
      "zh-Hans",
      2 // 长文本减少重复次数
    );
    // LogMem("长文本翻译测试后");

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
