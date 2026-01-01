import { IPlatform } from '../platform/interface.js';
import loadCLD2 from '../lib/cld2/cld2.js';

export interface DetectionResult {
  language: string;
  confidence: number;
}

export class LanguageDetector {
  private cldModule: any = null;
  private initPromise: Promise<void> | null = null;
  private wasmPath: string;

  constructor(private platform: IPlatform, wasmPath?: string) {
      this.wasmPath = wasmPath || 'cld2.wasm'; // Relative path handling might be tricky
  }

  async init(): Promise<void> {
    if (this.cldModule) return;
    if (this.initPromise) return this.initPromise;

    this.initPromise = (async () => {
        // Resolve WASM path. In Node, it might be an abs path. In web, a URL.
        // For simplicity, we assume the user provides a resolvable path/URL 
        // or we use a default relative to where we think assets are.
        
        let wasmBinary: ArrayBuffer;
        try {
            wasmBinary = await this.platform.loadWasm(this.wasmPath);
        } catch (e) {
            console.warn(`Failed to load CLD2 WASM from ${this.wasmPath}, detection disabled.`);
            return;
        }

        const module = await loadCLD2({
            wasmBinary,
            print: () => {},
            printErr: (msg: string) => console.error(`[CLD2]: ${msg}`)
        });
        
        if (module.LanguageInfo?.prototype?.detectLanguage) {
            module.LanguageInfo.detectLanguage = module.LanguageInfo.prototype.detectLanguage;
        }

        this.cldModule = module;
    })();

    return this.initPromise;
  }

  detect(text: string): DetectionResult {
      if (!this.cldModule) return { language: 'auto', confidence: 0 };
      
      const LanguageInfo = this.cldModule.LanguageInfo;
      const result = LanguageInfo.detectLanguage(text, true); // true = plain text
      
      const lang = result.getLanguageCode();
      const percent = result.get_languages(0).getPercent(); // approximate confidence
      
      const normalized = this.normalizeLang(lang);
      
      this.cldModule.destroy(result);
      
      return {
          language: normalized,
          confidence: percent / 100
      };
  }

  private normalizeLang(code: string): string {
      if (code === 'zh') return 'zh-Hans'; // CLD2 often returns 'zh' for simplified
      if (code === 'zh-Hant') return 'zh-Hant';
      return code;
  }
}
