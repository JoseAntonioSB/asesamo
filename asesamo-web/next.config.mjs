/** @type {import('next').NextConfig} */
const nextConfig = {
  output: "standalone",
  reactStrictMode: true,
  poweredByHeader: false,
  async rewrites() {
    const internalApi = process.env.INTERNAL_API_URL || "http://127.0.0.1:8080/api/v1";
    return [
      {
        source: "/api/proxy/:path*",
        destination: `${internalApi}/:path*`,
      },
    ];
  },
};

export default nextConfig;
