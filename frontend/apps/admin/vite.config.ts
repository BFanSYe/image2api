import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import path from 'node:path';

export default defineConfig({
  base: '/admin/',
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, 'src'),
      '@image2api/theme': path.resolve(__dirname, '../../packages/theme/src'),
    },
  },
  server: {
    port: 5174,
    host: true,
    proxy: {
      '/admin/api': {
        target: 'http://127.0.0.1:17188',
        changeOrigin: true,
      },
    },
  },
});
