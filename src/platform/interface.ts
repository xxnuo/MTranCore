export interface IFileSystem {
  readFile(path: string): Promise<Uint8Array>;
  writeFile(path: string, data: Uint8Array | string): Promise<void>;
  exists(path: string): Promise<boolean>;
  mkdir(path: string): Promise<void>;
  unlink(path: string): Promise<void>;
  listFiles(dir: string): Promise<string[]>;
}

export interface IPath {
  join(...paths: string[]): string;
  dirname(path: string): string;
  basename(path: string): string;
}

export interface IPlatform {
  fs: IFileSystem;
  path: IPath;
  fetch(url: string): Promise<Response>;
  loadWasm(path: string): Promise<ArrayBuffer>;
  // Environment identifier
  name: 'node' | 'web';
}
