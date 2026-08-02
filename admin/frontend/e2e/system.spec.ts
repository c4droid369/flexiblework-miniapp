import { expect, test, type Page } from '@playwright/test';

// Logs in once per test via storage state — login flow itself is covered
// in auth.spec.ts.
test.use({ storageState: undefined }); // each test logs in fresh

async function login(page: Page) {
  await page.goto('/login');
  await page.getByPlaceholder('admin').fill('admin');
  await page.getByPlaceholder('••••••').fill('admin123');
  await page.getByRole('button', { name: '登录' }).click();
  await page.waitForURL(/\/$|\/dashboard/);
}

test('sidebar shows seeded menu groups', async ({ page }) => {
  await login(page);
  // Top-level items always render. Sub-menus are collapsed by default and
  // require an extra click to reveal — covered by the navigation tests.
  await expect(page.getByText('仪表盘').first()).toBeVisible();
  await expect(page.getByText('个人中心').first()).toBeVisible();
  await expect(page.getByText('系统管理').first()).toBeVisible();
  // Expand to verify the children mount.
  await page.getByText('系统管理').first().click();
  await expect(page.getByText('用户管理').first()).toBeVisible();
  await expect(page.getByText('角色管理').first()).toBeVisible();
  await expect(page.getByText('操作日志').first()).toBeVisible();
});

test('navigating to user management shows the table', async ({ page }) => {
  await login(page);
  // Navigate by URL — sidebar sub-menus are collapsed by default and
  // clicking the parent is brittle for e2e.
  await page.goto('/system/user');

  // Wait for the table header — proves the page mounted.
  await expect(page.getByRole('columnheader', { name: '用户名' })).toBeVisible();
  await expect(page.getByRole('columnheader', { name: '邮箱' })).toBeVisible();
  // "admin" row should be present.
  await expect(page.getByRole('cell', { name: 'admin' }).first()).toBeVisible();
});

test('create then delete a user round-trips', async ({ page }) => {
  await login(page);
  await page.goto('/system/user');
  await expect(page.getByRole('columnheader', { name: '用户名' })).toBeVisible();

  const stamp = Date.now();
  const uname = `e2e_${stamp}`;

  // Open create dialog.
  await page.getByRole('button', { name: '新增' }).click();
  await expect(page.getByRole('dialog')).toBeVisible();

  // Fill form. Element Plus inputs render as <input> inside wrappers.
  await page.getByRole('dialog').locator('input').nth(0).fill(uname);
  await page.getByRole('dialog').locator('input[type="password"]').fill('test-pass-1234');
  await page.getByRole('dialog').locator('input').nth(2).fill('E2E user');

  await page.getByRole('button', { name: '保存' }).click();

  // New row appears.
  await expect(page.getByRole('cell', { name: uname }).first()).toBeVisible({ timeout: 10_000 });

  // Delete via row action — confirm dialog OK.
  await page.getByRole('row', { name: uname }).getByRole('button', { name: '删除' }).click();
  await page.getByRole('button', { name: '确定' }).click();
  await expect(page.getByRole('cell', { name: uname })).toHaveCount(0, { timeout: 10_000 });
});

test('logout returns to login', async ({ page }) => {
  await login(page);
  await expect(page.getByText('admin').first()).toBeVisible();

  // Open the avatar dropdown and pick logout.
  await page.locator('.user').click();
  await page.getByRole('menuitem', { name: '退出登录' }).click();

  await expect(page).toHaveURL(/\/login/);
  // After logout localStorage is cleared; visiting /dashboard kicks back.
  await page.goto('/dashboard');
  await expect(page).toHaveURL(/\/login/);
});
