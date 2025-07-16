"use strict";

const readline = require("readline");
const { performance } = require("perf_hooks");
const Translator = require("../js/translator");
const { LogMem } = require("./mem");

/**
 * 命令行翻译工具
 * 支持从命令行读取任意语言并翻译为中文
 */
async function cliTranslator() {
  console.log("=== 命令行翻译工具 ===");
  console.log("支持从命令行读取任意语言并翻译为中文");
  console.log("输入 'exit' 或 'quit' 退出程序");
  console.log("输入 'lang:xx' 可以指定源语言 (例如: lang:en)，默认为自动检测");
  console.log("==============================");

  // 创建命令行交互界面
  const rl = readline.createInterface({
    input: process.stdin,
    output: process.stdout,
    prompt: "翻译> ",
  });

  let sourceLang = "auto"; // 默认自动检测语言
  const targetLang = "zh-Hans"; // 目标语言固定为简体中文

  // 提示用户输入
  rl.prompt();

  // 处理用户输入
  rl.on("line", async (input) => {
    // 检查是否退出
    if (input.toLowerCase() === "exit" || input.toLowerCase() === "quit") {
      console.log("正在关闭翻译引擎...");
      await Translator.Shutdown();
      console.log("翻译工具已退出");
      process.exit(0);
    }

    // 检查是否更改源语言
    if (input.startsWith("lang:")) {
      const newLang = input.substring(5).trim();
      if (newLang) {
        sourceLang = newLang;
        console.log(`源语言已设置为: ${sourceLang}`);
      } else {
        console.log(`当前源语言: ${sourceLang}`);
      }
      rl.prompt();
      return;
    }

    // 如果输入为空，直接返回提示
    if (!input.trim()) {
      rl.prompt();
      return;
    }

    try {
      // 显示正在翻译的提示
      console.log(`正在翻译 (${sourceLang} -> ${targetLang})...`);

      // 记录开始时间
      const startTime = performance.now();

      // 执行翻译
      LogMem("翻译前");
      const result = await Translator.Translate(input, sourceLang, targetLang);
      LogMem("翻译后");

      // 计算翻译耗时
      const totalTime = performance.now() - startTime;

      // 显示翻译结果
      console.log("\n翻译结果:");
      console.log(result);
      console.log(`\n[耗时: ${totalTime.toFixed(2)}ms]`);
    } catch (error) {
      console.error(`翻译失败: ${error.message}`);
    }

    // 继续提示用户输入
    rl.prompt();
  });

  // 处理关闭事件
  rl.on("close", async () => {
    console.log("\n正在关闭翻译引擎...");
    await Translator.Shutdown();
    console.log("翻译工具已退出");
    process.exit(0);
  });

  // 处理程序退出
  process.on("SIGINT", async () => {
    console.log("\n正在关闭翻译引擎...");
    await Translator.Shutdown();
    console.log("翻译工具已退出");
    process.exit(0);
  });
}

// 执行命令行翻译工具
cliTranslator().catch((error) => {
  console.error("程序运行出错:", error);
  process.exit(1);
});
