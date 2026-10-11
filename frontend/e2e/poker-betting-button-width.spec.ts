import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_TRANSITION } from './helpers';

test('poker action button labels fit at 375px', async ({ page }) => {
  await page.setViewportSize({ width: 375, height: 667 });
  await navigateTo(page, '/poker');

  const actions = page.getByTestId('betting-actions');
  await expect(actions).toBeVisible({ timeout: TIMEOUT_TRANSITION });
  const buttons = actions.locator('button[aria-keyshortcuts]');
  await expect(buttons.first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });

  const widths = await buttons.evaluateAll((elements) =>
    elements.map((element) => ({
      label: element.textContent?.trim() ?? '',
      clientWidth: element.clientWidth,
      scrollWidth: element.scrollWidth,
    })),
  );
  expect(widths.length).toBeGreaterThan(0);
  for (const { label, clientWidth, scrollWidth } of widths) {
    expect(scrollWidth, `${label} should fit inside its button`).toBeLessThanOrEqual(clientWidth);
  }
});
