import { readFileSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { expect, test } from '@playwright/test';
import { isAllowedLabelWrap } from '../src/utils/labelWrapAllowlist';
import { navigateTo } from './helpers';

/**
 * Guards #11824: short labels can wrap in the middle of a word when squeezed.
 * Detect visible 2–10 character text nodes in `main` whose adjacent non-space
 * characters land on different lines; whitespace line wraps are allowed. The
 * layout fix uses whitespace-nowrap units with flex-wrap and must not restore
 * a minimum width (#11729). Ordinary sentence and paytable exceptions are
 * separately documented and limited to the observed mobile viewport.
 */

// Match desktop-action-visibility.spec.ts: derive paths from the route source.
const gameRoutesSource = readFileSync(
  resolve(dirname(fileURLToPath(import.meta.url)), '../src/constants/gameRoutes.ts'),
  'utf8',
);
const routes = [...gameRoutesSource.matchAll(/path:\s*'([^']+)'/g)].map(([, path]) => path);
const VIEWPORTS = [
  { width: 375, height: 667 },
  { width: 1280, height: 800 },
] as const;

const NOWRAP_VIEWPORTS = [
  { width: 375, height: 667 },
  { width: 1024, height: 768 },
  { width: 1280, height: 800 },
] as const;

test('route list is complete', () => {
  expect(routes.length, 'game route path declarations found').toBeGreaterThanOrEqual(300);
});

for (const viewport of VIEWPORTS) {
  test.describe(`${viewport.width}x${viewport.height}`, () => {
    test.use({ viewport });

    for (const path of routes) {
      test(`${path} has no short labels wrapped mid-word`, async ({ page }) => {
        await navigateTo(page, path);
        const result = await page.evaluate(
          ({ minLength, maxLength, topTolerance, minOpacity }) => {
            const out: string[] = [];
            let checked = 0;
            let spaceBreaks = 0;
            const walker = document.createTreeWalker(
              document.querySelector('main') || document.body,
              NodeFilter.SHOW_TEXT,
            );
            while (true) {
              const node = walker.nextNode();
              if (!node) break;
              const raw = node.nodeValue ?? '';
              const text = raw.trim();
              if (!text || text.length < minLength || text.length > maxLength) continue;
              const element = node.parentElement;
              if (!element || element.closest('.sr-only,[aria-hidden="true"],button')) continue;
              // Chromium keeps layout rects for closed <details> content; skip it unless in <summary>.
              const details = element.closest('details');
              if (details && !details.open && !element.closest('summary')) continue;
              let opacity = 1;
              let hidden = false;
              for (let ancestor: HTMLElement | null = element; ancestor; ancestor = ancestor.parentElement) {
                const style = getComputedStyle(ancestor);
                opacity *= Number.parseFloat(style.opacity);
                if (style.visibility === 'hidden' || style.display === 'none' || style.contentVisibility === 'hidden') {
                  hidden = true;
                }
              }
              if (hidden || opacity < minOpacity) continue;
              const range = document.createRange();
              range.selectNodeContents(node);
              if (![...range.getClientRects()].some((rect) => rect.width > 0)) continue;
              checked++;
              let previousTop: number | null = null;
              let broke = false;
              let midword = false;
              for (let index = 0; index < raw.length; index++) {
                if (/\s/.test(raw[index] ?? '')) continue;
                const character = document.createRange();
                character.setStart(node, index);
                character.setEnd(node, index + 1);
                const rect = character.getClientRects()[0];
                if (!rect || rect.width === 0) continue;
                const top = Math.round(rect.top);
                if (previousTop !== null && Math.abs(top - previousTop) > topTolerance) {
                  broke = true;
                  if (!/\s/.test(raw[index - 1] ?? ' ')) midword = true;
                }
                previousTop = top;
              }
              if (midword) out.push(text);
              else if (broke) spaceBreaks++;
            }
            return { checked, spaceBreaks, wrapped: [...new Set(out)] };
          },
          { minLength: 2, maxLength: 10, topTolerance: 4, minOpacity: 0.1 },
        );

        if (result.checked === 0) {
          test.info().annotations.push({ type: 'skipped page', description: `${path}: no eligible text nodes` });
        }
        const failures = result.wrapped.filter((label) => !isAllowedLabelWrap(path, label, viewport));
        expect(
          failures,
          `${path} (${viewport.width}x${viewport.height}) has short text nodes wrapped mid-word; checked ${result.checked}, whitespace wraps ${result.spaceBreaks}`,
        ).toEqual([]);
      });
    }
  });
}

for (const viewport of NOWRAP_VIEWPORTS) {
  test.describe(`nowrap overflow ${viewport.width}x${viewport.height}`, () => {
    test.use({ viewport });

    for (const path of routes) {
      test(`${path} nowrap text stays within its nearest parent content box`, async ({ page }) => {
        await navigateTo(page, path);
        const result = await page.evaluate(() => {
          const out: string[] = [];
          let checked = 0;
          const main = document.querySelector('main') || document.body;
          for (const element of main.querySelectorAll<HTMLElement>('*')) {
            if (![...element.childNodes].some((node) => node.nodeType === Node.TEXT_NODE && !!node.textContent?.trim()))
              continue;
            const style = getComputedStyle(element);
            if (!/nowrap|pre/.test(`${style.whiteSpace} ${style.textWrapMode ?? ''}`)) continue;
            // Absolutely positioned labels are intentional placements outside their parent box.
            if (style.position === 'absolute' || style.position === 'fixed') continue;
            if (
              style.overflow === 'hidden' ||
              style.overflow === 'clip' ||
              style.overflowX === 'hidden' ||
              style.overflowX === 'clip'
            )
              continue;
            if (element.closest('[aria-hidden="true"],.sr-only,[role="tooltip"]')) continue;
            const details = element.closest('details');
            if (details && !details.open && !element.closest('summary')) continue;
            let opacity = 1;
            let invisible = false;
            for (let ancestor: HTMLElement | null = element; ancestor; ancestor = ancestor.parentElement) {
              const ancestorStyle = getComputedStyle(ancestor);
              opacity *= Number.parseFloat(ancestorStyle.opacity);
              if (
                ancestorStyle.visibility === 'hidden' ||
                ancestorStyle.display === 'none' ||
                ancestorStyle.contentVisibility === 'hidden'
              )
                invisible = true;
            }
            const rect = element.getBoundingClientRect();
            if (invisible || opacity < 0.05 || rect.width < 1 || rect.height < 1) continue;
            checked++;
            let parent = element.parentElement;
            while (parent && ['contents', 'inline'].includes(getComputedStyle(parent).display))
              parent = parent.parentElement;
            if (!parent) continue;
            const parentStyle = getComputedStyle(parent);
            // Only the nearest box can intentionally scroll this label horizontally.
            // Ancestor overflow-y:auto can compute overflow-x:visible to auto and hide real defects.
            const parentRect = parent.getBoundingClientRect();
            const right =
              parentRect.right -
              Number.parseFloat(parentStyle.paddingRight) -
              Number.parseFloat(parentStyle.borderRightWidth);
            const left =
              parentRect.left +
              Number.parseFloat(parentStyle.paddingLeft) +
              Number.parseFloat(parentStyle.borderLeftWidth);
            const overflow = Math.max(rect.right - right, left - rect.left);
            if (overflow > 3)
              out.push(
                `${element.textContent?.trim().replace(/\s+/g, ' ').slice(0, 24)}|+${Math.round(overflow)}px|w${Math.round(parentRect.width)}`,
              );
          }
          return { checked, out };
        });
        if (result.checked === 0) {
          test.info().annotations.push({ type: 'skipped page', description: `${path}: no eligible nowrap text nodes` });
        }
        if (path === '/binokel') {
          expect(result.checked, `${path} eligible nowrap text nodes checked`).toBeGreaterThanOrEqual(1);
        }
        expect(
          result.out,
          `${path} (${viewport.width}x${viewport.height}) has nowrap text outside its nearest parent content box`,
        ).toEqual([]);
      });
    }
  });
}
