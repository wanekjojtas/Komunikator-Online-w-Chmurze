import type { NextConfig } from "next";

const port = process.env.PORT || 3000;

const nextConfig: NextConfig = {
  /* config options here */
  reactStrictMode: true,
  serverRuntimeConfig: {
    port,
  },
};

export default nextConfig;
