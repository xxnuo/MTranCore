import { MTranCoreBase, MTranConfig } from './MTranCoreBase.js';
import { WebPlatform } from './platform/web/index.js';

export * from './interfaces.js';

export class MTranCore extends MTranCoreBase {
  constructor(config: MTranConfig) {
    // In Web, defaults are relative URLs
    const defaultWasmPath = 'bergamot-translator.wasm';
    const defaultCld2Path = 'cld2.wasm';

    super(config, new WebPlatform(), defaultWasmPath, defaultCld2Path);
  }
}
