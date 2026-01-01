import { MTranCore } from '../../dist/index.node.js'; // Direct import for example
import path from 'path';
import fs from 'fs';
import { fileURLToPath } from 'url';

// Helper for ESM __dirname
const __dirname = path.dirname(fileURLToPath(import.meta.url));

async function main() {
  // 1. Configure where to store/load translation models
  const modelDir = path.join(__dirname, 'models');
  if (!fs.existsSync(modelDir)) {
    fs.mkdirSync(modelDir, { recursive: true });
  }

  console.log('🚀 MTran Node.js Example');
  console.log(`📂 Model Directory: ${modelDir}`);

  // 2. Initialize MTran
  const mtran = new MTranCore({
    modelPath: modelDir,
    // Node.js automatically finds the bundled WASM files, so no need to specify wasmPath/cld2Path
  });

  try {
    // 3. Define input
    const text = "Hello world! This library runs purely on local devices.";
    const fromLang = "en";
    const toLang = "zh-Hans";

    console.log(`
📝 Input: "${text}"`);
    console.log(`🔄 Translating ${fromLang} -> ${toLang}...`);

    // 4. Perform Translation
    // The first time, this will automatically download the necessary models (~30MB)
    const result = await mtran.translate(text, fromLang, toLang);

    console.log(`
✅ Result: "${result}"`);

    // 5. Auto-detect example
    const mixedText = "Bonjour le monde! 这是自动检测测试。";
    console.log(`
📝 Mixed Input: "${mixedText}"`);
    console.log(`🔄 Translating Auto -> English...`);
    
    // 'auto' triggers the CLD2 language detector
    const resultAuto = await mtran.translate(mixedText, 'auto', 'en');
    console.log(`✅ Result: "${resultAuto}"`);

  } catch (error) {
    console.error('❌ Error:', error);
  }
}

main();