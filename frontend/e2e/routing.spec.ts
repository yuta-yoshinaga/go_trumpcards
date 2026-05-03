import { expect, test } from '@playwright/test';
import { TIMEOUT_TRANSITION, waitForLoaded } from './helpers';

test.describe('Routing E2E', () => {
  test('/blackjack deep-link redirects to BlackJack at /', async ({ page }) => {
    await page.addInitScript(() => {
      localStorage.setItem('tutorial_no_suggest', 'true');
    });
    await page.goto('/#/blackjack');
    await waitForLoaded(page);
    await expect(page).toHaveURL(/#\/?$/, { timeout: TIMEOUT_TRANSITION });
    await expect(page.getByRole('button', { name: 'ベット' })).toBeVisible({
      timeout: TIMEOUT_TRANSITION,
    });
  });

  test('inline FOUC script honors persisted i18n_lang=en', async ({ page }) => {
    await page.addInitScript(() => {
      localStorage.setItem('i18n_lang', 'en');
      localStorage.setItem('tutorial_no_suggest', 'true');
    });
    await page.goto('/');
    await expect(page.locator('html')).toHaveAttribute('lang', 'en', {
      timeout: TIMEOUT_TRANSITION,
    });
  });
});
