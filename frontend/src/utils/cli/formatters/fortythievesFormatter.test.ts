import { describe, expect, it } from 'vitest';
import type { Card, FortyThievesResponse, FortyThievesTableauCard } from '../../../types/card';
import { FortyThievesPhase } from '../../../types/phases';
import { formatFortythievesState } from './fortythievesFormatter';

const card = (design: Card['design'], value: number): Card => ({ design, value });
const up = (c: Card): FortyThievesTableauCard => ({ card: c, faceUp: true });
const down = (): FortyThievesTableauCard => ({ card: null, faceUp: false });

const baseState = (overrides: Partial<FortyThievesResponse> = {}): FortyThievesResponse => ({
  tableau: [[up(card('SPADE', 5)), up(card('HEART', 4))], [down()], []],
  stockCount: 30,
  waste: [card('CLOVER', 9)],
  foundation: [[], [card('DIAMOND', 1)]],
  phase: FortyThievesPhase.PLAYING,
  moveCount: 7,
  canUndo: true,
  isStalemate: false,
  message: '',
  ...overrides,
});

describe('formatFortythievesState', () => {
  it('renders foundations, stock, waste, and tableau columns', () => {
    const out = formatFortythievesState(baseState());
    expect(out).toContain('Forty Thieves');
    expect(out).toContain('foundation: [  ] | ♦A');
    expect(out).toContain('stock: 30  waste: ♣9');
    expect(out).toContain('t0: [0]♠5 [1]♥4');
    expect(out).toContain('t1: [?]');
    expect(out).toContain('t2: [empty]');
    expect(out).toContain('moves: 7  undo:yes');
  });

  it('renders an empty waste placeholder', () => {
    const out = formatFortythievesState(baseState({ waste: [] }));
    expect(out).toContain('waste: [  ]');
  });

  it('renders a tableau hint with source index and target', () => {
    const out = formatFortythievesState(
      baseState({
        hint: { fromZone: 'tableau', fromCol: 0, cardIndex: 1, toZone: 'foundation', toCol: 2 },
        messageCode: 'fortythieves.hintAvailable',
      }),
    );
    expect(out).toContain('HINT: t0[1] → foundation2');
  });

  // #5525: 引くヒントは列を持たない。移動の体裁に落とすと t-1[-1] が出る。
  it('renders the draw hint without leaking the -1 columns', () => {
    const out = formatFortythievesState(
      baseState({
        hint: { fromZone: 'stock', fromCol: -1, cardIndex: -1, toZone: 'waste', toCol: -1 },
        messageCode: 'fortythieves.hintAvailable',
      }),
    );
    expect(out).toContain('HINT: draw from stock');
    expect(out).not.toContain('-1');
  });

  it('renders a waste hint source', () => {
    const out = formatFortythievesState(
      baseState({
        hint: { fromZone: 'waste', fromCol: -1, cardIndex: -1, toZone: 'tableau', toCol: 4 },
        messageCode: 'fortythieves.hintAvailable',
      }),
    );
    expect(out).toContain('HINT: waste → tableau4');
  });

  it('renders stalemate, message, and win lines', () => {
    const out = formatFortythievesState(
      baseState({ isStalemate: true, message: 'stuck', phase: FortyThievesPhase.GAME_CLEAR }),
    );
    expect(out).toContain('Stalemate - no more moves possible');
    expect(out).toContain('stuck');
    expect(out).toContain('Congratulations! You win!');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { fromZone: 'tableau', fromCol: 0, cardIndex: 1, toZone: 'foundation', toCol: 2 };
    expect(formatFortythievesState(baseState({ hint, messageCode: 'fortythieves.hintAvailable' }))).toContain('HINT:');
    expect(formatFortythievesState(baseState({ hint, messageCode: 'fortythieves.playing' }))).not.toContain('HINT:');
  });
});
