const Downloader = require("../js/downloader");

/**
 * 实际使用场景示例
 */
async function realWorldExample() {
  const downloader = new Downloader();

  // 创建下载目录
  const fs = require("fs");
  if (!fs.existsSync("./downloads")) {
    fs.mkdirSync("./downloads");
  }

  try {
    // 下载Node.js文档
    console.log("下载Node.js文档...");
    await downloader.download(
      "https://nodejs.org/dist/latest/docs/api/documentation.html",
      "./packages/downloads/nodejs-docs.html"
    );

    // 下载一个开源项目的ZIP包
    console.log("下载开源项目...");
    await downloader.download(
      "https://github.com/axios/axios/archive/refs/heads/master.zip",
      "./packages/downloads/axios-master.zip"
    );

    console.log("所有文件下载完成!");
  } catch (error) {
    console.error("实际示例下载失败:", error);
  }
}

// 取消注释下面的行来运行实际示例
realWorldExample().catch(console.error);
