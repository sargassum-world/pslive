module.exports = {
  parser: '@typescript-eslint/parser',
  extends: [
    'eslint:recommended',
  ],
  parserOptions: {
    ecmaVersion: 2020,
    sourceType: 'module',
    project: './tsconfig.json',
    extraFileExtensions: ['.svelte']
  },
  env: {
    browser: true,
    es6: true
  },
  overrides: [
    {
      files: ['**/*.svelte'],
      processor: 'svelte3/svelte3',
      extends: [
        'eslint:recommended',
        'plugin:@typescript-eslint/recommended',
        'plugin:@typescript-eslint/recommended-requiring-type-checking',
      ],
      plugins: ['svelte3', '@typescript-eslint'],
      settings: {
        'svelte3/typescript': require('typescript')
        // ignore style tags in Svelte because of Tailwind CSS
        // See https://github.com/sveltejs/eslint-plugin-svelte3/issues/70
        //'svelte3/ignore-styles': () => true
      },
    },
    {
      files: ['**/*.ts'],
      extends: [
        'eslint:recommended',
        'plugin:@typescript-eslint/recommended',
        'plugin:@typescript-eslint/recommended-requiring-type-checking',
      ],
      plugins: ['@typescript-eslint'],
    }
  ],
  rules: {
    'linebreak-style': 'off',
  },
  ignorePatterns: ['node_modules']
};
