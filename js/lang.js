// 模型支持的语言列表，更新日期：2025-07-12

// Language Model can translate to English: 支持从其他语言翻译到英文的模型列表
class Lang {
  static MC2E = [
    "cs",
    "fi",
    "id",
    "bg",
    "ml",
    "ro",
    "he",
    "hi",
    "de",
    "kn",
    "sk",
    "hr",
    "uk",
    "es",
    "el",
    "pt",
    "ta",
    "it",
    "ca",
    "tr",
    "sl",
    "nl",
    "bn",
    "gu",
    "sv",
    "lt",
    "hu",
    "et",
    "fr",
    "ru",
    "pl",
    "sq",
    "fa",
    "ms",
    "da",
    "te",
    "lv",
    "ko",
    "zh-Hans",
    "uk",
    "ja",
    "ar",
  ];
  // Language Model can translate from English: 支持从英文翻译到其他语言的模型列表
  static MEC2 = [
    "is",
    "sv",
    "bs",
    "fr",
    "fa",
    "gu",
    "de",
    "nl",
    "sk",
    "sr",
    "bg",
    "ca",
    "ml",
    "hu",
    "he",
    "id",
    "cs",
    "pt",
    "it",
    "tr",
    "uk",
    "bn",
    "be",
    "ro",
    "hi",
    "kn",
    "lv",
    "et",
    "sq",
    "nn",
    "mt",
    "ms",
    "da",
    "hr",
    "lt",
    "az",
    "vi",
    "el",
    "ta",
    "te",
    "pl",
    "sl",
    "zh-Hans",
    "ja",
    "ar",
    "ko",
  ];

  // Language Model alias: 模型支持的语言代码的别名，Key 是语言代码，Value 是别名
  static MALIAS = {
    "zh-CN": "zh-Hans",
  };

  // Language Model can translate to zh: 可以由 OpenCC 转换为 zh 然后翻译的语言代码列表
  static MC2ZH = ["zh-Hant", "zh-TW", "zh-HK"];
  static MCC = ["zh-Hans", "zh-Hant", "zh-TW", "zh-HK"];
  // Language Model All: 全部模型支持的语言列表，去重
  static MALL = [
    "en",
    ...new Set([
      ...Lang.MC2E,
      ...Lang.MEC2,
      ...Lang.MC2ZH,
      ...Object.keys(Lang.MALIAS),
    ]),
  ];

  // OpenCC 转换配置

  // 简体到繁体转换类型
  static ZH2CC = {
    "zh-Hant": "s2twp", // 繁体到台湾繁体并转换为台湾常用词条
    "zh-TW": "s2twp", // 简体到台湾繁体并转换为台湾常用词条
    "zh-HK": "s2hk", // 简体到香港繁体
  };
  // 繁体到简体转换类型
  static CC2ZH = {
    "zh-Hant": "tw2sp", // 台湾繁体到简体并转换为大陆常用词条
    "zh-TW": "tw2sp", // 台湾繁体到简体并转换为大陆常用词条
    "zh-HK": "hk2s", // 香港繁体到简体
  };

  // LD 转换配置
  static LD = ["zh"];
  static LD2M = {
    zh: "zh-Hans",
  };
}
module.exports = Lang;
