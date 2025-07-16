const Translator = require("./translator");

// 需要将代码包装在一个异步函数中才能使用顶层await
async function usage() {
  // 打印支持的语言列表，推荐中文使用 zh-Hans 可以省去转换步骤获得最佳性能
  console.log("支持的语言列表：", Translator.GetSupportLanguages());
  // 预加载引擎，可以提高翻译性能，如果模型未预加载，则需要连接网络下载模型
  console.log("预加载英译中引擎...");
  await Translator.Preload("en", "zh-Hans");
  console.log("预加载中译英引擎...");
  await Translator.Preload("zh-Hans", "en");
  console.log("预加载日译中引擎...");
  await Translator.Preload("ja", "zh-Hans");

  console.log(
    "'Hello, world!' 翻译为中文：",
    await Translator.Translate("Hello, world!", "en", "zh-Hans")
  );
  console.log(
    "'你好，世界！' 翻译为英文：",
    await Translator.Translate("你好，世界！", "zh-Hans", "en")
  );
  console.log(
    "'こんにちは、世界！' 翻译为中文：",
    await Translator.Translate("こんにちは、世界！", "ja", "zh-Hans")
  );
  // 支持自动检测源语言，若文本太短，会检测失败，默认使用英语为源语言
  console.log(
    "'Hi!' 翻译为中文：",
    await Translator.Translate("Hi!", "auto", "zh-Hans")
  );
  // 如果使用未预加载的引擎，则需要连接网络下载模型
  console.log(
    "'Супербыстрый движок перевода' 翻译为英文：",
    await Translator.Translate("Супербыстрый движок перевода", "ru", "en")
  );

  await Translator.Shutdown();
}

usage();
