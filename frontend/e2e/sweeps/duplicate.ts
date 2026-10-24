import { expect, type Page, test } from '@playwright/test';

const guidanceSelector =
  '[role="status"].glass-panel, [role="alert"].glass-panel, [role="status"] .glass-panel, [role="alert"] .glass-panel';

/** Asserts that initial GameMessageBox guidance is not repeated verbatim elsewhere on the page. */
export async function checkDuplicate(
  page: Page,
  path: string,
  _viewport: { width: number; height: number },
): Promise<void> {
  // duplicate-text.spec.ts historically skipped the root '/' page.
  if (path === '/') return;

  const result = await page.evaluate(
    ({ selector }) => {
      const main = document.querySelector('main');
      if (!main) return { mainFound: false, messages: [], duplicates: [] };
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
      return { mainFound: true, messages, duplicates };
    },
    { selector: guidanceSelector },
  );

  if (!result.mainFound) {
    expect.soft(result.mainFound, `Duplicate check: ${path}: game main content was not found`).toBe(true);
    return;
  }
  if (result.messages.length === 0) {
    test.info().annotations.push({
      type: 'note',
      description: `Duplicate check: ${path}: No initial GameMessageBox text; duplicate comparison skipped.`,
    });
    return;
  }
  expect
    .soft(result.duplicates, `Duplicate check: GameMessageBox guidance is repeated elsewhere on ${path}`)
    .toEqual([]);
}
