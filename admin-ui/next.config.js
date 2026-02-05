/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'standalone',
  basePath: '/admin-ui',
  transpilePackages: ['antd', '@ant-design/icons'],
};

module.exports = nextConfig;
