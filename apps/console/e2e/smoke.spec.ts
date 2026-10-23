import { test, expect } from '@playwright/test'
import { randomUUID } from 'node:crypto'
test('real authentication, navigation, Node creation and session revocation', async ({
  page,
}, testInfo) => {
  await page.goto('/nodes')
  await expect(page).toHaveURL('/login')
  await expect(
    page.getByRole('heading', { name: 'Welcome to Velora' }),
  ).toBeVisible()
  await page.getByLabel('Email address').fill(String(process.env.E2E_EMAIL))
  await page
    .getByLabel('Password', { exact: true })
    .fill(String(process.env.E2E_PASSWORD))
  await page.getByRole('button', { name: 'Sign in', exact: true }).click()
  await expect(
    page.getByRole('heading', { name: 'Overview', exact: true }),
  ).toBeVisible()
  await page.screenshot({
    path: testInfo.outputPath('overview.png'),
    fullPage: true,
  })
  await page
    .getByRole('navigation')
    .getByRole('link', { name: 'Nodes', exact: true })
    .click()
  await page.getByRole('button', { name: 'Add Node' }).click()
  const name = `Smoke ${randomUUID()}`
  await page.getByLabel('Node name').fill(name)
  await page
    .getByLabel('Description', { exact: true })
    .fill('Isolated browser test')
  await page.getByRole('button', { name: 'Create Node' }).click()
  await page.getByRole('link').filter({ hasText: name }).click()
  await expect(page.getByRole('heading', { name })).toBeVisible()
  await expect(page.getByText('unconfigured', { exact: true })).toBeVisible()
  await page.setViewportSize({ width: 390, height: 844 })
  await page.screenshot({
    path: testInfo.outputPath('node-mobile.png'),
    fullPage: true,
  })
  expect(
    await page.evaluate(
      () => document.documentElement.scrollWidth <= window.innerWidth,
    ),
  ).toBe(true)
  await page.setViewportSize({ width: 1280, height: 720 })
  await page.reload()
  await expect(page.getByRole('heading', { name })).toBeVisible()
  await page
    .getByRole('navigation')
    .getByRole('link', { name: 'Instances', exact: true })
    .click()
  await expect(
    page.getByRole('heading', { name: 'Nothing here yet' }),
  ).toBeVisible()
  await page.getByRole('button', { name: 'Account' }).click()
  await page.getByRole('menuitem', { name: 'Sign out' }).click()
  await expect(page).toHaveURL('/login')
  expect((await page.request.get('/api/v1/auth/me')).status()).toBe(401)
  await page.goto('/nodes')
  await expect(page).toHaveURL('/login')
})
