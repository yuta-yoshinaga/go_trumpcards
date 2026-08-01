import { describe, expect, it } from 'vitest';
import type { NertzResponse } from '../../../types/card';
import { formatNertzState } from './nertzFormatter';

function makeState(overrides?: Partial<NertzResponse>): NertzResponse {
  return {
    phase: 1,
    roundNumber: 1,
    winnerIdx: -1,
    matchWinner: -1,
    moveCount: 0,
    canUndo: false,
    playerCount: 2,
    drawCount: 3,
    targetScore: 100,
    cpuDifficulty: 0,
    cpuTickMoves: 0,
    players: [
      {
        name: 'You',
        isHuman: true,
        deckIdx: 0,
        score: 0,
        nertzSize: 13,
        tableau: [[], [], [], []],
        wasteSize: 0,
        stockSize: 31,
      },
      {
        name: 'CPU 1',
        isHuman: false,
        deckIdx: 1,
        score: 0,
        nertzSize: 13,
        tableau: [[], [], [], []],
        wasteSize: 0,
        stockSize: 31,
      },
    ],
    foundations: [],
    message: '',
    ...overrides,
  };
}

describe('formatNertzState', () => {
  it('renders header and phase', () => {
    const result = formatNertzState(makeState());
    expect(result).toContain('Nertz');
    expect(result).toContain('PLAYING');
    expect(result).toContain('target: 100');
  });

  it('shows player tableau for humans only', () => {
    const result = formatNertzState(
      makeState({
        players: [
          {
            name: 'You',
            isHuman: true,
            deckIdx: 0,
            score: 5,
            nertzSize: 12,
            nertzTop: { design: 'SPADE', value: 7 },
            tableau: [[{ card: { design: 'HEART', value: 3 }, faceUp: true } as never], [], [], []],
            wasteSize: 0,
            stockSize: 30,
          },
        ],
      }),
    );
    expect(result).toContain('[YOU]');
    expect(result).toContain('t0:');
    expect(result).toContain('[0]');
  });

  it('shows hint when present', () => {
    const result = formatNertzState(
      makeState({
        hint: { fromZone: 'tableau', fromCol: 0, cardIndex: 1, toZone: 'foundation', toCol: 2 },
        messageCode: 'nertz.hintAvailable',
      }),
    );
    expect(result).toContain('HINT');
  });

  it('shows match winner', () => {
    const result = formatNertzState(makeState({ matchWinner: 0 }));
    expect(result).toContain('Match Winner');
  });

  it('shows round winner', () => {
    const result = formatNertzState(makeState({ winnerIdx: 1 }));
    expect(result).toContain('Round Winner');
  });

  it('renders foundation top cards', () => {
    const result = formatNertzState(
      makeState({
        foundations: [{ top: { design: 'SPADE', value: 5 }, suit: 1, size: 5 }],
      }),
    );
    expect(result).toContain('foundation:');
  });

  it('renders empty foundations as placeholders', () => {
    const result = formatNertzState(makeState({ foundations: [{ suit: 1, size: 0 }] }));
    expect(result).toContain('foundation: [  ]');
  });

  it('renders human waste/nertz top placeholders when empty', () => {
    const result = formatNertzState(
      makeState({
        players: [
          {
            name: 'You',
            isHuman: true,
            deckIdx: 0,
            score: 0,
            nertzSize: 0,
            tableau: [[], [], [], []],
            wasteSize: 0,
            stockSize: 0,
          },
        ],
      }),
    );
    expect(result).toContain('waste: [  ]');
    expect(result).toContain('nertz: [  ]');
  });

  it('renders human waste/nertz top cards when present', () => {
    const result = formatNertzState(
      makeState({
        players: [
          {
            name: 'You',
            isHuman: true,
            deckIdx: 0,
            score: 0,
            nertzSize: 5,
            nertzTop: { design: 'CLOVER', value: 9 },
            tableau: [[], [], [], []],
            wasteTop: { design: 'DIAMOND', value: 3 },
            wasteSize: 1,
            stockSize: 25,
          },
        ],
      }),
    );
    expect(result).toContain('waste:');
    expect(result).toContain('nertz:');
  });

  it('renders empty tableau columns', () => {
    const result = formatNertzState(
      makeState({
        players: [
          {
            name: 'You',
            isHuman: true,
            deckIdx: 0,
            score: 0,
            nertzSize: 0,
            tableau: [[], [], [], []],
            wasteSize: 0,
            stockSize: 0,
          },
        ],
      }),
    );
    expect(result).toContain('t0: [empty]');
  });

  it('renders unknown phase as UNKNOWN', () => {
    const result = formatNertzState(makeState({ phase: 99 }));
    expect(result).toContain('UNKNOWN');
  });

  it('renders message when present', () => {
    const result = formatNertzState(makeState({ message: 'You drew a card' }));
    expect(result).toContain('You drew a card');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { fromZone: 'tableau', fromCol: 0, cardIndex: 1, toZone: 'foundation', toCol: 2 };
    expect(formatNertzState(makeState({ hint, messageCode: 'nertz.hintAvailable' }))).toContain('HINT');
    expect(formatNertzState(makeState({ hint, messageCode: 'nertz.playing' }))).not.toContain('HINT');
  });
});
