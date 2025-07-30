const Logger = require("./logger");

/**
 * 内存监控和分析工具
 */
class MemoryMonitor {
  // 内存使用历史记录
  static _memoryHistory = [];
  
  // 最大历史记录数量
  static MAX_HISTORY_SIZE = 100;

  /**
   * 格式化字节数为可读的字符串
   * @param {number} bytes - 字节数
   * @returns {string} 格式化后的字符串
   * @private
   */
  static _formatBytes(bytes) {
    const sizes = ['B', 'KB', 'MB', 'GB'];
    if (bytes === 0) return '0 B';
    
    const i = Math.floor(Math.log(bytes) / Math.log(1024));
    const size = (bytes / Math.pow(1024, i)).toFixed(2);
    return `${size} ${sizes[i]}`;
  }

  /**
   * 获取当前内存使用情况
   * @returns {Object} 内存使用数据
   */
  static getCurrentMemoryUsage() {
    const memUsage = process.memoryUsage();
    const timestamp = Date.now();
    
    return {
      timestamp,
      rss: memUsage.rss,
      heapTotal: memUsage.heapTotal,
      heapUsed: memUsage.heapUsed,
      external: memUsage.external,
      arrayBuffers: memUsage.arrayBuffers || 0,
    };
  }

  /**
   * 记录并输出内存使用情况
   * @param {string} label - 标签，用于标识当前操作
   * @param {boolean} [saveToHistory=true] - 是否保存到历史记录
   */
  static logMemory(label, saveToHistory = true) {
    const memData = MemoryMonitor.getCurrentMemoryUsage();
    
    // 保存到历史记录
    if (saveToHistory) {
      MemoryMonitor._addToHistory(label, memData);
    }
    
    // 输出到日志
    Logger.debug('MemoryMonitor', `内存使用情况 - ${label}`);
    Logger.debug('MemoryMonitor', `  RSS: ${MemoryMonitor._formatBytes(memData.rss)}`);
    Logger.debug('MemoryMonitor', `  堆总大小: ${MemoryMonitor._formatBytes(memData.heapTotal)}`);
    Logger.debug('MemoryMonitor', `  堆已用: ${MemoryMonitor._formatBytes(memData.heapUsed)}`);
    Logger.debug('MemoryMonitor', `  外部: ${MemoryMonitor._formatBytes(memData.external)}`);
    
    if (memData.arrayBuffers > 0) {
      Logger.debug('MemoryMonitor', `  ArrayBuffers: ${MemoryMonitor._formatBytes(memData.arrayBuffers)}`);
    }
  }

  /**
   * 添加内存数据到历史记录
   * @param {string} label - 标签
   * @param {Object} memData - 内存数据
   * @private
   */
  static _addToHistory(label, memData) {
    MemoryMonitor._memoryHistory.push({
      label,
      ...memData,
    });
    
    // 限制历史记录大小
    if (MemoryMonitor._memoryHistory.length > MemoryMonitor.MAX_HISTORY_SIZE) {
      MemoryMonitor._memoryHistory.shift();
    }
  }

  /**
   * 获取内存使用历史
   * @returns {Array} 内存使用历史数组
   */
  static getMemoryHistory() {
    return [...MemoryMonitor._memoryHistory];
  }

  /**
   * 清理内存历史记录
   */
  static clearHistory() {
    MemoryMonitor._memoryHistory.length = 0;
    Logger.debug('MemoryMonitor', '已清理内存使用历史记录');
  }

  /**
   * 计算内存增长情况
   * @param {string} startLabel - 开始标签
   * @param {string} endLabel - 结束标签
   * @returns {Object|null} 内存增长数据，如果找不到对应记录则返回null
   */
  static calculateMemoryGrowth(startLabel, endLabel) {
    const startRecord = MemoryMonitor._memoryHistory.find(record => record.label === startLabel);
    const endRecord = MemoryMonitor._memoryHistory.find(record => record.label === endLabel);
    
    if (!startRecord || !endRecord) {
      Logger.warn('MemoryMonitor', `无法找到标签 "${startLabel}" 或 "${endLabel}" 的内存记录`);
      return null;
    }
    
    const growth = {
      timeElapsed: endRecord.timestamp - startRecord.timestamp,
      rssGrowth: endRecord.rss - startRecord.rss,
      heapTotalGrowth: endRecord.heapTotal - startRecord.heapTotal,
      heapUsedGrowth: endRecord.heapUsed - startRecord.heapUsed,
      externalGrowth: endRecord.external - startRecord.external,
    };
    
    return growth;
  }

