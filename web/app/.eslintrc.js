module.exports = {
  extends: [
    'eslint:recommended',
  ],
  parserOptions: {
    ecmaVersion: 2022,
    sourceType: 'module',
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
      ],
      plugins: ['svelte3'],
      settings: {
        // ignore style tags in Svelte because of Tailwind CSS
        // See https://github.com/sveltejs/eslint-plugin-svelte3/issues/70
        //'svelte3/ignore-styles': () => true
      },
    }
  ],
  rules: {
    'linebreak-style': 'off',
  },
  ignorePatterns: ['node_modules']
};
