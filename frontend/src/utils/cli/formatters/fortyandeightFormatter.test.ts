import { describe, expect, it } from 'vitest';
import type { Card, FortyAndEightResponse, FortyAndEightTableauCard } from '../../../types/card';
import { FortyAndEightPhase } from '../../../types/phases';
import { formatFortyandeightState } from './fortyandeightFormatter';

const card = (design: Card['design'], value: number): Card => ({ design, value });
const up = (c: Card): FortyAndEightTableauCard => ({ card: c, faceUp: true });
const down = (): FortyAndEightTableauCard => ({ card: null, faceUp: false });

const baseState = (overrides: Partial<FortyAndEightResponse> = {}): FortyAndEightResponse => ({
  tableau: [[up(card('SPADE', 5)), up(card('HEART', 4))], [down()], []],
  stockCount: 30,
  waste: [card('CLOVER', 9)],
  foundation: [[], [card('DIAMOND', 1)]],
  redealUsed: false,
  canRedeal: false,
  phase: FortyAndEightPhase.PLAYING,
  moveCount: 7,
  canUndo: true,
  isStalemate: false,
  message: '',
  ...overrides,
});

describe('formatFortyandeightState', () => {
  it('renders foundations, stock, waste, redeal, and tableau columns', () => {
    const out = formatFortyandeightState(baseState());
    expect(out).toContain('Forty and Eight');
    expect(out).toContain('foundation: [  ] | ♦A');
    expect(out).toContain('stock: 30  waste: ♣9  redeal:available');
    expect(out).toContain('t0: [0]♠5 [1]♥4');
    expect(out).toContain('t1: [?]');
    expect(out).toContain('t2: [empty]');
    expect(out).toContain('moves: 7  undo:yes');
  });

  it('renders an empty waste placeholder', () => {
    const out = formatFortyandeightState(baseState({ waste: [] }));
    expect(out).toContain('waste: [  ]');
  });

  it('renders redeal used state', () => {
    const out = formatFortyandeightState(baseState({ redealUsed: true }));
    expect(out).toContain('redeal:used');
  });

  it('renders a tableau hint with source index and target', () => {
    const out = formatFortyandeightState(
      baseState({
        hint: { fromZone: 'tableau', fromCol: 0, cardIndex: 1, toZone: 'foundation', toCol: 2 },
        messageCode: 'fortyandeight.hintAvailable',
      }),
    );
    expect(out).toContain('HINT: t0[1] → foundation2');
  });

  it('renders a waste hint source', () => {
    const out = formatFortyandeightState(
      baseState({
        hint: { fromZone: 'waste', fromCol: -1, cardIndex: -1, toZone: 'tableau', toCol: 4 },
        messageCode: 'fortyandeight.hintAvailable',
      }),
    );
    expect(out).toContain('HINT: waste → tableau4');
  });

  it('renders stalemate, message, and win lines', () => {
    const out = formatFortyandeightState(
      baseState({ isStalemate: true, message: 'stuck', phase: FortyAndEightPhase.GAME_CLEAR }),
    );
    expect(out).toContain('Stalemate - no more moves possible');
    expect(out).toContain('stuck');
    expect(out).toContain('Congratulations! You win!');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { fromZone: 'tableau', fromCol: 0, cardIndex: 1, toZone: 'foundation', toCol: 2 };
    expect(formatFortyandeightState(baseState({ hint, messageCode: 'fortyandeight.hintAvailable' }))).toContain(
      'HINT:',
    );
    expect(formatFortyandeightState(baseState({ hint, messageCode: 'fortyandeight.playing' }))).not.toContain('HINT:');
  });
});
