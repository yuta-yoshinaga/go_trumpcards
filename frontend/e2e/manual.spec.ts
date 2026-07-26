import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_TRANSITION, waitForLoaded } from './helpers';

// The manual modal had no browser coverage at all before this. It needs some
// now because each manual became its own lazily-fetched chunk: the unit tests
// exercise the loader through Vite's dev transform, which cannot tell us that
// the chunk resolves in a production build. A broken dynamic import would show
// up as a modal stuck on its loading line — green unit tests, blank manual.
test.describe('Game manual', () => {
  test('opens the manual modal and renders the fetched markdown', async ({ page }) => {
    await navigateTo(page, '/poker');
    await waitForLoaded(page);

    await page.getByRole('button', { name: 'マニュアル' }).click();

    const dialog = page.getByRole('dialog', { name: 'ゲームマニュアル' });
    await expect(dialog).toBeVisible();

    // A heading only exists once the markdown has been fetched AND parsed, so
    // this fails if the chunk never arrives. Asserting on a heading rather
    // than on wording keeps it independent of the manual's copy.
    await expect(dialog.getByRole('heading').first()).toBeVisible({ timeout: TIMEOUT_TRANSITION });

    await page.getByRole('button', { name: '閉じる' }).click();
    await expect(dialog).not.toBeVisible();
  });
});
