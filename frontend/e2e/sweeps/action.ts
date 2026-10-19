import { expect, type Page, test } from '@playwright/test';
import { AUXILIARY_ACTION_LABELS } from '../auxiliaryActions';

const guidanceSelector =
  '[role="status"].glass-panel, [role="alert"].glass-panel, [role="status"] .glass-panel, [role="alert"] .glass-panel';

// Cards stay visible in the hand and are not phase actions; exclude every card marker used by game pages.
const CARD_BUTTON_SELECTORS = [
  '[data-hand-card-index]',
  '[aria-pressed]',
  '[data-card]',
  '[data-card-index]',
  '[data-testid*="card"]',
] as const;
const CARD_IMAGE_ALT_PATTERN = /^(?:[♠♥♦♣]\s*)?(?:[2-9]|10|[AJQK])(?:\s*[♠♥♦♣])?$/i;

/** Keeps every non-card, non-auxiliary action inside the desktop viewport. */
export async function checkAction(
  page: Page,
  path: string,
  _viewport: { width: number; height: number },
): Promise<void> {
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

      const buttons = [...document.querySelectorAll<HTMLElement>('main button, footer button')].filter((button) => {
        // #11825: aria-disabled primary actions remain visible because project policy (#8524/#8787) keeps the next action available in the UI until it becomes operable.
        if (button.matches(':disabled')) return false;
        const style = getComputedStyle(button);
        if (style.display === 'none' || style.visibility === 'hidden' || button.closest('[aria-hidden="true"]')) {
          return false;
        }
        if (button.closest('nav, header')) return false;
        if (button.closest('details:not([open])')) return false;
        if (cardButtonSelectors.some((selector) => button.matches(selector))) return false;
        if ([...button.querySelectorAll('img[alt]')].some((image) => cardImageAlt.test(image.alt.trim()))) {
          return false;
        }
        const label = `${button.innerText} ${button.getAttribute('aria-label') ?? ''}`.trim().replace(/\s+/g, ' ');
        return !auxiliaryLabels.some((excluded) => label.includes(excluded));
      });

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
          label: (button.innerText || button.getAttribute('aria-label') || button.getAttribute('title') || '(文言なし)')
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
      description: `Action check: ${path}: No non-empty visible GameMessageBox guidance was present.`,
    });
    return;
  }
  expect
    .soft(
      result.outside,
      `Action check: ${path}: 表示域外の有効な操作ボタン: ${JSON.stringify(result.outside)}。表示域の下端 y=${result.viewportBottom}`,
    )
    .toEqual([]);
}
