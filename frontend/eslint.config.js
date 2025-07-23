// import js from '@eslint/js'
// import globals from 'globals'
// import reactHooks from 'eslint-plugin-react-hooks'
// import reactRefresh from 'eslint-plugin-react-refresh'

// export default [
//   { ignores: ['dist', '**/*.test.js'] },

//   {
//     files: ['tailwind.config.js', 'postcss.config.js'],
//     languageOptions: {
//       ecmaVersion: 2020,
//       sourceType: 'module',
//       globals: globals.node,
//     },
//   },
  
//   {
//     files: ['**/*.{js,jsx}'],
//     languageOptions: {
//       ecmaVersion: 2020,
//       globals: Object.fromEntries(
//   Object.entries(globals.browser).map(([key, value]) => [key.trim(), value])
// ),
//       parserOptions: {
//         ecmaVersion: 'latest',
//         ecmaFeatures: { jsx: true },
//         sourceType: 'module',
//       },
//     },
//     plugins: {
//       'react-hooks': reactHooks,
//       'react-refresh': reactRefresh,
//     },
//     rules: {
//       ...js.configs.recommended.rules,
//       ...reactHooks.configs.recommended.rules,
//       'no-unused-vars': ['error', { varsIgnorePattern: '^[A-Z_]' }],
//       'react-refresh/only-export-components': [
//         'warn',
//         { allowConstantExport: true },
//       ],
//     },
//   },
// ]

import js from '@eslint/js';
import globals from 'globals';
import reactHooks from 'eslint-plugin-react-hooks';
import reactRefresh from 'eslint-plugin-react-refresh';
import tseslint from 'typescript-eslint';
import process from 'node:process';


// Helper to fix the whitespace bug in global keys
const cleanGlobals = (source) =>
  Object.fromEntries(Object.entries(source).map(([k, v]) => [k.trim(), v]));

export default [
  {
    ignores: ['dist', '**/*.test.*', 'node_modules'],
  },

  // Tailwind/PostCSS files
  {
    files: ['tailwind.config.js', 'postcss.config.js'],
    languageOptions: {
      ecmaVersion: 2020,
      sourceType: 'module',
      globals: cleanGlobals(globals.node),
    },
  },

  // JS + JSX
  {
    files: ['**/*.{js,jsx}'],
    languageOptions: {
      ecmaVersion: 2020,
      parserOptions: {
        ecmaFeatures: { jsx: true },
        sourceType: 'module',
      },
      globals: cleanGlobals(globals.browser),
    },
    plugins: {
      'react-hooks': reactHooks,
      'react-refresh': reactRefresh,
    },
    rules: {
      ...js.configs.recommended.rules,
      ...reactHooks.configs.recommended.rules,
      'no-unused-vars': ['error', { varsIgnorePattern: '^[A-Z_]' }],
      'react-refresh/only-export-components': ['warn', { allowConstantExport: true }],
    },
  },

  //  TypeScript + TSX
  {
    files: ['**/*.{ts,tsx}'],
    languageOptions: {
      parser: tseslint.parser,
      parserOptions: {
        project: ['./tsconfig.node.json'], // make sure this file exists
        tsconfigRootDir: process.cwd(),
        sourceType: 'module',
      },
      globals: cleanGlobals(globals.browser),
    },
    plugins: {
      '@typescript-eslint': tseslint.plugin,
      'react-hooks': reactHooks,
      'react-refresh': reactRefresh,
    },
    rules: {
      ...tseslint.configs.recommendedTypeChecked[0].rules, // only spread rules object
      ...reactHooks.configs.recommended.rules,
      "no-unused-vars": ["error", { "argsIgnorePattern": "^_" }],
      'react-refresh/only-export-components': ['warn', { allowConstantExport: true }],
    },
  },
];
