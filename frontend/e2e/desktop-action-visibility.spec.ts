import { readFileSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { expect, test } from '@playwright/test';
import { DESKTOP_VIEWPORTS } from './desktopViewports';
import { navigateTo } from './helpers';

const guidanceSelector =
  '[role="status"].glass-panel, [role="alert"].glass-panel, [role="status"] .glass-panel, [role="alert"] .glass-panel';

const marriageJa = JSON.parse(
  readFileSync(resolve(dirname(fileURLToPath(import.meta.url)), '../src/i18n/locales/ja/marriage.json'), 'utf8'),
) as { discardHistory: { show: string } };

// Cards stay visible in the hand and are not phase actions; exclude every card marker used by game pages.
const CARD_BUTTON_SELECTORS = [
  '[data-hand-card-index]',
  '[aria-pressed]',
  '[data-card]',
  '[data-card-index]',
  '[data-testid*="card"]',
] as const;
const CARD_IMAGE_ALT_PATTERN = /^(?:[♠♥♦♣]\s*)?(?:[2-9]|10|[AJQK])(?:\s*[♠♥♦♣])?$/i;

// Labels are the Japanese UI strings in src/i18n/locales/ja/common.json.
const AUXILIARY_ACTION_LABELS = [
  'リセット',
  '次のゲーム',
  'リセット確認',
  '投了確認',
  '投了',
  'ギブアップ',
  'ヒント',
  '棋譜を見る',
  '棋譜',
  '設定',
  'チュートリアル',
  'CLI',
  '説明を表示',
  '説明を非表示',
  'ヘルプ',
  'ルール',
  '言語',
  '日本語',
  'English',
  'ゲームナビゲーション',
  'お気に入りに追加',
  marriageJa.discardHistory.show,
] as const;

// Match desktop-board-visibility.spec.ts: derive paths from the route source, never maintain a list here.
const gameRoutesSource = readFileSync(
  resolve(dirname(fileURLToPath(import.meta.url)), '../src/constants/gameRoutes.ts'),
  'utf8',
);
const routes = [...gameRoutesSource.matchAll(/path:\s*'([^']+)'/g)].map(([, path]) => path);

test('route list is complete', () => {
  expect(routes.length, 'game route path declarations found').toBeGreaterThanOrEqual(300);
});

for (const viewport of DESKTOP_VIEWPORTS) {
  test.describe(`${viewport.width}x${viewport.height}`, () => {
    test.use({ viewport });
    for (const path of routes) {
      test(`${path} keeps every non-card, non-auxiliary action inside the desktop viewport`, async ({ page }) => {
        await navigateTo(page, path);
        const result = await page.evaluate(
          ({ selector, auxiliaryLabels, cardButtonSelectors, cardImageAltSource, cardImageAltFlags }) => {
            const guidance = [...document.querySelectorAll<HTMLElement>(selector)].some((element) => {
              const style = getComputedStyle(element);
              return (
                element.innerText.trim().length > 0 &&
                style.display !== 'none' &&
                style.visibility !== 'hidden' &&
                element.getBoundingClientRect().width > 0 &&
                element.getBoundingClientRect().height > 0
              );
            });
            const cardImageAlt = new RegExp(cardImageAltSource, cardImageAltFlags);
            if (!guidance) {
              return {
                guidance: false,
                outside: [] as { label: string; y: number }[],
                viewportBottom: window.innerHeight,
              };
            }

            const buttons = [...document.querySelectorAll<HTMLElement>('main button, footer button')].filter(
              (button) => {
                // #11825: aria-disabled primary actions remain visible because project policy (#8524/#8787) keeps the next action available in the UI until it becomes operable.
                if (button.matches(':disabled')) return false;
                const style = getComputedStyle(button);
                if (
                  style.display === 'none' ||
                  style.visibility === 'hidden' ||
                  button.closest('[aria-hidden="true"]')
                ) {
                  return false;
                }
                if (button.closest('nav, header')) return false;
                if (button.closest('details:not([open])')) return false;
                if (cardButtonSelectors.some((selector) => button.matches(selector))) return false;
                if ([...button.querySelectorAll('img[alt]')].some((image) => cardImageAlt.test(image.alt.trim()))) {
                  return false;
                }
                const label = `${button.innerText} ${button.getAttribute('aria-label') ?? ''}`
                  .trim()
                  .replace(/\s+/g, ' ');
                return !auxiliaryLabels.some((excluded) => label.includes(excluded));
              },
            );

            const bounds = (element: HTMLElement) => {
              const box = element.getBoundingClientRect();
              let top = 0;
              let right = window.innerWidth;
              let bottom = window.innerHeight;
              let left = 0;
              for (let ancestor = element.parentElement; ancestor; ancestor = ancestor.parentElement) {
                const style = getComputedStyle(ancestor);
                if (/^(auto|scroll|hidden)$/.test(style.overflowY) && ancestor.scrollHeight > ancestor.clientHeight) {
                  const rect = ancestor.getBoundingClientRect();
                  top = Math.max(top, rect.top);
                  right = Math.min(right, rect.right);
                  bottom = Math.min(bottom, rect.bottom);
                  left = Math.max(left, rect.left);
                }
              }
              return { box, visible: box.top >= top && box.right <= right && box.bottom <= bottom && box.left >= left };
            };

            const measured = buttons.map((button) => {
              const { box, visible } = bounds(button);
              return {
                label: (
                  button.innerText ||
                  button.getAttribute('aria-label') ||
                  button.getAttribute('title') ||
                  '(文言なし)'
                )
                  .trim()
                  .replace(/\s+/g, ' '),
                y: Math.round(box.top),
                visible,
              };
            });
            return {
              guidance: true,
              outside: measured.filter((button) => !button.visible).map(({ label, y }) => ({ label, y })),
              viewportBottom: window.innerHeight,
            };
          },
          {
            selector: guidanceSelector,
            auxiliaryLabels: AUXILIARY_ACTION_LABELS,
            cardButtonSelectors: CARD_BUTTON_SELECTORS,
            cardImageAltSource: CARD_IMAGE_ALT_PATTERN.source,
            cardImageAltFlags: CARD_IMAGE_ALT_PATTERN.flags,
          },
        );

        if (!result.guidance) {
          test.info().annotations.push({
            type: 'skipped page',
            description: 'No non-empty visible GameMessageBox guidance was present.',
          });
          return;
        }
        expect(
          result.outside,
          `${path}: 表示域外の有効な操作ボタン: ${JSON.stringify(result.outside)}。表示域の下端 y=${result.viewportBottom}`,
        ).toEqual([]);
      });
    }
  });
}
