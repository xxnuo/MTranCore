import { TranslationService } from './services/TranslationService.js';
import { ModelManager } from './services/ModelManager.js';
import { LanguageDetector } from './services/LanguageDetector.js';
import { IPlatform } from './platform/interface.js';

export * from './interfaces.js';

export interface MTranConfig {
  modelPath: string;
  wasmPath?: string;
  cld2Path?: string;
  worker?: boolean;
  recordsUrl?: string;
  attachmentsBaseUrl?: string;
}

export class MTranCoreBase {
  protected service: TranslationService;
  protected platform: IPlatform;
  protected modelManager: ModelManager;
  protected detector: LanguageDetector;

  constructor(config: MTranConfig, platform: IPlatform, defaultWasmPath: string, defaultCld2Path: string) {
    this.platform = platform;
    
    const wasmPath = config.wasmPath || defaultWasmPath;
    const cld2Path = config.cld2Path || defaultCld2Path;

    this.modelManager = new ModelManager(
        this.platform, 
        config.modelPath,
        config.recordsUrl,
        config.attachmentsBaseUrl
    );
    this.detector = new LanguageDetector(this.platform, cld2Path);
    
    this.service = new TranslationService(
        this.platform, 
        this.modelManager, 
        this.detector,
        wasmPath
    );
  }

  async init(from: string, to: string) {
      await this.modelManager.ensureModel(from, to);
  }

  async translate(text: string, from: string = 'auto', to: string = 'en', html: boolean = false): Promise<string> {
      return this.service.translate(text, from, to, html);
  }
}
