#!/usr/bin/env node

const Translator = require("./translator");
const fs = require('fs');
const path = require('path');
const os = require('os');

// 版本信息
const VERSION = "3.0.0";

// 加载配置文件
function loadConfig() {
  const configDir = path.join(os.homedir(), '.config', 'mtran');
  const configPath = path.join(configDir, 'cli.json');
  const defaultConfig = {
    inputLang: "auto",
    outputLang: "zh-Hans"
  };

  try {
    if (fs.existsSync(configPath)) {
      const configData = fs.readFileSync(configPath, 'utf8');
      const userConfig = JSON.parse(configData);
      return { ...defaultConfig, ...userConfig };
    } else {
      // 配置文件不存在，创建默认配置
      try {
        // 确保目录存在
        if (!fs.existsSync(configDir)) {
          fs.mkdirSync(configDir, { recursive: true });
        }
        // 写入默认配置
        fs.writeFileSync(configPath, JSON.stringify(defaultConfig, null, 2), 'utf8');
        // console.log(`已创建默认配置文件: ${configPath}`);
      } catch (createError) {
        console.error(`无法创建配置文件: ${createError.message}`);
      }
    }
  } catch (error) {
    console.error(`无法加载配置文件: ${error.message}`);
  }

  return defaultConfig;
}

// 解析命令行参数
function parseArgs() {
  const args = process.argv.slice(2);
  const config = loadConfig();

  const options = {
    inputLang: config.inputLang,
    outputLang: config.outputLang,
    modelDir: config.modelDir || null, // modelDir 作为隐藏参数，仅当用户在配置中指定时使用
    text: "",
    showVersion: false,
    showHelp: false
  };

  for (let i = 0; i < args.length; i++) {
    const arg = args[i];

    if (arg === "-il" || arg === "--input-lang") {
      options.inputLang = args[++i] || config.inputLang;
    } else if (arg === "-ol" || arg === "--output-lang") {
      options.outputLang = args[++i] || config.outputLang;
    } else if (arg === "-m" || arg === "--model-dir") {
      options.modelDir = args[++i];
    } else if (arg === "-v" || arg === "--version") {
      options.showVersion = true;
    } else if (arg === "-h" || arg === "--help") {
      options.showHelp = true;
    } else if (!options.text) {
      options.text = arg;
    }
  }

  return options;
}

// 显示帮助信息
function showHelp() {
  const config = loadConfig();
  console.log(`
使用方法: mt [选项] <文本>

选项:
  -il, --input-lang <语言>    指定源语言 (默认: ${config.inputLang})
  -ol, --output-lang <语言>   指定目标语言 (默认: ${config.outputLang})
  -m, --model-dir <路径>      指定模型文件夹路径
  -v, --version              显示版本信息
  -h, --help                 显示帮助信息

配置文件:
  ~/.config/mtran/cli.json   可配置默认设置

示例:
  mt "Hello World"                    # 自动检测语言并翻译为默认语言
  mt -il en -ol ja "Hello World"      # 将英文翻译为日文
  mt -m ./models "Hello World"        # 指定模型文件夹路径
  `);
}

// 显示版本信息
function showVersion() {
  console.log(`mt 翻译工具 v${VERSION}`);
}

// 主函数
async function main() {
  const options = parseArgs();

  // 显示版本信息
  if (options.showVersion) {
    showVersion();
    return;
  }

  // 显示帮助信息
  if (options.showHelp || !options.text) {
    showHelp();
    return;
  }

  try {
    if (options.modelDir) {
      process.env.DATA_DIR = options.modelDir;
    }

    const result = await Translator.Translate(
      options.text,
      options.inputLang,
      options.outputLang
    );

    console.log(result);
    await Translator.Shutdown();
  } catch (error) {
    process.exit(1);
  }
}

main().catch(error => {
  console.error(`发生错误: ${error.message}`);
  process.exit(1);
});
