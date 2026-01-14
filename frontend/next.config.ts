import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  env: {
    NEXT_PUBLIC_API_URL: process.env.NEXT_PUBLIC_API_URL,
    NEXT_PUBLIC_AWS_REGION: process.env.NEXT_PUBLIC_AWS_REGION,
    NEXT_PUBLIC_USER_POOL_ID: process.env.NEXT_PUBLIC_USER_POOL_ID,
    NEXT_PUBLIC_USER_POOL_WEB_CLIENT_ID: process.env.NEXT_PUBLIC_USER_POOL_WEB_CLIENT_ID,
    NEXT_PUBLIC_OAUTH_DOMAIN: process.env.NEXT_PUBLIC_OAUTH_DOMAIN,
  },
};

export default nextConfig;
