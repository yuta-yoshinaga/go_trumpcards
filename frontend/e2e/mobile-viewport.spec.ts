import { expect, test } from '@playwright/test';
import { navigateTo } from './helpers';

/**
 * A game page must fit a 375x667 phone and scroll inside its play area, never by
 * growing the document (#1861, #4373).
 *
 * This has regressed five times (#965, #993, #1058, #1367, #4373). The static half
 * of the contract — the two load-bearing classes — is guarded by
 * `scripts/check-design-tokens.mjs`; this is the half only a real browser can
 * check, because it catches a *page* adding tall `shrink-0` content outside its
 * scroll region, which no static check can see.
 */
test.use({ viewport: { width: 375, height: 667 } });

// A spread of layout shapes rather than a long list: the two tallest footers
// (watten 558px, spades 511px before the cap), a trick-taking page that used to
// fit at exactly 667, a deep tableau, a page with tall CPU meld display, a
// solitaire, a betting page, and the default route.
//
// Each runs against one unseeded deal, and page height IS deal-dependent — `/pan`
// varies by 46px across deals, and elsewhere `jass` varies by 200px. That does not
// make these assertions flaky, because what they assert is deal-independent: once
// the shell constrains the column, extra content lands in the `overflow-y-auto`
// play area and cannot change the document height. A flake here would therefore
// mean a page grew tall `shrink-0` content *outside* its play area — which is a
// real regression and exactly what this spec is for, so investigate rather than
// retry.
const PATHS = [
  '/watten',
  '/spades',
  '/hearts',
  '/crescent',
  '/pan',
  '/freecell',
  '/poker',
  '/',
  // These five had no play area at all until #4373 phase 3 — their content went
  // straight into the document height. `/speed` is the reason they are guarded
  // rather than trusted: it fitted on some deals and overflowed by 201px on
  // others, so it looked fine whenever it was spot-checked.
  '/speed',
  '/contractrummy',
  '/piquet',
  '/kalooki',
  '/carioca',
  // These six were the last holdouts (#4373 phase 4). Their children measurably
  // fitted — 606px in a 605px box — yet the document still grew, because stray
  // scrollable overflow from a descendant propagated past the shell column. They
  // are guarded because that is invisible to any static check and was only found
  // by bisecting the DOM in a real browser.
  '/memory',
  '/openfacechinese',
  '/jass',
  '/clocksolitaire',
  '/gaigel',
  '/spiteandmalice',
];

for (const path of PATHS) {
  test(`${path} does not scroll the document at 375x667`, async ({ page }) => {
    await navigateTo(page, path);
    const { docH, innerH } = await page.evaluate(() => ({
      docH: document.documentElement.scrollHeight,
      innerH: window.innerHeight,
    }));
    expect(docH, `${path} grew the document to ${docH}px inside a ${innerH}px viewport`).toBeLessThanOrEqual(innerH);
  });

  test(`${path} keeps its action footer within the mobile cap`, async ({ page }) => {
    await navigateTo(page, path);
    const footer = await page.evaluate(() => {
      const f = document.querySelector('main footer');
      if (!f) return null;
      return { h: Math.round(f.getBoundingClientRect().height), vh: window.innerHeight };
    });
    // Not every page has an action footer; those that do must leave the play area
    // usable, which is what the 45vh cap buys.
    if (footer === null) return;
    expect(footer.h, `${path} footer is ${footer.h}px of a ${footer.vh}px viewport`).toBeLessThanOrEqual(
      Math.ceil(footer.vh * 0.45) + 1,
    );
  });
}
