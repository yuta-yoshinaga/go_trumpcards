import { describe, expect, it } from 'vitest';
import { BOARD_CARD_SELECTOR } from './boardCardSelector';

describe('BOARD_CARD_SELECTOR', () => {
  it('excludes role=img containers with card images and keeps standalone procedural cards', () => {
    document.body.innerHTML = `
      <div role="img" aria-label="retired queens">
        <img alt="queen" />
        <img alt="queen" />
      </div>
      <div role="img" aria-label="tarot card"></div>
      <img alt="playing card" />
    `;

    const matches = [...document.querySelectorAll(BOARD_CARD_SELECTOR)];

    expect(matches).toHaveLength(4);
    expect(matches.map((element) => element.getAttribute('aria-label'))).toEqual([null, null, 'tarot card', null]);
  });
});
