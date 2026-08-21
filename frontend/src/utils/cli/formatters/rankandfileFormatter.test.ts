import { describe, expect, it } from 'vitest';
import type { Card, RankAndFileResponse, RankAndFileTableauCard } from '../../../types/card';
import { RankAndFilePhase } from '../../../types/phases';
import { formatRankandfileState } from './rankandfileFormatter';

const card = (design: Card['design'], value: number): Card => ({ design, value });
const up = (c: Card): RankAndFileTableauCard => ({ card: c, faceUp: true });
const down = (): RankAndFileTableauCard => ({ card: null, faceUp: false });

const baseState = (overrides: Partial<RankAndFileResponse> = {}): RankAndFileResponse => ({
  tableau: [[up(card('SPADE', 5)), up(card('HEART', 4))], [down()], []],
  stockCount: 30,
  waste: [card('CLOVER', 9)],
  foundation: [[], [card('DIAMOND', 1)]],
  phase: RankAndFilePhase.PLAYING,
  moveCount: 7,
  canUndo: true,
  isStalemate: false,
  message: '',
  ...overrides,
});

describe('formatRankandfileState', () => {
  it('renders foundations, stock, waste, and tableau columns', () => {
    const out = formatRankandfileState(baseState());
    expect(out).toContain('Rank and File');
    expect(out).toContain('foundation: [  ] | ♦A');
    expect(out).toContain('stock: 30  waste: ♣9');
    expect(out).toContain('t0: [0]♠5 [1]♥4');
    expect(out).toContain('t1: [?]');
    expect(out).toContain('t2: [empty]');
    expect(out).toContain('moves: 7  undo:yes');
  });

  it('renders an empty waste placeholder', () => {
    const out = formatRankandfileState(baseState({ waste: [] }));
    expect(out).toContain('waste: [  ]');
  });

  it('renders a tableau hint with source index and target', () => {
    const out = formatRankandfileState(
      baseState({
        hint: { fromZone: 'tableau', fromCol: 0, cardIndex: 1, toZone: 'foundation', toCol: 2 },
        messageCode: 'rankandfile.hintAvailable',
      }),
    );
    expect(out).toContain('HINT: t0[1] → foundation2');
  });

  // #5525: 引くヒントは列を持たない。移動の体裁に落とすと t-1[-1] が出る。
  it('renders the draw hint without leaking the -1 columns', () => {
    const out = formatRankandfileState(
      baseState({
        hint: { fromZone: 'stock', fromCol: -1, cardIndex: -1, toZone: 'waste', toCol: -1 },
        messageCode: 'rankandfile.hintAvailable',
      }),
    );
    expect(out).toContain('HINT: draw from stock');
    expect(out).not.toContain('-1');
  });

  it('renders a waste hint source', () => {
    const out = formatRankandfileState(
      baseState({
        hint: { fromZone: 'waste', fromCol: -1, cardIndex: -1, toZone: 'tableau', toCol: 4 },
        messageCode: 'rankandfile.hintAvailable',
      }),
    );
    expect(out).toContain('HINT: waste → tableau4');
  });

  it('renders stalemate, message, and win lines', () => {
    const out = formatRankandfileState(
      baseState({ isStalemate: true, message: 'stuck', phase: RankAndFilePhase.GAME_CLEAR }),
    );
    expect(out).toContain('Stalemate - no more moves possible');
    expect(out).toContain('stuck');
    expect(out).toContain('Congratulations! You win!');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { fromZone: 'tableau', fromCol: 0, cardIndex: 1, toZone: 'foundation', toCol: 2 };
    expect(formatRankandfileState(baseState({ hint, messageCode: 'rankandfile.hintAvailable' }))).toContain('HINT:');
    expect(formatRankandfileState(baseState({ hint, messageCode: 'rankandfile.playing' }))).not.toContain('HINT:');
  });
});
