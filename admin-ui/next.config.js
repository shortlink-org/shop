/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'standalone',
  basePath: '/admin',
  transpilePackages: ['antd', '@ant-design/icons'],
};

module.exports = nextConfig;
