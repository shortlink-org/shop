import nextCoreWebVitals from 'eslint-config-next/core-web-vitals';
import nextTypescript from 'eslint-config-next/typescript';

const config = [
  ...nextCoreWebVitals,
  ...nextTypescript,
  {
    files: ['cache-handler.js', 'tailwind.config.js'],
    rules: { '@typescript-eslint/no-require-imports': 'off' }
  }
];

export default config;
