import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

const games = [
  { path: '/briscola', label: 'briscola-trump-label' },
  { path: '/brusquembille', label: 'brusquembille-trump-label' },
  { path: '/schnapsen', label: 'schnapsen-trump-label' },
];

for (const viewport of [
  { width: 375, height: 667 },
  { width: 1280, height: 800 },
]) {
  for (const game of games) {
    test(`${game.path} rotated trump does not overlap its label at ${viewport.width}px`, async ({ page }) => {
      await page.setViewportSize(viewport);
      await navigateTo(page, game.path);
      await waitForLoaded(page);

      const label = page.getByTestId(game.label);
      const card = page.getByTestId('rotated-trump-card');
      await expect(label).toBeVisible();
      await expect(card).toBeVisible();
      const labelRect = await label.boundingBox();
      const cardRect = await card.boundingBox();
      expect(labelRect).not.toBeNull();
      expect(cardRect).not.toBeNull();
      if (!labelRect || !cardRect) throw new Error('label and trump card must have bounding boxes');
      expect(labelRect.x + labelRect.width).toBeLessThanOrEqual(cardRect.x);
    });
  }
}
