import { IPlatform, IFileSystem, IPath } from '../interface.js';

class WebPath implements IPath {
  join(...paths: string[]): string {
    return paths.join('/').replace(/\/+/g, '/');
  }

  dirname(path: string): string {
    const parts = path.split('/');
    parts.pop();
    return parts.join('/');
  }

  basename(path: string): string {
    return path.split('/').pop() || '';
  }
}

class WebFileSystem implements IFileSystem {
  private cacheName = 'mtran-models-v1';

  async readFile(path: string): Promise<Uint8Array> {
    const cache = await caches.open(this.cacheName);
    const response = await cache.match(path);
    if (!response) {
      throw new Error(`File not found in cache: ${path}`);
    }
    const buffer = await response.arrayBuffer();
    return new Uint8Array(buffer);
  }

  async writeFile(path: string, data: Uint8Array | string): Promise<void> {
    const cache = await caches.open(this.cacheName);
    const blob = new Blob([data as any]);
    const response = new Response(blob);
    await cache.put(path, response);
  }

  async exists(path: string): Promise<boolean> {
    const cache = await caches.open(this.cacheName);
    const response = await cache.match(path);
    return !!response;
  }

  async mkdir(path: string): Promise<void> {
    // Web cache doesn't need explicit directories
  }

  async unlink(path: string): Promise<void> {
    const cache = await caches.open(this.cacheName);
    await cache.delete(path);
  }
  
  async listFiles(dir: string): Promise<string[]> {
      // Not easily supported in Cache API without iterating all keys
      // For now, return empty or implement basic prefix matching if needed
      return [];
  }
}

export class WebPlatform implements IPlatform {
  fs = new WebFileSystem();
  path = new WebPath();
  name = 'web' as const;

  async fetch(url: string): Promise<Response> {
    return fetch(url);
  }

  async loadWasm(path: string): Promise<ArrayBuffer> {
    // For Web, we expect path to be a URL serving the WASM
    // Or it could be in our cache
    try {
        if (await this.fs.exists(path)) {
            const data = await this.fs.readFile(path);
            return data.buffer as ArrayBuffer;
        }
    } catch {}
    
    const response = await fetch(path);
    if (!response.ok) throw new Error(`Failed to load WASM from ${path}`);
    return response.arrayBuffer();
  }
}
