/**
 * 翻译引擎管理器类型定义
 * 
 * 使用示例:
 * ```javascript
 * const Translator = require('./translator');
 * 
 * // 获取支持的语言
 * const languages = Translator.GetSupportLanguages();
 * console.log('支持的语言:', languages);
 * 
 * // 翻译单个文本
 * async function translateText() {
 *   const result = await Translator.Translate('Hello world', 'en', 'zh-Hans');
 *   console.log('翻译结果:', result);
 * }
 * 
 * // 批量翻译文本
 * async function translateMultipleTexts() {
 *   const texts = ['Hello', 'Good morning', 'Thank you'];
 *   const results = await Translator.Translate(texts, 'en', 'zh-Hans');
 *   console.log('批量翻译结果:', results);
 * }
 * 
 * // 自动检测语言
 * async function autoDetectAndTranslate() {
 *   const result = await Translator.Translate('こんにちは', 'auto', 'zh-Hans');
 *   console.log('自动检测并翻译:', result);
 * }
 * 
 * // 使用完毕后关闭引擎
 * async function shutdown() {
 *   await Translator.Shutdown();
 *   console.log('翻译引擎已关闭');
 * }
 * ```
 */
declare module "./translator" {
    /**
     * 翻译引擎管理器
     */
    class Translator {
        /**
         * 获取支持的语言列表
         * @returns 支持的语言代码数组
         */
        static GetSupportLanguages(): string[];

        /**
         * 预加载翻译引擎
         * @param fromLang 源语言代码
         * @param toLang 目标语言代码
         * @returns 翻译引擎对象
         */
        static Preload(
            fromLang: string,
            toLang: string
        ): Promise<{
            translate: (texts: string | string[], isHTML?: boolean) => Promise<string | string[]>;
            discardTranslations: () => void;
        }>;

        /**
         * 关闭所有翻译引擎
         */
        static Shutdown(): Promise<void>;

        /**
         * 检测文本语言
         * @param text 要检测的文本
         * @returns 检测到的语言代码
         */
        static DetectLang(text: string): Promise<string>;

        /**
         * 翻译文本
         * @param text 要翻译的文本或文本数组
         * @param fromLang 源语言代码，使用"auto"表示自动检测
         * @param toLang 目标语言代码
         * @param isHTML 是否为HTML文本，默认为false
         * @returns 翻译后的文本或文本数组
         */
        static Translate(
            text: string | string[],
            fromLang: string,
            toLang: string,
            isHTML?: boolean
        ): Promise<string | string[]>;
    }

    export = Translator;
} 