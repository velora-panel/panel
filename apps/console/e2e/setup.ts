import { execFileSync } from 'node:child_process'
import { randomBytes } from 'node:crypto'
export default function setup() {
  const database = process.env.TEST_DATABASE_URL
  if (!database || !new URL(database).pathname.endsWith('_test'))
    throw new Error('Dedicated test database required')
  process.env.E2E_EMAIL = `browser-${randomBytes(8).toString('hex')}@example.com`
  process.env.E2E_PASSWORD = randomBytes(24).toString('base64url')
  const options = {
    cwd: '../core',
    env: {
      ...process.env,
      DATABASE_URL: database,
      APP_ENV: 'development',
      PUBLIC_URL: 'http://localhost:15173',
      COOKIE_SECURE: 'false',
    },
  }
  execFileSync('go', ['run', './cmd/velora', 'migrate'], options)
  execFileSync('go', ['run', './cmd/velora', 'admin'], {
    ...options,
    input: `${process.env.E2E_EMAIL}\n${process.env.E2E_PASSWORD}\n`,
  })
}
