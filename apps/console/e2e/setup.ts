import { execFileSync } from 'node:child_process'
import { randomBytes } from 'node:crypto'
export default function setup() {
  const database = process.env.TEST_DATABASE_URL
  if (!database || !new URL(database).pathname.endsWith('_test'))
    throw new Error('Dedicated test database required')
  process.env.E2E_EMAIL = `browser-${randomBytes(8).toString('hex')}@example.com`
  process.env.E2E_PASSWORD = randomBytes(24).toString('base64url')
  const options = {
    cwd: '../..',
    env: {
      ...process.env,
      DATABASE_URL: database,
      MIGRATION_DATABASE_URL: database,
      APP_ENV: 'development',
      PUBLIC_URL: 'http://localhost:15173',
      COOKIE_SECURE: 'false',
    },
  }
  execFileSync('pnpm', ['velora', 'migrate', 'up'], options)
  execFileSync('pnpm', ['velora', 'admin', 'create', '--password-stdin'], {
    ...options,
    input: JSON.stringify({
      email: process.env.E2E_EMAIL,
      password: process.env.E2E_PASSWORD,
      password_confirmation: process.env.E2E_PASSWORD,
    }),
  })
}
