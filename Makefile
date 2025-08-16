.PHONY: env prepare-git prepare-js prepare-py update patch benchmark benchmark-sub build pack publish

env:
	git lfs install
	if [ ! -d "packages/emsdk" ]; then \
		git clone https://github.com/emscripten-core/emsdk.git packages/emsdk; \
	fi
	cd packages/emsdk && git pull && ./emsdk install latest && ./emsdk activate latest
	if xmake --version > /dev/null 2>&1; then \
		echo "xmake found"; \
	else \
		echo "xmake not found, please install it first"; \
	fi
	@echo "Run source ./packages/emsdk/emsdk_env.sh to activate emsdk"

prepare-git:
	mkdir -p packages
	if [ ! -d "packages/firefox" ]; then git clone https://github.com/mozilla-firefox/firefox.git packages/firefox; fi
	cd packages/firefox && git pull
	if [ ! -d "packages/firefox-translations-models" ]; then git clone https://github.com/mozilla/firefox-translations-models.git packages/firefox-translations-models; fi
	cd packages/firefox-translations-models && git pull
	if [ ! -d "packages/simplecc-wasm" ]; then git clone https://github.com/xxnuo/simplecc-wasm.git packages/simplecc-wasm; fi
	cd packages/simplecc-wasm && git submodule update --init --recursive
	if [ ! -d "packages/fasttext.wasm.js" ]; then git clone https://github.com/yunsii/fasttext.wasm.js.git packages/fasttext.wasm.js; fi
	cd packages/fasttext.wasm.js && git pull && git submodule update --init --recursive

prepare-js:
	pnpm install
	if xmake --version > /dev/null 2>&1 && emcc --help > /dev/null 2>&1; then \
		source packages/emsdk/emsdk_env.sh && \
		cd packages/fasttext.wasm.js && pnpm install && pnpm run build; \
	else \
		echo "xmake or emsdk not found, please install or activate them first"; \
	fi
	pnpm install ./packages/fasttext.wasm.js
	cd packages/simplecc-wasm && pnpm install && pnpm run build:cargo && pnpm run build
	pnpm install ./packages/simplecc-wasm

prepare-py:
	if psrecord --help > /dev/null 2>&1; then \
		echo "psrecord found"; \
	else \
		echo "psrecord not found, please install it first, for example 'uv tool install psrecord --with matplotlib'"; \
		exit 1; \
	fi

update: prepare-git prepare-js prepare-py

patch: update
	@echo "下载 bergamot 二进制文件..."
	grep -A 5 "translations.inference:" packages/firefox/taskcluster/kinds/fetch/translations-fetch.yml | grep "url:" | awk '{print $$2}' | grep "\.wasm$$" | xargs curl -L -o js/bergamot.wasm
	@echo "复制 bergamot.js..."
	cp packages/firefox/toolkit/components/translations/bergamot-translator/bergamot-translator.js js/bergamot.js
	@echo "更新 bergamot.js..."
	perl -i -pe 's/bergamot-translator\.wasm/bergamot.wasm/g' js/bergamot.js
	echo "" >> js/bergamot.js
	echo "module.exports = loadBergamot;" >> js/bergamot.js
	@echo "复制翻译类型..."
	cp packages/firefox/toolkit/components/translations/translations.d.ts translations.d.ts

