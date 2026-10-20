import { expect, type Page, test } from '@playwright/test';
import { DESKTOP_VIEWPORTS } from './desktopViewports';
import { gameButton, navigateTo, TIMEOUT_ACTION, waitForLoaded } from './helpers';
import { measureCoveredElements } from './sweeps/cover';

async function enterPitchPlayPhase(page: Page) {
  const playButton = page.getByRole('button', { name: '出す' });
  // Pitch has four seats. Eight passes allow two full seat rotations for transition/render timing.
  for (let attempt = 0; attempt < 8; attempt += 1) {
    if (await playButton.isVisible().catch(() => false)) return;
    const pass = gameButton(page, 'パス');
    await expect(pass).toBeVisible({ timeout: TIMEOUT_ACTION });
    await pass.click();
    await waitForLoaded(page);
  }
  await expect(playButton).toBeVisible({ timeout: TIMEOUT_ACTION });
}

for (const viewport of DESKTOP_VIEWPORTS) {
  test.describe(`${viewport.width}x${viewport.height}`, () => {
    test.use({ viewport });

    test('/pitch has no sticky or fixed cover after passing the bid phase', async ({ page }) => {
      await navigateTo(page, '/pitch');
      await enterPitchPlayPhase(page);
      const result = await measureCoveredElements(page);
      test.info().annotations.push({
        type: result.count === 0 ? 'skipped page' : 'measured targets',
        description: `/pitch play phase: ${result.count} targets`,
      });
      expect(result.count, '/pitch play phase should expose measurable targets').toBeGreaterThan(0);
      expect(result.covered, '/pitch: covered targets (label, y, coverer testid/class)').toEqual([]);
    });
  });
}
