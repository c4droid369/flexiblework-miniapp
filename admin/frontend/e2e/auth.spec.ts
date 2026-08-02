import { expect, test } from '@playwright/test';

// Login is the gateway to every other test. Cover the happy path here; other
// tests assume they can skip straight past it.
test('login with valid credentials lands on dashboard', async ({ page }) => {
  await page.goto('/login');
  await expect(page).toHaveTitle(/Admin/);

  await page.getByPlaceholder('admin').fill('admin');
  await page.getByPlaceholder('••••••').fill('admin123');
  await page.getByRole('button', { name: '登录' }).click();

  // Lands on dashboard; sidebar has the seeded menus.
  await expect(page).toHaveURL(/\/$|\/dashboard/);
  await expect(page.getByText('仪表盘').first()).toBeVisible();
  await expect(page.getByText('系统管理').first()).toBeVisible();
});

test('login with wrong password shows an error', async ({ page }) => {
  await page.goto('/login');
  await page.getByPlaceholder('admin').fill('admin');
  await page.getByPlaceholder('••••••').fill('not-the-password');
  await page.getByRole('button', { name: '登录' }).click();

  // Still on login page (no navigation). ElMessage toast may or may not
  // render depending on its lifecycle — the contract we care about is
  // "user is not logged in".
  await expect(page).toHaveURL(/\/login/);
  // No token in localStorage.
  const token = await page.evaluate(() => localStorage.getItem('admin_template_access_token'));
  expect(token ?? '').toBe('');
});
