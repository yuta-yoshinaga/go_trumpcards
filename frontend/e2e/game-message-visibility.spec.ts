import { expect, test } from '@playwright/test';
import { navigateTo } from './helpers';

// Every page #11728 moved into the footer message slot that shows guidance on load (Somerset only
// shows one when stuck, so its slot is covered by SomersetPage.test.tsx instead).
const MESSAGE_PAGES = [
  '/literature',
  '/kaiser',
  '/sueca',
  '/spades',
  '/tute',
  '/marias',
  '/minchiate',
  '/scarto',
  '/bridge',
  '/catchten',
  '/belote',
  '/jass',
  '/watten',
  '/gaigel',
  '/omi',
  '/binokel',
  '/vint',
  '/karnoffel',
  '/unsunkaruta',
  '/quodlibet',
  '/dehlapakad',
  '/sutda',
  '/boston',
  '/klaberjass',
  '/kille',
  '/duchess',
  '/sthelena',
  '/indianrummy',
  '/marriage',
  '/machiavelli',
  '/pan',
  '/samba',
  '/bolivia',
  '/burraco',
  '/biriba',
  '/doppelkopf',
  '/quadrille',
  '/whist',
  '/bidwhist',
  '/fortyfives',
  '/bouillotte',
  '/marjapussi',
  '/bideuchre',
  '/sixbidsolo',
  '/saliclaw',
  '/yaniv',
  '/madrasso',
  '/tongits',
  '/fivehundred',
  '/twentynine',
  '/ohhell',
  '/wizard',
  '/napoleon',
  '/mighty',
  '/callbreak',
  '/batak',
  '/twotenjack',
  '/ninetynine',
  '/openfacechinese',
  '/canasta',
] as const;
test.use({ viewport: { width: 1280, height: 800 } });

for (const path of MESSAGE_PAGES) {
  test(`${path} keeps its game message fully visible on desktop`, async ({ page }) => {
    await navigateTo(page, path);
    const message = page.getByTestId('game-footer-message').locator('[role="status"], [role="alert"]');
    await expect(page.getByTestId('game-footer-message')).toBeAttached();
    const content = await message.textContent();
    expect(content?.trim()).toBeTruthy();

    const rect = await message.boundingBox();
    expect(rect, `${path} message should have a rendered box`).not.toBeNull();
    if (!rect) throw new Error(`${path} message should have a rendered box`);
    expect(rect.x).toBeGreaterThanOrEqual(0);
    expect(rect.y).toBeGreaterThanOrEqual(0);
    expect(rect.x + rect.width).toBeLessThanOrEqual(1280);
    expect(rect.y + rect.height).toBeLessThanOrEqual(800);

    const scrollport = await message.evaluate((element) => {
      let parent = element.parentElement;
      while (parent) {
        const overflowY = getComputedStyle(parent).overflowY;
        if (overflowY === 'auto' || overflowY === 'scroll') {
          const bounds = parent.getBoundingClientRect();
          return { top: bounds.top, bottom: bounds.bottom };
        }
        parent = parent.parentElement;
      }
      return null;
    });
    if (scrollport) {
      expect(rect.y).toBeGreaterThanOrEqual(scrollport.top);
      expect(rect.y + rect.height).toBeLessThanOrEqual(scrollport.bottom);
    }
  });
}

test.describe('375x667 mobile viewport', () => {
  test.use({ viewport: { width: 375, height: 667 } });

  // Below sm the footer is capped at 45vh, so the guidance stays in the play area there and
  // leaves the footer to the hand (hand-visibility.spec.ts guards the hand itself).
  for (const path of MESSAGE_PAGES) {
    test(`${path} keeps its game message out of the footer on a phone`, async ({ page }) => {
      await navigateTo(page, path);
      const message = page
        .locator('main :is([role="status"], [role="alert"]):is(.glass-panel, :has(.glass-panel))')
        .filter({ hasText: /\S/ })
        .first();
      await expect(message).toBeAttached();
      expect((await message.textContent())?.trim()).toBeTruthy();
      await expect(page.getByTestId('game-footer-message')).toHaveCount(0);
      expect(await message.evaluate((element) => element.closest('footer') === null)).toBe(true);
    });
  }
});
