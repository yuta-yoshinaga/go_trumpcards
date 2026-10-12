import { expect, test } from '@playwright/test';
import { navigateTo } from './helpers';

/** #11702: use the same 52px visible-card rule as #11655 for desktop boards. */
test.use({ viewport: { width: 1280, height: 800 } });

const PAGES = ['/grandfathersclock', '/bigben', '/colorado', '/slyfox'] as const;

for (const path of PAGES) {
  test(`${path} keeps every board card visible on desktop`, async ({ page }) => {
    await navigateTo(page, path);
    await page.waitForFunction(() =>
      document
        .getAnimations()
        .every(
          (animation) =>
            animation.playState === 'finished' ||
            animation.playState === 'idle' ||
            animation.effect?.getComputedTiming().iterations === Number.POSITIVE_INFINITY,
        ),
    );
    const result = await page.evaluate(() => {
      const cards = [...document.querySelectorAll('main img, main [role="img"]')].filter((element) => {
        const label = `${element.getAttribute('alt') ?? ''} ${element.getAttribute('aria-label') ?? ''}`;
        const box = element.getBoundingClientRect();
        return box.width >= 20 && box.height >= 30 && !/裏|back/i.test(label);
      });
      const hidden = cards.flatMap((element) => {
        const box = element.getBoundingClientRect();
        let visibleBottom = window.innerHeight;
        for (let ancestor = element.parentElement; ancestor; ancestor = ancestor.parentElement) {
          const style = getComputedStyle(ancestor);
          if (/^(auto|scroll|hidden)$/.test(style.overflowY) && ancestor.scrollHeight > ancestor.clientHeight) {
            visibleBottom = Math.min(visibleBottom, ancestor.getBoundingClientRect().bottom);
          }
        }
        return box.top + 52 > visibleBottom ? [Math.round(box.top)] : [];
      });
      return { count: cards.length, hidden };
    });
    expect(result.count, `${path}: no board cards found`).toBeGreaterThan(0);
    expect(result.hidden, `${path}: board cards hidden below their play area`).toEqual([]);
  });
}
