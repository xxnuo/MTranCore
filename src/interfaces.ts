
export interface FileSystem {
  readFile(filepath: string): Promise<Uint8Array>;
  fileExists(filepath: string): Promise<boolean>;
  joinPath(...paths: string[]): string;
}

export interface BergamotModule {
  wasmBinary?: ArrayBuffer | Buffer;
  print?: (msg: string) => void;
  printErr?: (msg: string) => void;
  onRuntimeInitialized?: () => void;
  onAbort?: (msg: string) => void;
  AlignedMemory: any;
  AlignedMemoryList: any;
  TranslationModel: any;
  BlockingService: any;
  VectorString: any;
  VectorResponseOptions: any;
  // Fallback support
  [key: string]: any;
}

export interface TranslationOptions {
  sourceLang?: string;
  targetLang?: string;
  cacheSize?: number;
}

export interface TranslateOptions {
  qualityScores?: boolean;
  alignment?: boolean;
  html?: boolean;
}

export interface TranslationConfig {
  'beam-size'?: number;
  'normalize'?: number;
  'word-penalty'?: number;
  'max-length-break'?: number;
  'mini-batch-words'?: number;
  'workspace'?: number;
  'max-length-factor'?: number;
  'skip-cost'?: boolean;
  'cpu-threads'?: number;
  'quiet'?: boolean;
  'quiet-translation'?: boolean;
  'gemm-precision'?: string;
  'alignment'?: string;
}

export interface ModelFileNames {
  model?: string;
  lex?: string;
  srcvocab?: string;
  trgvocab?: string;
}

export interface ModelBuffers {
  model: Uint8Array;
  lex: Uint8Array;
  srcvocab: Uint8Array;
  trgvocab: Uint8Array;
  qualityModel?: Uint8Array;
}
