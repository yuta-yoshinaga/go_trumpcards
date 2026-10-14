import { expect, test } from '@playwright/test';
import { navigateTo } from './helpers';

test.use({ viewport: { width: 1280, height: 800 } });

async function maxImageBottom(locator: ReturnType<import('@playwright/test').Page['locator']>) {
  const count = await locator.count();
  expect(count, 'tableau should contain at least one card image').toBeGreaterThan(0);
  return locator.evaluateAll((images) => Math.max(...images.map((image) => image.getBoundingClientRect().bottom)));
}

test('Big Ben tableau cards stay above the GameMessageBox', async ({ page }) => {
  await navigateTo(page, '/bigben');
  const cards = page.locator('[data-tutorial="gc-tableau"] img');
  const messageBox = page.locator('main [role="status"].glass-panel');

  await expect(messageBox).toBeVisible();
  const bottom = await maxImageBottom(cards);
  const messageTop = await messageBox.evaluate((element) => element.getBoundingClientRect().top);
  expect(bottom, 'Big Ben tableau card bottoms should not pass the GameMessageBox top').toBeLessThanOrEqual(
    messageTop + 0.5,
  );
});

test('Congress tableau cards stay above the action log when idle and selected', async ({ page }) => {
  await navigateTo(page, '/congress');
  const cards = page.locator('[data-tutorial="cg-tableau"] img');
  const actionLogButton = page.getByRole('button', { name: '棋譜を見る' });
  await expect(actionLogButton).toBeVisible();

  const assertTableauClearsActionLog = async () => {
    const bottom = await maxImageBottom(cards);
    const buttonTop = await actionLogButton.evaluate((element) => element.getBoundingClientRect().top);
    expect(bottom, 'Congress tableau card bottoms should not pass the action log button top').toBeLessThanOrEqual(
      buttonTop + 0.5,
    );
  };

  await assertTableauClearsActionLog();
  const source = page.locator('[data-tutorial="cg-tableau"] button[aria-pressed]').nth(2);
  await source.click();
  await expect(source).toHaveAttribute('aria-pressed', 'true');
  await assertTableauClearsActionLog();
});
