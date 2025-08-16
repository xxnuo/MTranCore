#!/usr/bin/env node

import esbuild from 'esbuild';
import fs from 'fs';
// import ObfuscatorPlugin from 'esbuild-obfuscator-plugin';

async function main() {
  try {
    // 清空 dist 目录
    fs.rmSync('dist', { recursive: true, force: true });
    fs.mkdirSync('dist');
    fs.mkdirSync('dist/assets', { recursive: true });

    // 构建 JS 文件
    await esbuild.build({
      entryPoints: [
        'js/translator.js',
        'js/worker.js',
        'js/mt.js'
      ],
      outdir: 'dist',
      bundle: true,
      platform: 'node',
      format: 'cjs',
      banner: {
        js: "/* https://github.com/xxnuo */const _x_url_alias_=require('url').pathToFileURL(__filename).toString();"
      },
      define: {
        'import.meta.url': '_x_url_alias_'
      },
      minify: true,
      plugins: [
        // ObfuscatorPlugin({
        // obfuscateOutput: false,
        // filter: ['**/*.js'],
        // compact: true,
        // controlFlowFlattening: true,
        // identifierNamesGenerator: 'mangled-shuffled',
        // selfDefending: true,
        // debugProtection: true,
        // simplify: true,
        // })
      ]
    });

    console.log('JS 文件构建完成');

    // 复制文件
    fs.copyFileSync('node_modules/simplecc-wasm/pkg/nodejs/simplecc_wasm_bg.wasm', 'dist/simplecc_wasm_bg.wasm');
    fs.copyFileSync('node_modules/fasttext.wasm.js/dist/core/fastText.node.wasm', 'dist/fastText.node.wasm');
    fs.copyFileSync('node_modules/fasttext.wasm.js/dist/core/fastText.common.wasm', 'dist/fastText.common.wasm');
    fs.copyFileSync('node_modules/fasttext.wasm.js/dist/fastText/models/lid.176.ftz', 'dist/assets/lid.176.ftz');
    fs.copyFileSync('js/bergamot.wasm', 'dist/bergamot.wasm');
    fs.copyFileSync('js/translator.d.ts', 'dist/translator.d.ts');
    fs.copyFileSync('js/usage.js', 'dist/usage.js');

    console.log('构建完成');
  } catch (error) {
    console.error('构建失败:', error);
    process.exit(1);
  }
}

main();

