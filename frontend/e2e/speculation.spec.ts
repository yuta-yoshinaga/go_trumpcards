import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Speculation E2E', () => {
  test('turns cards up until the round settles', async ({ page }) => {
    await navigateTo(page, '/speculation');

    // The face-down count is public; the cards themselves are not.
    await expect(page.getByTestId('sp-seat-0')).toBeVisible();
    await expect(page.getByTestId('sp-hidden-0')).toBeVisible();
    await expect(page.getByTestId('sp-trump')).toBeVisible();

    // Flip repeatedly; an auction may open at any point, so answer it and
    // carry on. The round is over when 次のラウンドへ appears.
    const flip = page.getByRole('button', { name: 'めくる' });
    const next = page.getByRole('button', { name: '次のラウンドへ' });
    const decline = page.getByRole('button', { name: '断る' });

    await expect(flip).toBeVisible();
    for (let i = 0; i < 20; i++) {
      if (await next.isVisible().catch(() => false)) break;
      if (await decline.isVisible().catch(() => false)) {
        await decline.click();
      } else if (await flip.isVisible().catch(() => false)) {
        await flip.click();
      }
      await waitForLoaded(page);
    }

    await expect(next).toBeVisible({ timeout: 10_000 });
  });

  test('answering an auction keeps the round moving', async ({ page }) => {
    await navigateTo(page, '/speculation');

    const flip = page.getByRole('button', { name: 'めくる' });
    const accept = page.getByRole('button', { name: '受ける' });
    const next = page.getByRole('button', { name: '次のラウンドへ' });
    await expect(flip).toBeVisible();

    // Drive until an auction opens, then accept it. **Accepting must not
    // dead-end the round** — the flip control has to come back (or the round
    // has to finish).
    let sawAuction = false;
    for (let i = 0; i < 20; i++) {
      if (await accept.isVisible().catch(() => false)) {
        sawAuction = true;
        await accept.click();
        await waitForLoaded(page);
        break;
      }
      if (await next.isVisible().catch(() => false)) break;
      if (await flip.isVisible().catch(() => false)) {
        await flip.click();
        await waitForLoaded(page);
      }
    }

    if (sawAuction) {
      await expect(flip.or(next)).toBeVisible({ timeout: 10_000 });
    }
  });
});