  /**
   * 输出内存增长报告
   * @param {string} startLabel - 开始标签
   * @param {string} endLabel - 结束标签
   */
  static logMemoryGrowth(startLabel, endLabel) {
    const growth = MemoryMonitor.calculateMemoryGrowth(startLabel, endLabel);
    
    if (!growth) {
      return;
    }
    
    Logger.info('MemoryMonitor', `内存增长报告: ${startLabel} -> ${endLabel}`);
    Logger.info('MemoryMonitor', `  时间间隔: ${growth.timeElapsed}ms`);
    Logger.info('MemoryMonitor', `  RSS 增长: ${MemoryMonitor._formatBytes(growth.rssGrowth)}`);
    Logger.info('MemoryMonitor', `  堆总大小增长: ${MemoryMonitor._formatBytes(growth.heapTotalGrowth)}`);
    Logger.info('MemoryMonitor', `  堆已用增长: ${MemoryMonitor._formatBytes(growth.heapUsedGrowth)}`);
    Logger.info('MemoryMonitor', `  外部内存增长: ${MemoryMonitor._formatBytes(growth.externalGrowth)}`);
  }

  /**
   * 强制垃圾回收（如果可用）
   * @returns {boolean} 是否成功执行垃圾回收
   */
  static forceGarbageCollection() {
    if (global.gc) {
      Logger.debug('MemoryMonitor', '执行强制垃圾回收');
      global.gc();
      return true;
    } else {
      Logger.warn('MemoryMonitor', '垃圾回收不可用，请使用 --expose-gc 标志启动 Node.js');
      return false;
    }
  }

  /**
   * 获取内存使用摘要
   * @returns {Object} 内存使用摘要
   */
  static getMemorySummary() {
    const current = MemoryMonitor.getCurrentMemoryUsage();
    const history = MemoryMonitor._memoryHistory;
    
    if (history.length === 0) {
      return {
        current,
        peak: current,
        average: current,
        samples: 0,
      };
    }
    
    const peak = history.reduce((max, record) => ({
      rss: Math.max(max.rss, record.rss),
      heapTotal: Math.max(max.heapTotal, record.heapTotal),
      heapUsed: Math.max(max.heapUsed, record.heapUsed),
      external: Math.max(max.external, record.external),
    }), { rss: 0, heapTotal: 0, heapUsed: 0, external: 0 });
    
    const total = history.reduce((sum, record) => ({
      rss: sum.rss + record.rss,
      heapTotal: sum.heapTotal + record.heapTotal,
      heapUsed: sum.heapUsed + record.heapUsed,
      external: sum.external + record.external,
    }), { rss: 0, heapTotal: 0, heapUsed: 0, external: 0 });
    
    const average = {
      rss: Math.round(total.rss / history.length),
      heapTotal: Math.round(total.heapTotal / history.length),
      heapUsed: Math.round(total.heapUsed / history.length),
      external: Math.round(total.external / history.length),
    };
    
    return {
      current,
      peak,
      average,
      samples: history.length,
    };
  }

  /**
   * 输出内存使用摘要
   */
  static logMemorySummary() {
    const summary = MemoryMonitor.getMemorySummary();
    
    Logger.info('MemoryMonitor', '内存使用摘要');
    Logger.table('MemoryMonitor', '内存使用情况', [
      {
        类型: 'RSS',
        当前: MemoryMonitor._formatBytes(summary.current.rss),
        峰值: MemoryMonitor._formatBytes(summary.peak.rss),
        平均: MemoryMonitor._formatBytes(summary.average.rss),
      },
      {
        类型: '堆总大小',
        当前: MemoryMonitor._formatBytes(summary.current.heapTotal),
        峰值: MemoryMonitor._formatBytes(summary.peak.heapTotal),
        平均: MemoryMonitor._formatBytes(summary.average.heapTotal),
      },
      {
        类型: '堆已用',
        当前: MemoryMonitor._formatBytes(summary.current.heapUsed),
        峰值: MemoryMonitor._formatBytes(summary.peak.heapUsed),
        平均: MemoryMonitor._formatBytes(summary.average.heapUsed),
      },
      {
        类型: '外部内存',
        当前: MemoryMonitor._formatBytes(summary.current.external),
        峰值: MemoryMonitor._formatBytes(summary.peak.external),
        平均: MemoryMonitor._formatBytes(summary.average.external),
      },
    ]);
    
    Logger.info('MemoryMonitor', `样本数量: ${summary.samples}`);
  }
}

/**
 * 兼容性函数，保持向后兼容
 * @param {string} label - 标签
 * @deprecated 请使用 MemoryMonitor.logMemory() 替代
 */
function LogMem(label) {
  MemoryMonitor.logMemory(label);
}

module.exports = {
  MemoryMonitor,
  LogMem, // 保持向后兼容
};
