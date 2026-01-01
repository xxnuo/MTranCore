import { MTran } from '../dist/index.js';
import path from 'path';
import fs from 'fs';

async function main() {
  const modelDir = path.resolve('models_test'); // Use local test dir
  if (!fs.existsSync(modelDir)) fs.mkdirSync(modelDir);
  console.log(`Using model directory: ${modelDir}`);

  const mt = new MTran({
    modelPath: modelDir,
  });

  try {
    const text = "你好，世界！这是一个自动检测语言并翻译的测试。";
    console.log(`Original: "${text}"`);
    
    // Test 1: Auto-detection and Translation (zh -> en)
    console.log('Translating (Auto -> En)...');
    const result = await mt.translate(text, 'auto', 'en');
    console.log(`Result: "${result}"`);

    // Test 2: Pivot Translation (zh -> es, assuming we have en->es model, if not it might fail or we skip)
    // For safety, let's just reverse it: En -> Zh
    const textEn = "Hello world! This is a test.";
    console.log(`\nOriginal: "${textEn}"`);
    console.log('Translating (En -> Zh)...');
    const resultZh = await mt.translate(textEn, 'en', 'zh-Hans');
    console.log(`Result: "${resultZh}"`);

  } catch (err) {
    console.error('Error:', err);
  } finally {
      // Force cleanup if needed
      process.exit(0);
  }
}

main();
