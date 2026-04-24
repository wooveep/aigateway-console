import { expect, test, type Locator, type Page } from '@playwright/test';

const adminUsername = process.env.CONSOLE_E2E_ADMIN_USERNAME || 'admin';
const adminPassword = process.env.CONSOLE_E2E_ADMIN_PASSWORD || 'admin';

async function clickPrimarySubmit(page: Page) {
  await page.locator('button.ant-btn-primary').last().click();
}

async function ensureSignedIn(page: Page) {
  await page.goto('/');

  if (page.url().includes('/init')) {
    await page.getByPlaceholder(/用户名|Username/).fill(adminUsername);
    await page.getByPlaceholder(/密码|Password/).fill(adminPassword);
    await page.getByPlaceholder(/确认密码|Confirm Password/).fill(adminPassword);
    await clickPrimarySubmit(page);
    await page.waitForURL(/\/login(?:\?|$)/);
  }

  if (page.url().includes('/login')) {
    await page.getByPlaceholder(/用户名|Username/).fill(adminUsername);
    await page.getByPlaceholder(/密码|Password/).fill(adminPassword);
    await clickPrimarySubmit(page);
  }

  await page.waitForURL(/\/dashboard(?:\?|$)|\/route(?:\?|$)|\/ai\/route(?:\?|$)/);
}

async function expectDrawerVisible(drawer: Locator) {
  await expect(drawer).toBeVisible();
  await expect(drawer.locator('.ant-drawer-body')).toBeVisible();
}

async function closeDrawer(drawer: Locator) {
  await drawer.getByRole('button', { name: /取\s*消|Cancel/ }).click();
  await expect(drawer).toBeHidden();
}

test.describe('console management page smoke', () => {
  test.beforeEach(async ({ page }) => {
    await ensureSignedIn(page);
  });

  test('loads dashboard and consumer management', async ({ page }) => {
    await page.goto('/dashboard');
    await expect(page.getByText('监控视图', { exact: true })).toBeVisible();

    await page.goto('/consumer');
    await expect(page.getByRole('button', { name: /创建|新增/ })).toBeVisible();
    await page.getByRole('button', { name: /创建|新增/ }).first().click();
    const modal = page.locator('.ant-modal-content').last();
    await expect(modal).toBeVisible();
    await modal.getByRole('button', { name: /取\s*消|Cancel/ }).click();
    await expect(modal).toBeHidden();
  });

  test('opens model asset, provider and agent catalog drawers', async ({ page }) => {
    await page.goto('/ai/model-assets');
    await expect(page.getByText('模型资产管理', { exact: true })).toBeVisible();
    await page.getByRole('button', { name: '创建模型资产' }).click();
    let drawer = page.locator('.ant-drawer-content').last();
    await expectDrawerVisible(drawer);
    await expect(drawer.getByText('新建模型资产', { exact: true })).toBeVisible();
    await closeDrawer(drawer);

    await page.goto('/ai/provider');
    await expect(page.getByText('AI 服务提供者管理', { exact: true })).toBeVisible();
    await page.getByRole('button', { name: '新增 Provider' }).click();
    drawer = page.locator('.ant-drawer-content').last();
    await expectDrawerVisible(drawer);
    await closeDrawer(drawer);

    await page.goto('/ai/agent-catalog');
    await expect(page.getByText('智能体目录管理', { exact: true })).toBeVisible();
    await page.getByRole('button', { name: '创建智能体目录' }).click();
    drawer = page.locator('.ant-drawer-content').last();
    await expectDrawerVisible(drawer);
    await expect(drawer.getByText('创建智能体目录', { exact: true })).toBeVisible();
    await closeDrawer(drawer);
  });

  test('loads ai quota, mcp list and system jobs pages', async ({ page }) => {
    await page.goto('/ai/quota');
    await expect(page.getByText(/AI 配额管理|AI Quota/)).toBeVisible();

    await page.goto('/mcp/list');
    await expect(page.getByText('MCP 配置', { exact: true })).toBeVisible();
    await page.getByRole('button', { name: '创建 MCP 服务' }).click();
    let drawer = page.locator('.ant-drawer-content').last();
    await expectDrawerVisible(drawer);
    await expect(drawer.getByText('创建 MCP 服务', { exact: true })).toBeVisible();
    await closeDrawer(drawer);

    await page.goto('/system/jobs');
    await expect(page.getByText('Jobs 运维', { exact: true })).toBeVisible();
    await expect(page.getByRole('button', { name: '刷新' })).toBeVisible();
  });
});
