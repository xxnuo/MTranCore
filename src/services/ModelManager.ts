import { IPlatform } from '../platform/interface.js';
import * as fzstd from 'fzstd';

const DEFAULT_RECORDS_URL = 'https://firefox.settings.services.mozilla.com/v1/buckets/main-preview/collections/translations-models-v2/records';
const DEFAULT_ATTACHMENTS_BASE_URL = 'https://firefox-settings-attachments.cdn.mozilla.net';

export interface Attachment {
  hash: string;
  size: number;
  filename: string;
  location: string;
  mimetype: string;
}

export interface RecordItem {
  name: string;
  schema: number;
  version: string;
  fileType: string;
  attachment: Attachment;
  architecture?: string;
  sourceLanguage: string;
  targetLanguage: string;
  decompressedHash?: string;
  decompressedSize?: number;
  id: string;
}

export interface RecordsData {
  data: RecordItem[];
}

export interface ModelPaths {
  model: string;
  lex: string;
  vocab_src: string;
  vocab_trg: string;
}

export class ModelManager {
  private globalRecords: RecordsData | null = null;
  private recordsUrl: string;
  private attachmentsBaseUrl: string;
  
  constructor(
      private platform: IPlatform, 
      private modelDir: string,
      recordsUrl?: string,
      attachmentsBaseUrl?: string
  ) {
      this.recordsUrl = recordsUrl || DEFAULT_RECORDS_URL;
      this.attachmentsBaseUrl = attachmentsBaseUrl || DEFAULT_ATTACHMENTS_BASE_URL;
  }

  async init(): Promise<void> {
    const recordsPath = this.platform.path.join(this.modelDir, 'records.json');

    // Try to load local records first
    if (await this.platform.fs.exists(recordsPath)) {
        try {
            const data = await this.platform.fs.readFile(recordsPath);
            const text = new TextDecoder().decode(data);
            this.globalRecords = JSON.parse(text);
        } catch (e) {
            console.warn("Failed to load local records, will download.");
        }
    }

    // If not loaded or we want to update (TODO: add update logic), download
    if (!this.globalRecords) {
        await this.downloadRecords(recordsPath);
    }
  }

  private async downloadRecords(savePath: string): Promise<void> {
      const response = await this.platform.fetch(this.recordsUrl);
      if (!response.ok) throw new Error(`Failed to fetch records: ${response.statusText}`);
      const data = await response.json();
      this.globalRecords = data as RecordsData;
      
      await this.platform.fs.mkdir(this.platform.path.dirname(savePath));
      await this.platform.fs.writeFile(savePath, JSON.stringify(data));
  }


  async ensureModel(fromLang: string, toLang: string): Promise<ModelPaths> {
      if (!this.globalRecords) await this.init();
      
      // Filter records
      const matched = this.globalRecords!.data.filter(r => 
          r.sourceLanguage === fromLang && r.targetLanguage === toLang
      );
      
      if (matched.length === 0) {
          throw new Error(`No model found for ${fromLang} -> ${toLang}`);
      }

      // Group by fileType and pick latest version
      const bestRecords: RecordItem[] = [];
      const types = new Set(matched.map(r => r.fileType));
      
      for (const type of types) {
          const typeRecords = matched.filter(r => r.fileType === type);
          // Simple string sort for version usually works for "v1.0" etc, 
          // but for robustness we might want semver. For now, picking the last one (usually newest)
          // or implementing a simple version compare.
          typeRecords.sort((a, b) => a.version.localeCompare(b.version, undefined, { numeric: true }));
          bestRecords.push(typeRecords[typeRecords.length - 1]);
      }

      const langPairDir = this.platform.path.join(this.modelDir, `${fromLang}_${toLang}`);
      await this.platform.fs.mkdir(langPairDir);

      // Download files
      for (const record of bestRecords) {
          await this.downloadFile(record, langPairDir);
      }

      return this.getModelPaths(langPairDir, bestRecords);
  }

  private async downloadFile(record: RecordItem, dir: string): Promise<void> {
      const filename = record.attachment.filename;
      const decompressedFilename = filename.replace(/\.zst$/, '');
      const finalPath = this.platform.path.join(dir, decompressedFilename);

      if (await this.platform.fs.exists(finalPath)) {
          // TODO: Check hash
          return;
      }

      const url = `${this.attachmentsBaseUrl}/${record.attachment.location}`;
      console.log(`Downloading ${filename}...`);
      
      const response = await this.platform.fetch(url);
      if (!response.ok) throw new Error(`Failed to download ${filename}`);
      
      const buffer = await response.arrayBuffer();
      let data = new Uint8Array(buffer);

      if (filename.endsWith('.zst')) {
          console.log(`Decompressing ${filename}...`);
          data = fzstd.decompress(data) as any;
      }

      await this.platform.fs.writeFile(finalPath, data);
  }

  private getModelPaths(dir: string, records: RecordItem[]): ModelPaths {
      const paths: any = {};
      const typeMap: any = {
          'model': 'model',
          'lex': 'lex',
          'vocab': 'vocab', // shared vocab
          'srcvocab': 'vocab_src',
          'trgvocab': 'vocab_trg'
      };

      for (const record of records) {
          const filename = record.attachment.filename.replace(/\.zst$/, '');
          const fullPath = this.platform.path.join(dir, filename);
          const key = typeMap[record.fileType];
          if (key) {
             if (key === 'vocab') {
                 paths.vocab_src = fullPath;
                 paths.vocab_trg = fullPath;
             } else {
                 paths[key] = fullPath;
             }
          }
      }

      if (!paths.model || !paths.lex || !paths.vocab_src || !paths.vocab_trg) {
          throw new Error(`Incomplete model files for directory ${dir}`);
      }

      return paths as ModelPaths;
  }
}
