import { readFileSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { expect, test } from '@playwright/test';
import { navigateTo } from './helpers';

const viewports = [
  { name: 'mobile', width: 375, height: 667 },
  { name: 'desktop', width: 1280, height: 800 },
];
const guidanceSelector =
  '[role="status"].glass-panel, [role="alert"].glass-panel, [role="status"] .glass-panel, [role="alert"] .glass-panel';

// Read the canonical game route declarations without loading React modules.
const gameRoutesSource = readFileSync(
  resolve(dirname(fileURLToPath(import.meta.url)), '../src/constants/gameRoutes.ts'),
  'utf8',
);
const allGamePaths = [...gameRoutesSource.matchAll(/path:\s*'([^']+)'/g)].map(([, path]) => path);

for (const viewport of viewports) {
  test.describe(`duplicate visible text (${viewport.name})`, () => {
    test.use({ viewportSize: { width: viewport.width, height: viewport.height } });

    for (const path of allGamePaths) {
      test(`${path} does not repeat game guidance elsewhere`, async ({ page }) => {
        await navigateTo(page, path);
        const duplicates = await page.locator('main').evaluate((main, selector) => {
          const boxes = main.querySelectorAll<HTMLElement>(selector);
          const messages = [...boxes].map((box) => box.innerText.trim().replace(/\s+/g, ' ')).filter(Boolean);
          const visibleText = [...main.querySelectorAll<HTMLElement>('*')].filter((element) => {
            const style = getComputedStyle(element);
            return (
              style.display !== 'none' &&
              style.visibility !== 'hidden' &&
              !element.closest('[aria-hidden="true"]') &&
              !element.closest('.sr-only')
            );
          });
          const duplicates = messages.flatMap((message) => {
            const repeated = visibleText.some((element) => {
              if (element.closest(selector)) return false;
              // Only compare leaf elements with the exact guidance text; concatenated or paraphrased text is out of scope.
              return element.children.length === 0 && element.innerText.trim().replace(/\s+/g, ' ') === message;
            });
            return repeated ? [message] : [];
          });
          return { messages, duplicates };
        }, guidanceSelector);
        if (duplicates.messages.length === 0) {
          test.info().annotations.push({
            type: 'note',
            description: 'No initial GameMessageBox text; duplicate comparison skipped.',
          });
          return;
        }
        expect(duplicates.duplicates, `GameMessageBox guidance is repeated elsewhere on ${path}`).toEqual([]);
      });
    }
  });
}
