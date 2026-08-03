import { describe, expect, it } from 'vitest';
import type { AmericanToadResponse } from '../../../types/card';
import { formatAmericanToadState } from './americantoadFormatter';

function makeState(overrides?: Partial<AmericanToadResponse>): AmericanToadResponse {
  return {
    reserve: [],
    tableau: Array.from({ length: 8 }, () => []),
    foundation: Array.from({ length: 8 }, () => []),
    stockCount: 75,
    waste: [],
    baseRank: 5,
    passesUsed: 0,
    canRedeal: false,
    phase: 0,
    moveCount: 0,
    canUndo: false,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('formatAmericanToadState', () => {
  it('renders header, piles and an empty board', () => {
    const result = formatAmericanToadState(makeState());
    expect(result).toContain('American Toad');
    expect(result).toContain('base rank: 5');
    expect(result).toContain('foundations:');
    expect(result).toContain('reserve: [  ] (0)');
    expect(result).toContain('stock: 75');
    expect(result).toContain('t0: [empty]');
    expect(result).toContain('t7: [empty]');
  });

  // The reserve's depth decides how empty columns behave, so it is status.
  it('renders the reserve top with its depth', () => {
    const result = formatAmericanToadState(
      makeState({
        reserve: [
          { design: 'SPADE', value: 3 },
          { design: 'HEART', value: 9 },
        ],
      }),
    );
    expect(result).toContain('(2)');
  });

  it('renders the waste top', () => {
    expect(formatAmericanToadState(makeState({ waste: [{ design: 'DIAMOND', value: 4 }] }))).toContain('waste:');
  });

  it('renders cards in the tableau', () => {
    const result = formatAmericanToadState(
      makeState({
        tableau: [[{ card: { design: 'SPADE', value: 9 }, faceUp: true }], ...Array.from({ length: 7 }, () => [])],
      }),
    );
    expect(result).toContain('[0]');
  });

  it('renders a placeholder for a missing card', () => {
    const result = formatAmericanToadState(
      makeState({ tableau: [[{ card: null, faceUp: true }], ...Array.from({ length: 7 }, () => [])] }),
    );
    expect(result).toContain('[?]');
  });

  it('announces the single redeal', () => {
    expect(formatAmericanToadState(makeState({ stockCount: 0, canRedeal: true }))).toContain('once only');
  });

  it('shows a tableau hint with the run head', () => {
    const result = formatAmericanToadState(
      makeState({
        hint: { fromZone: 'tableau', fromIdx: 3, cardIndex: 1, toZone: 'tableau', toIdx: 7 },
        messageCode: 'americantoad.hintAvailable',
      }),
    );
    expect(result).toContain('HINT');
    expect(result).toContain('t3[1]');
  });

  // **頼んでいないヒントは CLI に出さない。**
  it('does not print a passive hint carried on an ordinary response', () => {
    const result = formatAmericanToadState(
      makeState({
        hint: { fromZone: 'tableau', fromIdx: 3, cardIndex: 1, toZone: 'tableau', toIdx: 7 },
        messageCode: 'americantoad.playing',
      }),
    );
    expect(result).not.toContain('HINT:');
  });

  // A draw has no destination index.
  it('renders a draw hint without a destination index', () => {
    const result = formatAmericanToadState(
      makeState({
        hint: { fromZone: 'stock', fromIdx: -1, cardIndex: -1, toZone: 'waste', toIdx: -1 },
        messageCode: 'americantoad.hintAvailable',
      }),
    );
    expect(result).toContain('stock → waste');
  });

  it('shows stalemate message', () => {
    expect(formatAmericanToadState(makeState({ isStalemate: true }))).toContain('Stalemate');
  });

  it('shows the server message', () => {
    expect(formatAmericanToadState(makeState({ message: 'nope' }))).toContain('nope');
  });

  it('shows congrats on win phase', () => {
    expect(formatAmericanToadState(makeState({ phase: 1 }))).toContain('Congratulations');
  });
});
