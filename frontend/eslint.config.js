import js from '@eslint/js'
import { defineConfig, globalIgnores } from 'eslint/config'
import svelte from 'eslint-plugin-svelte'
import globals from 'globals'
import tseslint from 'typescript-eslint'

const projectRules = {
  'func-style': ['error', 'expression', { allowArrowFunctions: true }],
  'prefer-arrow-callback': ['error', { allowNamedFunctions: false }],
  'object-shorthand': 'error',
  'no-empty': ['error', { allowEmptyCatch: true }],
  'no-console': ['warn', { allow: ['warn', 'error'] }],
}

const typescriptRules = {
  '@typescript-eslint/no-explicit-any': 'error',
  '@typescript-eslint/no-non-null-assertion': 'error',
  '@typescript-eslint/no-unused-vars': [
    'error',
    {
      argsIgnorePattern: '^_',
      varsIgnorePattern: '^_',
      caughtErrorsIgnorePattern: '^_',
    },
  ],
  '@typescript-eslint/consistent-type-imports': [
    'error',
    { prefer: 'type-imports', fixStyle: 'inline-type-imports' },
  ],
  '@typescript-eslint/consistent-type-exports': [
    'error',
    { fixMixedExportsWithInlineTypeSpecifier: true },
  ],
}

export default defineConfig([
  globalIgnores(['dist/**', 'coverage/**', 'wailsjs/**']),
  {
    linterOptions: {
      reportUnusedDisableDirectives: 'error',
    },
  },
  {
    files: ['**/*.ts'],
    extends: [js.configs.recommended, tseslint.configs.recommended],
    languageOptions: {
      ecmaVersion: 'latest',
      globals: globals.browser,
      parserOptions: {
        projectService: true,
        tsconfigRootDir: import.meta.dirname,
      },
    },
    rules: {
      ...projectRules,
      ...typescriptRules,
    },
  },
  ...svelte.configs.recommended,
  {
    files: ['**/*.svelte'],
    plugins: {
      '@typescript-eslint': tseslint.plugin,
    },
    languageOptions: {
      globals: globals.browser,
      parserOptions: {
        parser: tseslint.parser,
        extraFileExtensions: ['.svelte'],
        projectService: true,
        tsconfigRootDir: import.meta.dirname,
      },
    },
    rules: {
      ...projectRules,
      ...typescriptRules,
      'no-undef': 'off',
      'no-unused-vars': 'off',
      'svelte/button-has-type': 'error',
      'svelte/require-event-dispatcher-types': 'error',
    },
  },
  ...svelte.configs.prettier,
  {
    files: ['src/lib/logger.ts'],
    rules: {
      'no-console': 'off',
    },
  },
])
