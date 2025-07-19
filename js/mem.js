/**
 * 记录内存使用情况
 */
function LogMem(label) {
  const memUsage = process.memoryUsage();
  console.log(`[内存使用 - ${label}]`);
  console.log(`  RSS: ${(memUsage.rss / 1024 / 1024).toFixed(2)} MB`);
  console.log(
    `  堆总大小: ${(memUsage.heapTotal / 1024 / 1024).toFixed(2)} MB`
  );
  console.log(`  堆已用: ${(memUsage.heapUsed / 1024 / 1024).toFixed(2)} MB`);
  console.log(`  外部: ${(memUsage.external / 1024 / 1024).toFixed(2)} MB`);
}

module.exports = {
  LogMem,
};
