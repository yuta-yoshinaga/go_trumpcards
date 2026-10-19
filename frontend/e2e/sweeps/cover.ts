import { expect, type Page, test } from '@playwright/test';

// Keep empty only while every game route is expected to render a button or card target.
// Add a route here only with a comment explaining why that page has no measurable target.
export const ZERO_TARGET_ROUTES = new Set<string>([]);

/** #11826: existing desktop guards check viewport bounds but miss a sticky hand painting over the board. */
export async function measureCoveredElements(page: Page) {
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

/** Asserts that no sticky or fixed element covers board targets. */
export async function checkCover(
  page: Page,
  path: string,
  _viewport: { width: number; height: number },
): Promise<void> {
  const result = await measureCoveredElements(page);
  test.info().annotations.push({
    type: result.count === 0 ? 'skipped page' : 'measured targets',
    description: `Cover check: ${path}: ${result.count} targets`,
  });
  if (!ZERO_TARGET_ROUTES.has(path)) {
    expect
      .soft(result.count, `Cover check: ${path}: no measurable targets; the selector or page rendering may be broken`)
      .toBeGreaterThan(0);
  }
  expect.soft(result.covered, `Cover check: ${path}: covered targets (label, y, coverer testid/class)`).toEqual([]);
}
