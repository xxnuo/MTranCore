#!/usr/bin/env node

const Translator = require("./translator");
const fs = require("fs");
const path = require("path");
const os = require("os");

// 版本信息
const VERSION = "3.0.0";

// 配置管理类
class ConfigManager {
  static CONFIG_DIR = path.join(os.homedir(), ".config", "mtran");
  static CONFIG_PATH = path.join(ConfigManager.CONFIG_DIR, "cli.json");
  static DEFAULT_CONFIG = {
    inputLang: "auto",
    outputLang: "zh-Hans",
  };

  /**
   * 加载配置文件
   * @returns {Object} 配置对象
   */
  static loadConfig() {
    try {
      if (fs.existsSync(ConfigManager.CONFIG_PATH)) {
        const configData = fs.readFileSync(ConfigManager.CONFIG_PATH, "utf8");
        const userConfig = JSON.parse(configData);
        return { ...ConfigManager.DEFAULT_CONFIG, ...userConfig };
      } else {
        // 配置文件不存在，创建默认配置
        ConfigManager.createDefaultConfig();
      }
    } catch (error) {
      console.error(`无法加载配置文件: ${error.message}`);
    }

    return ConfigManager.DEFAULT_CONFIG;
  }

  /**
   * 创建默认配置文件
   */
  static createDefaultConfig() {
    try {
      // 确保目录存在
      if (!fs.existsSync(ConfigManager.CONFIG_DIR)) {
        fs.mkdirSync(ConfigManager.CONFIG_DIR, { recursive: true });
      }

      // 写入默认配置
      fs.writeFileSync(
        ConfigManager.CONFIG_PATH,
        JSON.stringify(ConfigManager.DEFAULT_CONFIG, null, 2),
        "utf8"
      );
    } catch (createError) {
      console.error(`无法创建配置文件: ${createError.message}`);
    }
  }
}

// 命令行参数解析器
class ArgumentParser {
  /**
   * 解析命令行参数
   * @returns {Object} 解析后的选项
   */
  static parseArgs() {
    const args = process.argv.slice(2);
    const config = ConfigManager.loadConfig();

    const options = {
      inputLang: config.inputLang,
      outputLang: config.outputLang,
      modelDir: config.modelDir || null,
      text: "",
      showVersion: false,
      showHelp: false,
    };

    for (let i = 0; i < args.length; i++) {
      const arg = args[i];

      switch (arg) {
        case "-il":
        case "--input-lang":
          options.inputLang = args[++i] || config.inputLang;
          break;
        case "-ol":
        case "--output-lang":
          options.outputLang = args[++i] || config.outputLang;
          break;
        case "-m":
        case "--model-dir":
          options.modelDir = args[++i];
          break;
        case "-v":
        case "--version":
          options.showVersion = true;
          break;
        case "-h":
        case "--help":
          options.showHelp = true;
          break;
        default:
          if (!options.text && !arg.startsWith("-")) {
            options.text = arg;
          }
          break;
      }
    }

    return options;
  }
}

// 帮助和版本信息显示器
class InfoDisplay {
  /**
   * 显示帮助信息
   */
  static showHelp() {
    const config = ConfigManager.loadConfig();
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

  /**
   * 显示版本信息
   */
  static showVersion() {
    console.log(`mt 翻译工具 v${VERSION}`);
  }
}

// 翻译执行器
class TranslationExecutor {
  /**
   * 执行翻译任务
   * @param {Object} options - 翻译选项
   */
  static async execute(options) {
    try {
      // 设置模型目录
      if (options.modelDir) {
        process.env.DATA_DIR = options.modelDir;
      }

      // 执行翻译
      const result = await Translator.Translate(
        options.text,
        options.inputLang,
        options.outputLang
      );

      console.log(result);

      // 清理资源
      await Translator.Shutdown();
    } catch (error) {
      console.error(`翻译过程中发生错误: ${error.message}`);
      throw error;
    }
  }
}

// 主函数
async function main() {
  try {
    const options = ArgumentParser.parseArgs();

    // 显示版本信息
    if (options.showVersion) {
      InfoDisplay.showVersion();
      return;
    }

    // 显示帮助信息或缺少文本参数
    if (options.showHelp || !options.text) {
      InfoDisplay.showHelp();
      return;
    }

    // 执行翻译
    await TranslationExecutor.execute(options);
  } catch (error) {
    console.error(`发生错误: ${error.message}`);
    process.exit(1);
  }
}

// 全局错误处理
process.on("uncaughtException", (error) => {
  console.error(`未捕获的异常: ${error.message}`);
  process.exit(1);
});

process.on("unhandledRejection", (reason, promise) => {
  console.error(`未处理的Promise拒绝:`, reason);
  process.exit(1);
});

// 启动应用
main().catch((error) => {
  console.error(`启动失败: ${error.message}`);
  process.exit(1);
});
