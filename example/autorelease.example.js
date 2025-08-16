"use strict";

// 设置较短的释放间隔，方便测试
process.env.MTRAN_RELEASE_INTERVAL = "0.1"; // 0.1分钟

const Translator = require("../js/translator");
const { LogMem } = require("../js/mem");
const { gc, isGCAvailable } = require("../js/utils");
const { performance } = require("perf_hooks");

/**
 * 等待指定的毫秒数
 * @param {number} ms 等待的毫秒数
 * @returns {Promise<void>}
 */
function sleep(ms) {
  return new Promise(resolve => setTimeout(resolve, ms));
}

/**
 * 测试单个模型的自动释放
 */
async function testSingleModelAutoRelease() {
  console.log("\n===== 测试单个模型的自动释放 =====");
  
  // 记录初始内存
  LogMem("初始状态");
  
  console.log("\n加载英语到中文的翻译模型...");
  const startTime = performance.now();
  
  // 加载模型并翻译一些文本
  const text = "This is a test for automatic model memory release.";
  console.log(`翻译文本: "${text}"`);
  
  const result = await Translator.Translate(text, "en", "zh-Hans");
  console.log(`翻译结果: "${result}"`);
  
  const endTime = performance.now();
  console.log(`翻译耗时: ${(endTime - startTime).toFixed(2)}ms`);
  
  // 记录加载模型后的内存
  LogMem("加载模型后");
  
  // 等待自动释放时间（略大于设置的释放间隔）
  const releaseInterval = parseFloat(process.env.MTRAN_RELEASE_INTERVAL) || 30;
  console.log(`\n等待 ${releaseInterval + 0.2} 分钟让模型自动释放...`);
  await sleep((releaseInterval * 60 + 12) * 1000);
  
  // 记录释放后的内存
  LogMem("模型应该已被自动释放");
  
  // 强制垃圾回收（如果可用）
  try {
    if (isGCAvailable) {
      console.log("\n强制垃圾回收...");
      gc();
      LogMem("垃圾回收后");
    }
  } catch (e) {
    console.log("\n无法强制垃圾回收。如需测试垃圾回收，请使用 --expose-gc 参数运行 Node.js");
  }
}

/**
 * 测试多个模型的自动释放
 */
async function testMultipleModelsAutoRelease() {
  console.log("\n===== 测试多个模型的自动释放 =====");
  
  // 记录初始内存
  LogMem("初始状态");
  
  // 语言对列表
  const languagePairs = [
    { from: "en", to: "zh-Hans", text: "This is a test for English to Chinese translation." },
    { from: "en", to: "ja", text: "This is a test for English to Japanese translation." },
    { from: "en", to: "fr", text: "This is a test for English to French translation." }
  ];
  
  // 加载多个模型并翻译
  console.log("\n依次加载多个翻译模型...");
  
  for (const pair of languagePairs) {
    const startTime = performance.now();
    console.log(`\n翻译 ${pair.from} 到 ${pair.to}:`);
    console.log(`原文: "${pair.text}"`);
    
    const result = await Translator.Translate(pair.text, pair.from, pair.to);
    console.log(`译文: "${result}"`);
    
    const endTime = performance.now();
    console.log(`翻译耗时: ${(endTime - startTime).toFixed(2)}ms`);
    
    // 短暂等待
    await sleep(1000);
  }
  
  // 记录所有模型加载后的内存
  LogMem("所有模型加载后");
  
  // 等待自动释放时间（略大于设置的释放间隔）
  const releaseInterval = parseFloat(process.env.MTRAN_RELEASE_INTERVAL) || 30;
  console.log(`\n等待 ${releaseInterval + 0.2} 分钟让模型自动释放...`);
  await sleep((releaseInterval * 60 + 12) * 1000);
  
  // 记录释放后的内存
  LogMem("模型应该已被自动释放");
  
  // 强制垃圾回收（如果可用）
  try {
    if (isGCAvailable) {
      console.log("\n强制垃圾回收...");
      gc();
      LogMem("垃圾回收后");
    }
  } catch (e) {
    console.log("\n无法强制垃圾回收。如需测试垃圾回收，请使用 --expose-gc 参数运行 Node.js");
  }
}

/**
 * 测试模型使用中不被释放
 */
async function testModelKeepAlive() {
  console.log("\n===== 测试模型使用中不被释放 =====");
  
  // 记录初始内存
  LogMem("初始状态");
  
  console.log("\n加载英语到中文的翻译模型...");
  
  // 加载模型并翻译一些文本
  const text = "This is a test for model keep-alive functionality.";
  console.log(`翻译文本: "${text}"`);
  
  const result = await Translator.Translate(text, "en", "zh-Hans");
  console.log(`翻译结果: "${result}"`);
  
  // 记录加载模型后的内存
  LogMem("加载模型后");
  
  // 定期翻译，保持模型活跃
  const releaseInterval = parseFloat(process.env.MTRAN_RELEASE_INTERVAL) || 30;
  const keepAliveInterval = Math.floor(releaseInterval * 60 * 1000 / 2); // 释放间隔的一半
  
  console.log(`\n每 ${keepAliveInterval / 1000} 秒翻译一次，保持模型活跃，持续 ${releaseInterval + 0.2} 分钟...`);
  
  const startTime = Date.now();
  const endTime = startTime + (releaseInterval * 60 + 12) * 1000;
  
  let counter = 1;
  while (Date.now() < endTime) {
    await sleep(keepAliveInterval);
    const keepAliveText = `Keep-alive translation ${counter++}.`;
    const keepAliveResult = await Translator.Translate(keepAliveText, "en", "zh-Hans");
    console.log(`${new Date().toISOString()} - 保活翻译: "${keepAliveText}" => "${keepAliveResult}"`);
  }
  
  // 记录模型应该仍然活跃的内存
  LogMem("模型应该仍然活跃");
  
  // 现在停止使用模型，等待自动释放
  console.log(`\n停止使用模型，等待 ${releaseInterval + 0.2} 分钟让模型自动释放...`);
  await sleep((releaseInterval * 60 + 12) * 1000);
  
  // 记录释放后的内存
  LogMem("模型应该已被自动释放");
}

/**
 * 主测试函数
 */
async function runTests() {
  console.log("开始测试模型自动释放功能");
  console.log(`当前设置的自动释放间隔: ${process.env.MTRAN_RELEASE_INTERVAL || 30} 分钟`);
  
  // 测试单个模型的自动释放
  await testSingleModelAutoRelease();
  
  // 测试多个模型的自动释放
  await testMultipleModelsAutoRelease();
  
  // 测试模型使用中不被释放
  await testModelKeepAlive();
  
  console.log("\n所有测试完成");
  
  // 关闭翻译器
  await Translator.Shutdown();
}

// 运行测试
runTests().catch(err => {
  console.error("测试过程中发生错误:", err);
  process.exit(1);
});