benchmark: prepare-py
	rm -f benchmark.html
	mkdir -p benchmark/logs
	mkdir -p benchmark/plots
	@echo "开始性能基准测试..."
	@echo "<!DOCTYPE html>" > benchmark.html
	@echo "<html lang=\"en\">" >> benchmark.html
	@echo "<head>" >> benchmark.html
	@echo "  <meta charset=\"UTF-8\">" >> benchmark.html
	@echo "  <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">" >> benchmark.html
	@echo "  <title>Benchmark Results</title>" >> benchmark.html
	@echo "  <style>" >> benchmark.html
	@echo "    .benchmark-container { display: flex; flex-wrap: wrap; gap: 20px; }" >> benchmark.html
	@echo "    .benchmark-item { flex: 1; min-width: 30%; }" >> benchmark.html
	@echo "    .benchmark-image { width: 100%; max-width: 500px; }" >> benchmark.html
	@echo "  </style>" >> benchmark.html
	@echo "</head>" >> benchmark.html
	@echo "<body>" >> benchmark.html

	$(MAKE) benchmark-sub name=engine.example.js desc="Machine Translation WASM Engine (CPU)"
	$(MAKE) benchmark-sub name=ld.example.js desc="Language Code Detection (CPU)"
	$(MAKE) benchmark-sub name=models.example.js desc="Model File Download Detection (IO)"
	$(MAKE) benchmark-sub name=opencc.example.js desc="OpenCC WASM Conversion (CPU)"
	$(MAKE) benchmark-sub name=translator.example.js desc="Machine Translation Engine (CPU + IO)"
	
	@echo "</body>" >> benchmark.html
	@echo "</html>" >> benchmark.html
	@echo "基准测试完成，结果保存在 benchmark 目录"

benchmark-sub:
	@echo "测试 $(name) 在 Node.js 环境下的性能..."
	uv run psrecord --plot ./benchmark/plots/$(name).node.png --include-children --log ./benchmark/logs/$(name).node.log "node --expose-gc ./example/$(name)"
	
	@echo "测试 $(name) 在 Bun 环境下的性能..."
	uv run psrecord --plot ./benchmark/plots/$(name).bun.png --include-children --log ./benchmark/logs/$(name).bun.log "bun ./example/$(name)"

	@echo "测试 $(name) 在 Deno 环境下的性能..."
	#uv run psrecord --plot ./benchmark/plots/$(name).deno.png --include-children --log ./benchmark/logs/$(name).deno.log "deno run --allow-all ./example/$(name)"
	@echo "Deno 在语言检测模块加载有兼容性问题，暂不测试"
	
	@echo "  <h2>$(desc) - $(name)</h2>" >> benchmark.html
	@echo "  <div class=\"benchmark-container\">" >> benchmark.html
	@echo "    <div class=\"benchmark-item\">" >> benchmark.html
	@echo "      <h3>Node.js</h3>" >> benchmark.html
	@echo "      <img src=\"./benchmark/plots/$(name).node.png\" alt=\"Node $(name)\" class=\"benchmark-image\">" >> benchmark.html
	@echo "    </div>" >> benchmark.html
	@echo "    <div class=\"benchmark-item\">" >> benchmark.html
	@echo "      <h3>Bun</h3>" >> benchmark.html
	@echo "      <img src=\"./benchmark/plots/$(name).bun.png\" alt=\"Bun $(name)\" class=\"benchmark-image\">" >> benchmark.html
	@echo "    </div>" >> benchmark.html
	# @echo "    <div class=\"benchmark-item\">" >> benchmark.html
	# @echo "      <h3>Deno</h3>" >> benchmark.html
	# @echo "      <img src=\"./benchmark/plots/$(name).deno.png\" alt=\"Deno $(name)\" class=\"benchmark-image\">" >> benchmark.html
	# @echo "    </div>" >> benchmark.html
	@echo "  </div>" >> benchmark.html
	@echo "$(name) 测试完成"

build:
	node build.mjs

pack: build
	rm -rf release
	mkdir -p release/latest
	cp package.json .package.json.bak
	node -e "const pkg = require('./package.json'); \
		delete pkg.devDependencies; \
		delete pkg.scripts; \
		require('fs').writeFileSync('package.json', JSON.stringify(pkg, null, 2));"
	pnpm pack --pack-destination release
	mv .package.json.bak package.json
	mkdir -p release/latest
	tar -xf release/mtran-core-*.tgz -C release/latest --strip-components=1