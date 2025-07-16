const axios = require("axios");
const fs = require("fs");
const path = require("path");
const crypto = require("crypto");

const Config = require("./config");

/**
 * 高级下载器 - 支持错误重试、超时和校验
 */
class Downloader {
  /**
   * 创建下载器实例
   * @param {Object} options 下载器配置选项
   * @param {number} options.maxRetries 最大重试次数
   * @param {number} options.timeout 超时时间(毫秒)
   * @param {number} options.retryDelay 重试延迟(毫秒)
   * @param {boolean} options.validateChecksum 是否校验文件完整性
   * @param {Object} options.headers 默认请求头
   */
  constructor(options = {}) {
    this.maxRetries = options.maxRetries || 3;
    this.timeout = options.timeout || 30000;
    this.retryDelay = options.retryDelay || 1000;
    this.validateChecksum = options.validateChecksum || true;

    // 默认使用 Firefox Nightly 浏览器的用户代理
    this.defaultHeaders = {
      "User-Agent": Config.DOWNLOADER_UA,
      "Accept-Encoding": "gzip, deflate, br",
      Connection: "keep-alive",
      ...options.headers,
    };
  }

  /**
   * 下载文件
   * @param {string} url 下载链接
   * @param {string} destination 保存路径
   * @param {Object} options 下载选项
   * @param {string} options.checksum 预期的校验和
   * @param {string} options.algorithm 校验算法 (md5, sha1, sha256等)
   * @param {Object} options.headers 请求头
   * @returns {Promise<Object>} 下载结果
   */
  async download(url, destination, options = {}) {
    if (Config.OFFLINE) {
      throw new Error("Offline mode is not supported for downloading files.");
    }
    const { checksum, algorithm = "sha256", headers = {} } = options;
    let retries = 0;
    let error = null;

    // 确保目标目录存在
    const dir = path.dirname(destination);
    if (!fs.existsSync(dir)) {
      fs.mkdirSync(dir, { recursive: true });
    }

    while (retries <= this.maxRetries) {
      try {
        if (retries > 0) {
          console.log(
            `Retrying download (${retries}/${this.maxRetries}): ${url}`
          );
          // 重试前等待
          await new Promise((resolve) => setTimeout(resolve, this.retryDelay));
        }

        const response = await axios({
          method: "GET",
          url: url,
          responseType: "stream",
          timeout: this.timeout,
          // 合并默认头部和自定义头部
          headers: { ...this.defaultHeaders, ...headers },
        });

        const writer = fs.createWriteStream(destination);

        await new Promise((resolve, reject) => {
          response.data.pipe(writer);

          let error = null;
          writer.on("error", (err) => {
            error = err;
            writer.close();
            reject(err);
          });

          writer.on("close", () => {
            if (!error) {
              resolve(true);
            }
          });
        });

        // 如果需要校验文件完整性
        if (this.validateChecksum && checksum) {
          const isValid = await this.verifyChecksum(
            destination,
            checksum,
            algorithm
          );
          if (!isValid) {
            throw new Error(
              `Checksum mismatch, file may be corrupted. ${destination}`
            );
          }
        }

        return {
          success: true,
          url,
          destination,
          retries,
          fileSize: fs.statSync(destination).size,
        };
      } catch (err) {
        error = err;
        retries++;

        // 如果下载失败且文件存在，则删除不完整的文件
        if (fs.existsSync(destination)) {
          fs.unlinkSync(destination);
        }
      }
    }

    // 所有重试都失败
    throw new Error(`Download failed (${url}): ${error.message}`);
  }

  /**
   * 校验文件完整性
   * @param {string} filePath 文件路径
   * @param {string} expectedChecksum 预期的校验和
   * @param {string} algorithm 校验算法
   * @returns {Promise<boolean>} 校验结果
   */
  async verifyChecksum(filePath, expectedChecksum, algorithm) {
    return new Promise((resolve, reject) => {
      const hash = crypto.createHash(algorithm);
      const stream = fs.createReadStream(filePath);

      stream.on("error", (err) => reject(err));
      stream.on("data", (chunk) => hash.update(chunk));
      stream.on("end", () => {
        const fileChecksum = hash.digest("hex");
        resolve(fileChecksum.toLowerCase() === expectedChecksum.toLowerCase());
      });
    });
  }

  /**
   * 批量下载多个文件
   * @param {Array<Object>} files 文件列表
   * @param {string} files[].url 下载链接
   * @param {string} files[].destination 保存路径
   * @param {Object} files[].options 下载选项
   * @returns {Promise<Array>} 下载结果列表
   */
  async batchDownload(files) {
    const results = [];

    for (const file of files) {
      try {
        const result = await this.download(
          file.url,
          file.destination,
          file.options || {}
        );
        results.push({
          ...result,
          status: "success",
        });
      } catch (error) {
        results.push({
          url: file.url,
          destination: file.destination,
          status: "failed",
          error: error.message,
        });
      }
    }

    return results;
  }
}

module.exports = Downloader;
