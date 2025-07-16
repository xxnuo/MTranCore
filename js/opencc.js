const { simplecc } = require("simplecc-wasm");

// 预热simplecc模块，提高首次转换速度
async function warmup() {
  try {
    simplecc("MTranCore 预热文本", "s2t");
    simplecc("MTranCore 預熱文本", "t2s");
    console.log("OpenCC module warmed up successfully");
  } catch (err) {
    console.warn("OpenCC warm-up failed:", err);
  }
}

module.exports = {
  convert: simplecc,
  warmup,
};
