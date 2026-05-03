/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'standalone',
  async rewrites() {
    return [
      {
        source: '/api/:path*',
        destination: 'http://localhost:8080/api/:path*',
      },
      {
        source: '/webhook/:path*',
        destination: 'http://localhost:8080/webhook/:path*',
      },
    ];
  },
};

module.exports = nextConfig;
