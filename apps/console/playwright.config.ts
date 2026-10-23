import { defineConfig } from '@playwright/test'
const database = process.env.TEST_DATABASE_URL
if (!database || !new URL(database).pathname.endsWith('_test'))
  throw new Error(
    'Set TEST_DATABASE_URL to an isolated database ending in _test.',
  )
export default defineConfig({
  testDir: './e2e',
  workers: 1,
  retries: 0,
  globalSetup: './e2e/setup.ts',
  use: { baseURL: 'http://localhost:15173', trace: 'retain-on-failure' },
  webServer: [
    {
      command: 'pnpm velora serve',
      cwd: '../..',
      url: 'http://127.0.0.1:18080/api/v1/health/live',
      reuseExistingServer: false,
      env: {
        DATABASE_URL: database,
        APP_ENV: 'development',
        PUBLIC_URL: 'http://localhost:15173',
        HTTP_PORT: '18080',
        COOKIE_SECURE: 'false',
      },
    },
    {
      command: 'pnpm exec vite --host localhost --port 15173 --strictPort',
      url: 'http://localhost:15173',
      reuseExistingServer: false,
      env: { CORE_PROXY: 'http://127.0.0.1:18080' },
    },
  ],
})
