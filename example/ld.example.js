const LanguageDetector = require("../js/ld.js");

async function d(t) {
  console.log("--------------------------------");
  console.log("文本:", t);
  const start = performance.now();
  const result = await LanguageDetector.detect(t);
  const end = performance.now();
  console.log("检测结果:", result.language);
  console.log(`检测耗时: ${(end - start).toFixed(2)}ms`);
}

async function testLanguageDetection() {
  try {
    const englishText = "The quick brown fox jumps over the lazy dog.";
    await d(englishText);
    const chineseText =
      "机器翻译是计算机科学的一个分支，致力于开发能够将文本从一种语言翻译成另一种语言的软件。";
    await d(chineseText);
    const japaneseText =
      "機械翻訳はコンピュータ科学の一分野で、テキストをある言語から別の言語に翻訳するソフトウェアを開発することを目的としています。";
    await d(japaneseText);
    const koreanText =
      "기계 번역은 컴퓨터 과학의 한 분야로, 텍스트를 한 언어에서 다른 언어로 번역하는 소프트웨어를 개발하는 것을 목적으로 합니다.";
    await d(koreanText);
    const yueText = "我哋谂住晏昼去街。";
    await d(yueText);
    const yueText2 = "佢话嚟我度。";
    await d(yueText2);
    const htmlText =
      "<div><p>This is English text</p><p>这是中文文本</p></div>";
    await d(htmlText);
    const mixedText = "This is English. 这是中文。 Esto es español.";
    await d(mixedText);
    const mixedHtmlText =
      "<div><p>This is English text</p><p>这是中文文本</p><p>Esto es español</p></div>";
    await d(mixedHtmlText);
    const mixedHtmlText2 =
      "<div><p>This is English text</p><p>这是中文文本</p><p>Esto es español</p><p>これは日本語のテキストです</p><p>이것은 한국어 텍스트입니다</p></div>";
    await d(mixedHtmlText2);
  } catch (error) {
    console.error("测试过程中出错:", error);
  } finally {
    // 确保脚本执行完成后退出
    process.exit(0);
  }
}

// 执行测试
testLanguageDetection();
