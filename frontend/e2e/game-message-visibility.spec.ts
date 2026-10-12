import { expect, test } from '@playwright/test';
import { navigateTo } from './helpers';

const MESSAGE_PAGES = ['/literature', '/kaiser', '/sueca', '/spades'] as const;
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

  for (const path of MESSAGE_PAGES) {
    test(`${path} keeps its game message fully visible`, async ({ page }) => {
      await navigateTo(page, path);
      const message = page.getByTestId('game-footer-message').locator('[role="status"], [role="alert"]').first();
      await expect(page.getByTestId('game-footer-message')).toBeAttached();
      const content = await message.textContent();
      expect(content?.trim()).toBeTruthy();

      const rect = await message.boundingBox();
      expect(rect, `${path} message should have a rendered box`).not.toBeNull();
      if (!rect) throw new Error(`${path} message should have a rendered box`);
      expect(rect.x).toBeGreaterThanOrEqual(0);
      expect(rect.y).toBeGreaterThanOrEqual(0);
      expect(rect.x + rect.width).toBeLessThanOrEqual(375);
      expect(rect.y + rect.height).toBeLessThanOrEqual(667);

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
        expect(rect.y, `${path} message should fit inside its scroll parent`).toBeGreaterThanOrEqual(scrollport.top);
        expect(rect.y + rect.height, `${path} message should fit inside its scroll parent`).toBeLessThanOrEqual(
          scrollport.bottom,
        );
      } else {
        expect(rect.y, `${path} has no scroll parent; checking viewport only`).toBeGreaterThanOrEqual(0);
        expect(rect.y + rect.height, `${path} has no scroll parent; checking viewport only`).toBeLessThanOrEqual(667);
      }
    });
  }
});
