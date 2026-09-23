import { expect, test } from '@playwright/test';
import { navigateTo } from './helpers';

test.describe('manual modal blocks Hearts hand shortcuts', () => {
  test('selects a hand card with a number key when no modal is open', async ({ page }) => {
    await navigateTo(page, '/hearts');
    const handCards = page.locator('button[aria-pressed]:has(img)');
    await expect(handCards.first()).toBeVisible();

    await page.keyboard.press('1');

    await expect(handCards.nth(0)).toHaveAttribute('aria-pressed', 'true');
  });

  test('does not select hand cards with number keys while the manual is open', async ({ page }) => {
    await navigateTo(page, '/hearts');
    const handCards = page.locator('button[aria-pressed]:has(img)');
    await expect(handCards.first()).toBeVisible();

    await page.getByRole('button', { name: 'マニュアル', exact: true }).click();
    await expect(page.getByRole('dialog')).toBeVisible();
    await page.keyboard.press('1');
    await page.keyboard.press('2');

    await expect(handCards.nth(0)).toHaveAttribute('aria-pressed', 'false');
    await expect(handCards.nth(1)).toHaveAttribute('aria-pressed', 'false');
  });

  test('keeps the selected hand card after closing the manual with Escape', async ({ page }) => {
    await navigateTo(page, '/hearts');
    const handCards = page.locator('button[aria-pressed]:has(img)');
    await expect(handCards.first()).toBeVisible();
    await handCards.nth(0).click();
    await expect(handCards.nth(0)).toHaveAttribute('aria-pressed', 'true');

    await page.getByRole('button', { name: 'マニュアル', exact: true }).click();
    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible();

    await page.keyboard.press('Escape');

    await expect(dialog).toBeHidden();
    await expect(handCards.nth(0)).toHaveAttribute('aria-pressed', 'true');
  });
});
