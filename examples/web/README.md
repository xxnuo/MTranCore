# MTran Web Example

A Vite-based example demonstrating MTran Core in the browser.

## Setup & Run

1.  **Prepare dependencies**:
    Ensure the parent `mtran-core` is built (`npm run build` in `../../`).

2.  **Install example dependencies**:
    ```bash
    npm install
    ```

3.  **Start Dev Server**:
    ```bash
    npm run dev
    ```

4.  Open the browser URL provided (usually `http://localhost:5173`).

## How it works

- **WASM Assets**: The `bergamot-translator.wasm` and `cld2.wasm` files are placed in `public/` so they are served at the root URL.
- **Model Storage**: The models are downloaded from Mozilla's CDN directly by the browser and stored in **Cache Storage** (inspect Application -> Cache Storage in DevTools).
