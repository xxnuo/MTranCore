import { MTranCore } from '../../../dist/index.web.js'; // Import directly from dist

// Define UI elements
const input = document.getElementById('input');
const output = document.getElementById('output');
const fromLangSelect = document.getElementById('fromLang');
const toLangSelect = document.getElementById('toLang');
const translateBtn = document.getElementById('translateBtn');
const statusDiv = document.getElementById('status');

// Initialize MTranCore with Proxy URLs
const mtran = new MTranCore({
  modelPath: 'mtran-models-cache', // Cache Storage name
  wasmPath: '/bergamot-translator.wasm',
  cld2Path: '/cld2.wasm',
  // Use local proxy paths configured in vite.config.js
  recordsUrl: '/mozilla-api/v1/buckets/main-preview/collections/translations-models-v2/records',
  attachmentsBaseUrl: '/mozilla-cdn'
});

function setStatus(msg, type = 'normal') {
  statusDiv.textContent = msg;
  statusDiv.className = `status ${type}`;
}

async function doTranslate() {
  const text = input.value.trim();
  if (!text) return;

  const from = fromLangSelect.value;
  const to = toLangSelect.value;

  translateBtn.disabled = true;
  setStatus('Initializing engine and checking models... (First run may take time to download)', 'loading');
  output.textContent = '';

  try {
    setStatus('Downloading/Loading models...', 'loading');
    
    const start = performance.now();
    // Auto-download happens here internally
    const result = await mtran.translate(text, from, to);
    const end = performance.now();

    output.textContent = result;
    setStatus(`Done in ${Math.round(end - start)}ms`, 'normal');
  } catch (error) {
    console.error(error);
    setStatus(`Error: ${error.message}`, 'error');
  } finally {
    translateBtn.disabled = false;
  }
}

translateBtn.addEventListener('click', doTranslate);

// Initial status
setStatus('MTran Core loaded. Click Translate to start.');