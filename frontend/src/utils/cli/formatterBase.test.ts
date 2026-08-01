import { describe, expect, it } from 'vitest';
import type { Card } from '../../types/card';
import {
  formatCard,
  formatCardList,
  formatHeader,
  formatIndexedCards,
  formatPlayerName,
  formatSeparator,
  isRequestedHint,
} from './formatterBase';

const S = '\u2660'; // spade
const H = '\u2665'; // heart
const D = '\u2666'; // diamond
const C = '\u2663'; // clover

describe('formatCard', () => {
  it('formats spade card', () => {
    expect(formatCard({ design: 'SPADE', value: 5 })).toBe(`${S}5`);
  });

  it('formats heart card', () => {
    expect(formatCard({ design: 'HEART', value: 10 })).toBe(`${H}10`);
  });

  it('formats diamond ace', () => {
    expect(formatCard({ design: 'DIAMOND', value: 1 })).toBe(`${D}A`);
  });

  it('formats clover king', () => {
    expect(formatCard({ design: 'CLOVER', value: 13 })).toBe(`${C}K`);
  });

  it('formats joker', () => {
    expect(formatCard({ design: 'JOKER', value: 0 })).toBe('\uD83C\uDCCF');
  });

  it('formats jack', () => {
    expect(formatCard({ design: 'SPADE', value: 11 })).toBe(`${S}J`);
  });

  it('formats queen', () => {
    expect(formatCard({ design: 'HEART', value: 12 })).toBe(`${H}Q`);
  });

  it('formats a procedural card from its glyph and label', () => {
    expect(formatCard({ design: 'JOKER', value: 1, glyph: '✦', label: 'Wizard', deck: 'wizard' })).toBe('✦Wizard');
  });

  it('formats a procedural card from label alone when glyph is absent', () => {
    expect(formatCard({ design: 'JOKER', value: 1, label: 'Jester', deck: 'wizard' })).toBe('Jester');
  });
});

describe('formatCardList', () => {
  it('formats multiple cards comma-separated', () => {
    const cards: Card[] = [
      { design: 'SPADE', value: 5 },
      { design: 'HEART', value: 10 },
    ];
    expect(formatCardList(cards)).toBe(`${S}5, ${H}10`);
  });

  it('returns empty string for empty array', () => {
    expect(formatCardList([])).toBe('');
  });
});

describe('formatIndexedCards', () => {
  it('formats cards with indices', () => {
    const cards: Card[] = [
      { design: 'SPADE', value: 5 },
      { design: 'HEART', value: 13 },
    ];
    expect(formatIndexedCards(cards)).toBe(`[0]${S}5  [1]${H}K`);
  });

  it('returns empty string for empty array', () => {
    expect(formatIndexedCards([])).toBe('');
  });
});

describe('formatSeparator', () => {
  it('returns separator line', () => {
    expect(formatSeparator()).toBe('==========');
  });
});

describe('formatHeader', () => {
  it('wraps title in separators', () => {
    expect(formatHeader('BlackJack')).toBe('==========\nBlackJack\n==========');
  });
});

describe('formatPlayerName', () => {
  it('returns you label for human', () => {
    expect(formatPlayerName(0, true)).toBe('\u3042\u306a\u305f');
  });

  it('returns CPU label for non-human', () => {
    expect(formatPlayerName(1, false)).toBe('CPU 1');
  });

  it('returns CPU label with given index', () => {
    expect(formatPlayerName(3, false)).toBe('CPU 3');
  });
});

describe('isRequestedHint', () => {
  // **頼んでいないヒントは CLI に出さない。**#4483 以降 Output() もヒントを載せる
  // ので、state.hint だけを見ると毎手 HINT が印字される。要求への応答だけが
  // hintAvailable を持つので、そこで区別する。
  it('is true only for a hint command response', () => {
    expect(isRequestedHint({ messageCode: 'braid.hintAvailable' })).toBe(true);
    expect(isRequestedHint({ messageCode: 'braid.playing' })).toBe(false);
    expect(isRequestedHint({ messageCode: 'braid.noHint' })).toBe(false);
  });

  // messageCode が無いレスポンスもある。undefined で落ちないこと。
  it('is false when the response carries no message code', () => {
    expect(isRequestedHint({})).toBe(false);
    expect(isRequestedHint({ messageCode: undefined })).toBe(false);
  });

  // 別ゲームの接尾辞でも同じ規則で効くこと。
  it('works for any game prefix', () => {
    expect(isRequestedHint({ messageCode: 'terrace.hintAvailable' })).toBe(true);
    expect(isRequestedHint({ messageCode: 'americantoad.hintAvailable' })).toBe(true);
  });

  // **hintRequested も「頼んだヒント」。**トリックテイキング系は
  // hintAvailable がラベルとして埋まっているため別キーを使う (#4483)。
  it('accepts hintRequested as well as hintAvailable', () => {
    expect(isRequestedHint({ messageCode: 'sedma.hintRequested' })).toBe(true);
    expect(isRequestedHint({ messageCode: 'klondike.hintAvailable' })).toBe(true);
    expect(isRequestedHint({ messageCode: 'sedma.playing' })).toBe(false);
    expect(isRequestedHint({})).toBe(false);
  });
});
