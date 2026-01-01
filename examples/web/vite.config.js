import { defineConfig } from 'vite';

export default defineConfig({
  server: {
    proxy: {
      '/mozilla-api': {
        target: 'https://firefox.settings.services.mozilla.com',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/mozilla-api/, '')
      },
      '/mozilla-cdn': {
        target: 'https://firefox-settings-attachments.cdn.mozilla.net',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/mozilla-cdn/, '')
      }
    }
  }
});
