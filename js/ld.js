const { getLIDModel } = require("fasttext.wasm.js");

const CODEMAP = {
  // BCP47 语言代码映射，将 alpha2 为 null 的语言代码根据 alpha3 映射到相近语言
  // https://github.com/yunsii/fasttext.wasm.js/blob/master/src/models/language-identification/assets/languages.json
  gsw: "de", // Swiss German -> German
  arz: "ar", // Egyptian Arabic -> Arabic
  ast: "es", // Asturian -> Spanish
  azb: "fa", // South Azerbaijani -> Persian
  bar: "de", // Bavarian -> German
  bcl: "tl", // Central Bikol -> Tagalog
  bpy: "bn", // Bishnupriya -> Bengali
  bxr: "ru", // Russia Buriat -> Russian
  cbk: "es", // Chavacano -> Spanish
  ceb: "fi", // Cebuano -> Finnish
  zza: "tr", // Zaza -> Turkish
  dsb: "de", // Lower Sorbian -> German
  dty: "ne", // Dotyali -> Nepali
  eml: "it", // Emiliano-Romagnolo -> Italian
  frr: "de", // Northern Frisian -> German
  gom: "hi", // Goan Konkani -> Hindi
  hif: "hi", // Fiji Hindi -> Hindi
  hsb: "de", // Upper Sorbian -> German
  ilo: "fi", // Iloko -> Finnish
  jbo: "ar", // Lojban -> Arabic
  krc: "ru", // Karachay-Balkar -> Russian
  lez: "ru", // Lezghian -> Russian
  lmo: "it", // Lombard -> Italian
  lrc: "fa", // Northern Luri -> Persian
  mai: "hi", // Maithili -> Hindi
  mhr: "ru", // Eastern Mari -> Russian
  min: "id", // Minangkabau -> Indonesian
  mrj: "ru", // Western Mari -> Russian
  mwl: "pt", // Mirandese -> Portuguese
  myv: "ru", // Erzya -> Russian
  mzn: "fa", // Mazanderani -> Persian
  nah: "es", // Nahuatl languages -> Spanish
  nap: "it", // Neapolitan -> Italian
  nds: "de", // Low German -> German
  new: "ne", // Newari -> Nepali
  pam: "fi", // Pampanga -> Finnish
  pfl: "de", // Pfaelzisch -> German
  pms: "it", // Piemontese -> Italian
  pnb: "pa", // Western Panjabi -> Panjabi
  rue: "uk", // Rusyn -> Ukrainian
  sah: "ru", // Yakut -> Russian
  scn: "it", // Sicilian -> Italian
  sco: "en", // Scots -> English
  tyv: "ru", // Tuvinian -> Russian
  vec: "it", // Venetian -> Italian
  vep: "ru", // Veps -> Russian
  vls: "nl", // Vlaams -> Dutch
  war: "fi", // Waray (Philippines) -> Finnish
  wuu: "zh-Hans", // Wu Chinese -> Simplified Chinese
  xal: "ru", // Kalmyk -> Russian
  xmf: "ka", // Mingrelian -> Georgian
  yue: "zh-Hant", // Yue Chinese -> Traditional Chinese
};

const CODEMAP_KEYS = Object.keys(CODEMAP);

class LanguageDetector {
  /**
   * 检测结果对象
   * @typedef {Object} DetectionResult
   * @property {string} language - 检测到的主要语言代码，如果不确定则为 "un"
   */

  static #UNKNOWN_LANG = "un";
  static #EMPTY_STRING = "";
  static #SPACE_REGEX = /\s+/g;
  static #SPACE = " ";
  static #lidModel = null;
  static #modelInitialized = false;
  static #modelInitializing = false;

  /**
   * 初始化语言检测模型
   * @returns {Promise<void>}
   */
  static async #initModel() {
    if (this.#modelInitialized || this.#modelInitializing) return;

    try {
      this.#modelInitializing = true;
      this.#lidModel = await getLIDModel();
      await this.#lidModel.load();
      this.#modelInitialized = true;
    } catch (err) {
      console.error("Failed to initialize language detection model:", err);
      throw err;
    } finally {
      this.#modelInitializing = false;
    }
  }

  /**
   * 清理文本，移除不必要的字符以提高检测效率
   * @param {string} text - 原始文本
   * @returns {string} 清理后的文本
   */
  static #cleanText(text) {
    if (!text) return this.#EMPTY_STRING;
    return text.replace(this.#SPACE_REGEX, this.#SPACE).trim();
  }

  /**
   * 检测文本的语言
   * @param {string} text - 要检测的文本
   * @param {Object} options - 检测选项
   * @returns {Promise<DetectionResult>} 检测结果
   */
  static async detect(text, options = {}) {
    try {
      // 确保模型已初始化
      if (!this.#modelInitialized) {
        await this.#initModel();
      }

      const textToDetect = this.#cleanText(text);

      if (!textToDetect) {
        return {
          language: this.#UNKNOWN_LANG,
        };
      }

      // 使用 fasttext 进行语言识别
      const result = await this.#lidModel.identify(textToDetect);
      if (result.alpha2 === null) {
        if (result.alpha3) {
          if (CODEMAP_KEYS.includes(result.alpha3)) {
            return {
              language: CODEMAP[result.alpha3],
            };
          }
          return {
            language: this.#UNKNOWN_LANG,
          };
        }
      } else {
        return {
          language: result.alpha2,
        };
      }
    } catch (error) {
      console.error(`Language detection failed:`, error);
      return {
        language: this.#UNKNOWN_LANG,
      };
    }
  }

  /**
   * 批量检测多个文本
   * @param {string[]} texts - 要检测的文本数组
   * @param {Object} options - 检测选项
   * @returns {Promise<DetectionResult[]>} 检测结果数组
   */
  static async batchDetect(texts, options = {}) {
    // 确保模型已初始化
    if (!this.#modelInitialized) {
      await this.#initModel();
    }

    const promises = [];
    for (let i = 0; i < texts.length; i++) {
      promises.push(this.detect(texts[i], options));
    }
    return Promise.all(promises);
  }
}

module.exports = LanguageDetector;
