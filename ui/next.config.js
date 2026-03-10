// ENVIRONMENT VARIABLE ================================================================================================
const isProd = process.env.NODE_ENV === 'production';

const API_URI = process.env.API_URI || 'http://127.0.0.1:7070';

console.info('API_URI', API_URI);

/** @type {import('next').NextConfig} */
module.exports = {
  output: 'standalone',
  cacheHandler: require.resolve('./cache-handler.js'),
  cacheMaxMemorySize: 0,
  reactStrictMode: true,
  reactCompiler: true,
  env: {
    // ShortLink API
    NEXT_PUBLIC_SERVICE_NAME: 'shortlink-shop-ui',
    NEXT_PUBLIC_GIT_TAG: process.env.CI_COMMIT_TAG
  },
  generateEtags: isProd,
  images: {
    remotePatterns: [
      {
        protocol: 'https',
        hostname: 'picsum.photos'
      }
    ]
  },
  trailingSlash: false,
  experimental: {
    turbopackFileSystemCacheForBuild: true,
    webVitalsAttribution: ['CLS', 'FCP', 'FID', 'INP', 'LCP', 'TTFB'],
    viewTransition: true,
    webpackMemoryOptimizations: true
  }
};
