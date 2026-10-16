import { readFileSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { expect, type Page, test } from '@playwright/test';
import { DESKTOP_VIEWPORTS } from './desktopViewports';
import { gameButton, navigateTo, TIMEOUT_ACTION, waitForLoaded } from './helpers';

const gameRoutesSource = readFileSync(
  resolve(dirname(fileURLToPath(import.meta.url)), '../src/constants/gameRoutes.ts'),
  'utf8',
);
const routes = [...new Set([...gameRoutesSource.matchAll(/path:\s*'([^']+)'/g)].map(([, path]) => path))];

// Keep empty only while every game route is expected to render a button or card target.
// Add a route here only with a comment explaining why that page has no measurable target.
const ZERO_TARGET_ROUTES = new Set<string>([]);

test('reads the full game route list for the cover sweep', () => {
  expect(routes.length, 'game route path declarations found').toBeGreaterThanOrEqual(360);
  expect(new Set(routes).size, 'duplicate game routes removed').toBe(routes.length);
  expect(
    routes.every((path) => path.startsWith('/')),
    'every game route starts with /',
  ).toBe(true);
});

/** #11826: existing desktop guards check viewport bounds but miss a sticky hand painting over the board. */
async function measureCoveredElements(page: Page) {
  return page.evaluate(() => {
    const targets = new Set<Element>();
    for (const element of document.querySelectorAll('main img, main [role="img"], main button:not(:disabled)')) {
      const box = element.getBoundingClientRect();
      if (box.width < 20 || box.height < 20) continue;
      const label = `${element.getAttribute('alt') ?? ''} ${element.getAttribute('aria-label') ?? ''}`.trim();
      if (element.tagName !== 'BUTTON' && /裏|back/i.test(label)) continue;
      targets.add(element);
    }
    // GameMessageBox renders its visible panel inside a status/alert live region.
    for (const element of document.querySelectorAll(
      'main [role="status"] .glass-panel, main [role="alert"] .glass-panel',
    )) {
      const box = element.getBoundingClientRect();
      if (box.width >= 20 && box.height >= 20) targets.add(element);
    }

    const covered: { label: string; y: number; coverer: string }[] = [];
    for (const element of targets) {
      const box = element.getBoundingClientRect();
      // One point per target is an intentional tradeoff: #11826's cover painted the board's upper
      // edge, so the top-center point (12px down) catches that regression at low test cost.
      const x = box.left + box.width / 2;
      const y = box.top + Math.min(12, box.height / 2);
      // Off-screen targets are outside this viewport's visible board and are deliberately ignored.
      if (y < 0 || y > innerHeight || x < 0 || x > innerWidth) continue;
      const hit = document.elementFromPoint(x, y);
      if (!hit || element.contains(hit) || hit.contains(element)) continue;
      let coverer: Element | null = null;
      for (
        let ancestor: Element | null = hit;
        ancestor && ancestor !== document.body;
        ancestor = ancestor.parentElement
      ) {
        const position = getComputedStyle(ancestor).position;
        if (position === 'sticky' || position === 'fixed') {
          coverer = ancestor;
          break;
        }
      }
      if (coverer && !coverer.contains(element)) {
        const label = (
          element.getAttribute('alt') ||
          element.getAttribute('aria-label') ||
          element.textContent ||
          element.tagName
        )
          .trim()
          .replace(/\s+/g, ' ')
          .slice(0, 80);
        const covererLabel =
          `${coverer.getAttribute('data-testid') ?? ''} ${typeof coverer.className === 'string' ? coverer.className : ''}`.trim();
        covered.push({ label, y: Math.round(y), coverer: covererLabel || coverer.tagName.toLowerCase() });
      }
    }
    return { count: targets.size, covered };
  });
}

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
    for (const path of routes) {
      test(`${path} has no sticky or fixed cover over board targets`, async ({ page }) => {
        await navigateTo(page, path);
        const result = await measureCoveredElements(page);
        test.info().annotations.push({
          type: result.count === 0 ? 'skipped page' : 'measured targets',
          description: `${path}: ${result.count} targets`,
        });
        if (!ZERO_TARGET_ROUTES.has(path)) {
          expect(
            result.count,
            `${path}: no measurable targets; the selector or page rendering may be broken`,
          ).toBeGreaterThan(0);
        }
        expect(result.covered, `${path}: covered targets (label, y, coverer testid/class)`).toEqual([]);
      });
    }

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
