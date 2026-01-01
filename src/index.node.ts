import { MTranCoreBase, MTranConfig } from './MTranCoreBase.js';
import { NodePlatform } from './platform/node/index.js';
import { dirname, join } from 'path';
import { fileURLToPath } from 'url';

export * from './interfaces.js';

export class MTranCore extends MTranCoreBase {
  constructor(config: MTranConfig) {
    const __dirname = dirname(fileURLToPath(import.meta.url));
    const defaultWasmPath = join(__dirname, 'lib/bergamot/bergamot-translator.wasm');
    const defaultCld2Path = join(__dirname, 'lib/cld2/cld2.wasm');

    super(config, new NodePlatform(), defaultWasmPath, defaultCld2Path);
  }
}
