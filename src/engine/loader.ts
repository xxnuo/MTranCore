import { FileSystem, BergamotModule, ModelFileNames, ModelBuffers } from '../interfaces.js';
import * as fs from 'fs/promises';
import * as path from 'path';

export class NodeFileSystem implements FileSystem {
  async readFile(filepath: string): Promise<Uint8Array> {
    return fs.readFile(filepath);
  }

  async fileExists(filepath: string): Promise<boolean> {
    try {
      await fs.access(filepath);
      return true;
    } catch {
      return false;
    }
  }

  joinPath(...paths: string[]): string {
    return path.join(...paths);
  }
}

export class ResourceLoader {
  constructor(private fileSystem: FileSystem) {}

  async loadBergamotModule(wasmBinary: ArrayBuffer | Uint8Array, loadBergamot: any): Promise<BergamotModule> {
    return new Promise((resolve, reject) => {
      const timeout = setTimeout(() => {
        reject(new Error('WASM initialization timeout'));
      }, 30000);

      const moduleConfig: any = {
        wasmBinary: wasmBinary,
        print: (msg: string) => {}, // Silence stdout by default
        printErr: (msg: string) => console.error(`[Bergamot Error]: ${msg}`),
        onRuntimeInitialized: function(this: BergamotModule) {
          clearTimeout(timeout);
          resolve(this);
        },
        onAbort: (msg: string) => {
          clearTimeout(timeout);
          reject(new Error(`WASM aborted: ${msg}`));
        }
      };

      loadBergamot(moduleConfig);
    });
  }

  async loadModelFiles(modelPath: string, fileNames: ModelFileNames | null = null): Promise<ModelBuffers> {
    const defaultFiles: Required<ModelFileNames> = {
      model: 'model.enzh.intgemm.alphas.bin',
      lex: 'lex.50.50.enzh.s2t.bin',
      srcvocab: 'srcvocab.enzh.spm',
      trgvocab: 'trgvocab.enzh.spm'
    };

    const files = { ...defaultFiles, ...fileNames };
    const buffers: Partial<ModelBuffers> = {};

    for (const [key, filename] of Object.entries(files)) {
      if (!filename) continue;
      const filepath = this.fileSystem.joinPath(modelPath, filename);
      if (await this.fileSystem.fileExists(filepath)) {
        buffers[key as keyof ModelBuffers] = await this.fileSystem.readFile(filepath);
      } else {
        throw new Error(`Model file not found: ${filepath} (Key: ${key})`);
      }
    }

    if (!buffers.model || !buffers.lex || !buffers.srcvocab || !buffers.trgvocab) {
      throw new Error('Missing required model files (model, lex, srcvocab, trgvocab)');
    }

    return buffers as ModelBuffers;
  }

  async loadWasmBinary(wasmPath: string): Promise<Uint8Array> {
    if (await this.fileSystem.fileExists(wasmPath)) {
      return this.fileSystem.readFile(wasmPath);
    }
    throw new Error(`WASM binary not found at: ${wasmPath}`);
  }
}
