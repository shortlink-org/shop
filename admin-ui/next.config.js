/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'standalone',
  basePath: '/admin-ui',
  reactCompiler: true,
  transpilePackages: ['antd', '@ant-design/icons'],
  experimental: {
    turbopackFileSystemCacheForBuild: true
  }
};

module.exports = nextConfig;
