"use strict";

const { LogMem } = require("../js/mem");
const { performance } = require("perf_hooks");

/**
 * 创建大数组来测试内存使用
 * @param {number} size 数组大小
 * @returns {Array} 创建的数组
 */
function createLargeArray(size) {
  console.log(`创建大小为 ${size} 的数组...`);
  return new Array(size).fill(0).map((_, i) => i);
}

/**
 * 运行内存测试
 */
async function runMemoryTest() {
  console.log("开始内存使用测试");

  // 初始内存使用
  console.log("\n");
  LogMem("初始状态");

  // 创建小数组
  let smallArray = createLargeArray(1000);
  console.log("\n");
  LogMem("创建小数组后");

  // 创建中等数组
  let mediumArray = createLargeArray(100000);
  console.log("\n");
  LogMem("创建中等数组后");

  // 创建大数组
  let largeArray = createLargeArray(1000000);
  console.log("\n");
  LogMem("创建大数组后");

  // 释放小数组和中等数组
  smallArray = null;
  mediumArray = null;
  console.log("\n");
  LogMem("释放小数组和中等数组后");

  // 强制垃圾回收（注意：这在生产环境中通常不可靠）
  try {
    if (global.gc) {
      console.log("\n");
      LogMem("尝试强制垃圾回收");
      global.gc();
      console.log("\n");
      LogMem("垃圾回收后");
    }
  } catch (e) {
    console.log(
      "\n无法强制垃圾回收。如需测试垃圾回收，请使用 --expose-gc 参数运行 Node.js"
    );
    console.log("例如: node --expose-gc js/mem.example.js");
  }

  // 释放大数组
  largeArray = null;
  console.log("\n");
  LogMem("释放大数组后");

  // 测试内存泄漏场景
  console.log("\n测试潜在的内存泄漏场景（创建并释放多个大数组）...");
  const startTime = performance.now();

  for (let i = 0; i < 5; i++) {
    console.log(`\n`);
    LogMem(`迭代 ${i + 1}`);
    const tempArray = createLargeArray(500000);
    LogMem(`迭代 ${i + 1} 后`);
    // 模拟一些处理
    await new Promise((resolve) => setTimeout(resolve, 1000));
    // 释放数组
    tempArray.length = 0;
  }

  const endTime = performance.now();
  console.log(
    `\n内存泄漏测试完成，耗时: ${(endTime - startTime).toFixed(2)}ms`
  );
  LogMem("内存泄漏测试完成");

  console.log("\n内存使用测试完成");
}

// 运行测试
runMemoryTest().catch((err) => {
  console.error("测试过程中发生错误:", err);
});
