import { expect, test } from '@playwright/test';
import { navigateTo, TIMEOUT_TRANSITION } from './helpers';
import { measureTextContrast } from './textContrast';

const pages = ['/cucumber', '/anaconda', '/deuceswild', '/spades', '/holdem', '/klondike'] as const;

// Measured at 375x667 after #11722 (stable over three runs). Lower a ceiling when a page improves;
// raising one needs a reason in the PR.
const overallViolationCeilings: Record<(typeof pages)[number], number> = {
  '/cucumber': 1,
  '/anaconda': 1,
  '/deuceswild': 5,
  '/spades': 0,
  '/holdem': 1,
  '/klondike': 0,
};

test.describe('text contrast regression guard', () => {
  test.use({ viewport: { width: 375, height: 667 } });

  test('measures composited backgrounds and classifies muted and colored-button text', async ({ page }) => {
    const measure = async (content: string) => {
      await page.setContent(content);
      return page.evaluate(measureTextContrast);
    };
    const shell = (content: string) => `
      <html style="--ds-text-muted-surface: #6b7280">
        <body style="margin: 0">
          <div style="background: #0b6838; padding: 24px">
            <div style="background: rgba(255,255,255,0.08); padding: 24px; font: 16px Arial">
              ${content}
            </div>
          </div>
        </body>
      </html>`;

    const bright = await measure(shell('<span style="color: #f8fafc">Bright text</span>'));
    expect(bright.checked).toBeGreaterThanOrEqual(1);
    expect(bright.violations).toHaveLength(0);

    const muted = await measure(shell('<span style="color: #6b7280">Muted text</span>'));
    expect(muted.checked).toBeGreaterThanOrEqual(1);
    expect(muted.violations.filter(({ kind }) => kind === 'muted').length).toBeGreaterThanOrEqual(1);

    const button = await measure(
      '<html><body style="margin:0"><button class="text-ds-text-on-accent" style="background:#facc15;color:#fff;font:16px Arial;padding:24px">Yellow button</button></body></html>',
    );
    expect(button.checked).toBeGreaterThanOrEqual(1);
    expect(button.violations.filter(({ kind }) => kind === 'colored-button').length).toBeGreaterThanOrEqual(1);
  });

  for (const path of pages) {
    test(`${path} has no muted or colored-button AA violations`, async ({ page }) => {
      await navigateTo(page, path);
      await expect(page.locator('[aria-busy="false"]')).toBeVisible({ timeout: TIMEOUT_TRANSITION });
      const { violations, checked } = await page.evaluate(measureTextContrast);
      const muted = violations.filter(({ kind }) => kind === 'muted');
      const coloredButtons = violations.filter(({ kind }) => kind === 'colored-button');

      expect(muted, `${path} muted text violations`).toEqual([]);
      expect(coloredButtons, `${path} colored button violations`).toEqual([]);
      expect(checked, `${path} text nodes checked`).toBeGreaterThanOrEqual(10);
      const ceiling = overallViolationCeilings[path];
      if (ceiling !== undefined) {
        expect(violations.length, `${path} overall AA violations`).toBeLessThanOrEqual(ceiling);
      }
    });
  }
});
