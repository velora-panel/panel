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

test('account administration, password changes, other-session revocation and disabled access', async ({
  page,
  browser,
}) => {
  const adminPassword = String(process.env.E2E_PASSWORD)
  const email = `user-${randomUUID()}@example.com`
  const password = `Initial-${randomUUID()}`
  const changed = `Changed-${randomUUID()}`
  const signIn = async (
    target: typeof page,
    address: string,
    secret: string,
  ) => {
    await target.goto('/login')
    await target.getByLabel('Email address').fill(address)
    await target.getByLabel('Password', { exact: true }).fill(secret)
    await target.getByRole('button', { name: 'Sign in', exact: true }).click()
    await expect(
      target.getByRole('heading', { name: 'Overview', exact: true }),
    ).toBeVisible()
  }
  await signIn(page, String(process.env.E2E_EMAIL), adminPassword)
  await page.getByRole('link', { name: 'Users', exact: true }).click()
  await page.getByLabel('Email', { exact: true }).fill(email)
  await page.getByLabel('Initial password').fill(password)
  await page.getByLabel('Your current password').fill(adminPassword)
  await page.getByRole('button', { name: 'Create user', exact: true }).click()
  await expect(
    page.getByRole('button', { name: `Manage ${email}`, exact: true }),
  ).toBeVisible()
  const userContext = await browser.newContext()
  const otherContext = await browser.newContext()
  const userPage = await userContext.newPage()
  const otherPage = await otherContext.newPage()
  try {
    await signIn(userPage, email, password)
    await signIn(otherPage, email, password)
    expect((await userPage.request.get('/api/v1/users')).status()).toBe(403)
    expect(
      (
        await userPage.request.post('/api/v1/nodes', {
          data: { name: 'forbidden' },
        })
      ).status(),
    ).toBe(403)
    await userPage.getByRole('link', { name: 'Settings', exact: true }).click()
    await userPage.locator('#password-current').fill(password)
    await userPage.getByLabel('New password', { exact: true }).fill(changed)
    await userPage.getByLabel('Confirm new password').fill(changed)
    await userPage
      .getByRole('button', { name: 'Change password', exact: true })
      .click()
    await expect(userPage.getByRole('status')).toHaveText(
      'Account updated successfully.',
    )
    expect((await otherPage.request.get('/api/v1/auth/me')).status()).toBe(401)
    await otherPage.reload()
    await expect(otherPage).toHaveURL('/login')
    expect((await userPage.request.get('/api/v1/auth/me')).status()).toBe(200)
    await page
      .getByRole('button', { name: `Manage ${email}`, exact: true })
      .click()
    await page.getByRole('switch', { name: 'Account disabled' }).click()
    await page
      .getByLabel('Your current password', { exact: true })
      .last()
      .fill(adminPassword)
    await page
      .getByRole('button', { name: 'Save changes', exact: true })
      .click()
    await expect(page.getByRole('dialog')).not.toBeVisible()
    await userPage.reload()
    await expect(userPage).toHaveURL('/login')
    await page
      .getByRole('button', { name: `Manage ${email}`, exact: true })
      .click()
    await page.getByRole('switch', { name: 'Account disabled' }).click()
    await page
      .getByLabel('Your current password', { exact: true })
      .last()
      .fill(adminPassword)
    await page
      .getByRole('button', { name: 'Save changes', exact: true })
      .click()
    await expect(page.getByRole('dialog')).not.toBeVisible()
    await signIn(userPage, email, changed)
    await page
      .getByRole('button', { name: `Manage ${email}`, exact: true })
      .click()
    await page
      .getByLabel('Your current password', { exact: true })
      .last()
      .fill(adminPassword)
    await page
      .getByRole('button', { name: 'Revoke sessions', exact: true })
      .click()
    await expect(page.getByRole('dialog')).not.toBeVisible()
    await userPage.reload()
    await expect(userPage).toHaveURL('/login')
  } finally {
    await userContext.close()
    await otherContext.close()
  }
})
