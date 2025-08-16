// 缓存 GC 可用性检查结果，避免重复检查
const isGCAvailable = typeof global.gc === "function";

/**
 * 执行垃圾回收，如果可用的话
 * @param {boolean} [force=false] - 是否强制执行完全垃圾回收
 */
function gc(force = false) {
  if (isGCAvailable) {
    try {
      // 在支持的环境中，可以通过参数控制垃圾回收的强度
      global.gc(force);
    } catch (e) {
      // 如果参数不被支持，回退到无参数调用
      global.gc();
    }
  }
}

module.exports = {
  gc,
  isGCAvailable,
};
