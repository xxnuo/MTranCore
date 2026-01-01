import { MTran } from '../dist/index.js';
import path from 'path';

async function main() {
  const modelPath = path.resolve('../MTranServer/models/en_zh-Hans');
  console.log(`Loading models from: ${modelPath}`);

  const mt = new MTran({
    modelPath: modelPath,
    // Ensure we look for the specific file names present in MTranServer
    modelFileNames: {
      model: 'model.enzh.intgemm.alphas.bin',
      lex: 'lex.50.50.enzh.s2t.bin',
      srcvocab: 'srcvocab.enzh.spm',
      trgvocab: 'trgvocab.enzh.spm'
    }
  });

  try {
    console.log('Initializing engine...');
    await mt.init('en', 'zh-Hans');
    console.log('Engine initialized.');

    const text = "Hello world! This is a test of MTran Core.";
    console.log(`Translating: "${text}"`);
    
    const result = await mt.translate(text);
    console.log(`Result: "${result}"`);

  } catch (err) {
    console.error('Error:', err);
  } finally {
    // Keep alive briefly to ensure no pending async ops? No, just exit.
  }
}

main();
