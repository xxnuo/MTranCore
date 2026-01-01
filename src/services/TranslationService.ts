import { TranslationEngine } from '../engine/core.js';
import { ModelManager } from './ModelManager.js';
import { LanguageDetector } from './LanguageDetector.js';
import { IPlatform } from '../platform/interface.js';
import { ResourceLoader, NodeFileSystem } from '../engine/loader.js'; // We might need to abstract ResourceLoader too or make it use IPlatform
import loadBergamot from '../lib/bergamot/bergamot-translator.js';

export class TranslationService {
  private engines = new Map<string, TranslationEngine>();
  private resourceLoader: ResourceLoader; // To be refactored to use IPlatform

  constructor(
      private platform: IPlatform,
      private modelManager: ModelManager,
      private detector: LanguageDetector,
      private bergamotWasmPath: string
  ) {
      // Temporary adapter: ResourceLoader expects FileSystem interface which is slightly different from IPlatform.fs
      // We'll stick to the existing ResourceLoader for now, but inject the IPlatform.fs implementation wrapped.
      // Actually, let's just use IPlatform in TranslationEngine init if possible, 
      // OR adapt IPlatform.fs to the FileSystem interface expected by ResourceLoader.
      
      const fsAdapter = {
          readFile: async (p: string) => {
              const data = await this.platform.fs.readFile(p);
              return data; // Now expecting Uint8Array, which IPlatform returns
          },
          fileExists: (p: string) => this.platform.fs.exists(p),
          joinPath: (...args: string[]) => this.platform.path.join(...args)
      };
      
      this.resourceLoader = new ResourceLoader(fsAdapter);
  }

  async translate(text: string, from: string, to: string, isHTML: boolean = false): Promise<string> {
      if (!text.trim()) return "";
      
      let sourceLang = from;
      if (sourceLang === 'auto') {
          await this.detector.init();
          const detected = this.detector.detect(text);
          sourceLang = detected.language;
          if (sourceLang === 'un' || !sourceLang) sourceLang = 'en'; // Fallback
      }

      if (sourceLang === to) return text;

      // Pivot logic
      if (this.needsPivot(sourceLang, to)) {
          const intermediate = await this.translateOneStep(text, sourceLang, 'en', isHTML);
          return this.translateOneStep(intermediate, 'en', to, isHTML);
      } else {
          return this.translateOneStep(text, sourceLang, to, isHTML);
      }
  }

  private needsPivot(from: string, to: string): boolean {
      if (from === 'en' || to === 'en') return false;
      // TODO: Check if direct model exists in registry
      return true; // Default to pivot for non-English pairs
  }

  private async translateOneStep(text: string, from: string, to: string, isHTML: boolean): Promise<string> {
      const key = `${from}-${to}`;
      let engine = this.engines.get(key);
      
      if (!engine) {
          engine = await this.createEngine(from, to);
          this.engines.set(key, engine);
      }

      return engine.translateAsync(text, { html: isHTML });
  }

  private async createEngine(from: string, to: string): Promise<TranslationEngine> {
      // 1. Ensure model exists
      const modelPaths = await this.modelManager.ensureModel(from, to);

      // 2. Load WASM
      const wasmBinary = await this.platform.loadWasm(this.bergamotWasmPath);
      // ResourceLoader now accepts ArrayBuffer or Uint8Array
      
      const bergamotModule = await this.resourceLoader.loadBergamotModule(new Uint8Array(wasmBinary), loadBergamot);

      // 3. Load Model Buffers
      // We can reuse resourceLoader.loadModelFiles logic but we have absolute paths now
      const modelBuffers = {
          model: await this.platform.fs.readFile(modelPaths.model),
          lex: await this.platform.fs.readFile(modelPaths.lex),
          srcvocab: await this.platform.fs.readFile(modelPaths.vocab_src),
          trgvocab: await this.platform.fs.readFile(modelPaths.vocab_trg)
      };

      // 4. Init Engine
      const engine = new TranslationEngine({ sourceLang: from, targetLang: to });
      // @ts-ignore - Buffer vs Uint8Array mismatch in types, but runtime handles it
      await engine.init(bergamotModule, modelBuffers);
      
      return engine;
  }
}
