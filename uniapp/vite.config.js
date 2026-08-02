import { defineConfig } from 'vite'
import uni from '@dcloudio/vite-plugin-uni'

// Standard uni-app Vue 3 build config. The `uni()` plugin internally
// configures @vitejs/plugin-vue + the uni-app runtime, so we don't add
// it separately. Adding it twice (as in an earlier attempt) caused
// "At least one <template> or <script> is required" because two Vue
// parsers raced over the SFC AST.
export default defineConfig({
  plugins: [uni()],
  build: {
    // Output is consumed by the admin frontend's /showcase-app/ path.
    outDir: 'dist/build/h5',
    emptyOutDir: true,
  },
})