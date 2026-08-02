// ESLint v9 flat config for Vue3 + TypeScript. Prettier handles formatting;
// eslint-config-prettier disables any lint rule that would fight Prettier.
import js from '@eslint/js';
import tsParser from '@typescript-eslint/parser';
import tsPlugin from '@typescript-eslint/eslint-plugin';
import vueParser from 'vue-eslint-parser';
import vuePlugin from 'eslint-plugin-vue';
import prettierConfig from 'eslint-config-prettier';

export default [
  // Always start with the recommended baseline.
  js.configs.recommended,
  ...vuePlugin.configs['flat/recommended'],

  // Global ignores — keep this minimal so we don't accidentally hide bugs.
  {
    ignores: [
      '**/dist/**',
      '**/node_modules/**',
      '**/test-results/**',
      'src/auto-imports.d.ts',
      'src/components.d.ts',
    ],
  },

  // TypeScript & Vue source — strict but not pedantic.
  {
    files: ['src/**/*.{ts,vue}', 'e2e/**/*.ts'],
    languageOptions: {
      parser: vueParser,
      parserOptions: {
        parser: tsParser,
        ecmaVersion: 'latest',
        sourceType: 'module',
        extraFileExtensions: ['.vue'],
      },
    },
    plugins: {
      '@typescript-eslint': tsPlugin,
    },
    rules: {
      // --- correctness we DO want ---
      'no-console': ['warn', { allow: ['warn', 'error'] }],
      'no-debugger': 'error',
      'no-undef': 'off', // TypeScript handles this
      'no-unused-vars': 'off', // covered by TS noUnusedLocals
      'no-unused-imports': 'off', // TS handles imports
      'no-empty': ['error', { allowEmptyCatch: true }],
      'no-prototype-builtins': 'off', // common with DTO types

      // --- typescript-eslint ---
      '@typescript-eslint/no-explicit-any': 'warn',
      '@typescript-eslint/no-unused-vars': [
        'error',
        { argsIgnorePattern: '^_', varsIgnorePattern: '^_' },
      ],
      '@typescript-eslint/no-non-null-assertion': 'off', // common in test assertions and DTO narrowing
      '@typescript-eslint/consistent-type-imports': ['warn', { prefer: 'type-imports' }],

      // --- Vue ---
      'vue/multi-word-component-names': 'off', // we have index/Pagination; rename churn not worth it
      'vue/no-v-html': 'error',
      'vue/component-api-style': 'off', // composition API only — let any style pass
      'vue/attribute-hyphenation': 'off', // element-plus accepts both
      'vue/v-on-event-hyphenation': 'off',
      'vue/html-self-closing': 'off', // element-plus prefers explicit children
      'vue/max-attributes-per-line': 'off',
      'vue/singleline-html-element-content-newline': 'off',
      'vue/html-indent': 'off',
    },
  },

  // Plain JS / config files — keep it minimal.
  {
    files: ['*.{js,cjs,mjs,json}', 'vite.config.ts', 'vitest.config.ts', 'playwright.config.ts'],
    languageOptions: {
      parser: tsParser,
      parserOptions: { ecmaVersion: 'latest', sourceType: 'module' },
    },
    rules: {
      'no-undef': 'off',
    },
  },

  // Test files — relax rules that fight with test patterns.
  {
    files: ['**/*.test.ts', '**/*.spec.ts', 'e2e/**/*.ts'],
    rules: {
      '@typescript-eslint/no-explicit-any': 'off',
      'no-console': 'off',
    },
  },

  // Prettier must be LAST so its formatter overrides win.
  prettierConfig,
];
