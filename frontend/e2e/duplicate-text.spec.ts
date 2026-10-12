import { expect, test } from '@playwright/test';
import { navigateTo } from './helpers';

const viewports = [
  { name: 'mobile', width: 375, height: 667 },
  { name: 'desktop', width: 1280, height: 800 },
];

for (const viewport of viewports) {
  test.describe(`duplicate visible text (${viewport.name})`, () => {
    test.use({ viewportSize: { width: viewport.width, height: viewport.height } });

    for (const path of ['/cucumber', '/comet', '/teenpatti', '/threecardbrag', '/julepe']) {
      test(`${path} does not repeat game guidance elsewhere`, async ({ page }) => {
        await navigateTo(page, path);
        const duplicates = await page.locator('main').evaluate((main) => {
          const boxes = main.querySelectorAll<HTMLElement>(
            '[role="status"].glass-panel, [role="alert"].glass-panel, [role="status"] .glass-panel, [role="alert"] .glass-panel',
          );
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
              if (
                element.closest(
                  '[role="status"].glass-panel, [role="alert"].glass-panel, [role="status"] .glass-panel, [role="alert"] .glass-panel',
                )
              )
                return false;
              return element.children.length === 0 && element.innerText.trim().replace(/\s+/g, ' ') === message;
            });
            return repeated ? [message] : [];
          });
          return { messages, duplicates };
        });
        expect(duplicates.messages, `Expected a non-empty GameMessageBox guidance message on ${path}`).not.toEqual([]);
        expect(duplicates.duplicates, `GameMessageBox guidance is repeated elsewhere on ${path}`).toEqual([]);
      });
    }
  });
}
