import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  build: { outDir: '../web/dist', emptyOutDir: true, assetsDir: 'assets', assetsInlineLimit: 0 },
  server: {
    proxy: Object.fromEntries(['/api', '/health'].map((path) => [path, 'http://127.0.0.1:8096'])),
  },
});
