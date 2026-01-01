import { IPlatform, IFileSystem, IPath } from '../interface.js';
import * as fs from 'fs/promises';
import * as path from 'path';

class NodeFileSystem implements IFileSystem {
  async readFile(filePath: string): Promise<Uint8Array> {
    return fs.readFile(filePath);
  }

  async writeFile(filePath: string, data: Uint8Array | string): Promise<void> {
    await fs.writeFile(filePath, data);
  }

  async exists(filePath: string): Promise<boolean> {
    try {
      await fs.access(filePath);
      return true;
    } catch {
      return false;
    }
  }

  async mkdir(dirPath: string): Promise<void> {
    await fs.mkdir(dirPath, { recursive: true });
  }

  async unlink(filePath: string): Promise<void> {
    await fs.unlink(filePath);
  }
  
  async listFiles(dir: string): Promise<string[]> {
      try {
          return await fs.readdir(dir);
      } catch (e) {
          return [];
      }
  }
}

class NodePath implements IPath {
  join(...paths: string[]): string {
    return path.join(...paths);
  }

  dirname(p: string): string {
    return path.dirname(p);
  }

  basename(p: string): string {
    return path.basename(p);
  }
}

export class NodePlatform implements IPlatform {
  fs = new NodeFileSystem();
  path = new NodePath();
  name = 'node' as const;

  async fetch(url: string): Promise<Response> {
    return fetch(url);
  }

  async loadWasm(filePath: string): Promise<ArrayBuffer> {
    const buffer = await fs.readFile(filePath);
    return buffer.buffer.slice(buffer.byteOffset, buffer.byteOffset + buffer.byteLength);
  }
}
