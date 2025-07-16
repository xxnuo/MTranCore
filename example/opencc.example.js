const OpenCC = require("../js/opencc.js");

async function cc(text, type) {
  const start = performance.now();
  const result = await OpenCC.convert(text, type);
  const end = performance.now();
  console.log(
    `${type} ${(end - start).toFixed(2)}ms\n${text}\n->\n${result}\n`
  );
}

async function testOpencc() {
  try {
    await OpenCC.warmup();
    // await cc("简体到繁体：\n世间无限丹青手 一片伤心画不成", "s2t");
    // await cc("繁体到简体：\n落葉他鄉樹 寒燈獨夜人", "t2s");
    // await cc("简体到台湾繁体：\n我见青山多妩媚 料青山见我应如是", "s2tw");
    await cc("简体到香港繁体：\n昔去雪如花 今来花似雪", "s2hk");
    await cc("香港繁体到简体：\n位卑未敢忘憂國 事定猶須待闔棺", "hk2s");
    await cc(
      "简体到台湾繁体并转换为台湾常用词条：\n只身千里客 孤枕一灯秋",
      "s2twp"
    );
    await cc(
      "台湾繁体到简体并转换为大陆常用词条：\n位卑未敢忘憂國 事定猶須待闔棺",
      "tw2sp"
    );
    await cc(
      "台湾繁体到简体并转换为大陆常用词条：\n我哋谂住晏昼去街",
      "tw2sp"
    );
  } catch (error) {
    console.error("Error in test:", error);
  }
}

testOpencc();
