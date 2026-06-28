import { describe, expect, it } from 'vitest';
import type { Card, CardDesign, OpenFaceChinesePlayer, OpenFaceChineseResponse } from '../../../types/card';
import { OpenFaceChinesePhase } from '../../../types/phases';
import { formatOpenfacechineseState } from './openfacechineseFormatter';

const card = (design: CardDesign, value: number): Card => ({ design, value });

const player = (overrides: Partial<OpenFaceChinesePlayer> = {}): OpenFaceChinesePlayer => ({
  id: 0,
  isHuman: true,
  front: [],
  middle: [],
  back: [],
  pending: [],
  roundScore: 0,
  royalty: 0,
  fouled: false,
  fantasyland: false,
  totalScore: 0,
  ...overrides,
});

const baseState: OpenFaceChineseResponse = {
  phase: OpenFaceChinesePhase.PLACING,
  roundNumber: 1,
  currentPlayerIdx: 0,
  dealerIdx: 1,
  currentCard: card('SPADE', 14),
  gameEndFlag: false,
  winnerIdx: -1,
  isHumanTurn: true,
  players: [player({ id: 0, isHuman: true, front: [card('HEART', 5)] }), player({ id: 1, isHuman: false })],
  config: { cpuDifficulty: 1, playerCount: 2, targetRounds: 1 },
  message: '',
};

describe('formatOpenfacechineseState', () => {
  it('renders the header, round, and phase', () => {
    const out = formatOpenfacechineseState(baseState);
    expect(out).toContain('Open Face Chinese Poker');
    expect(out).toContain('round: 1');
    expect(out).toContain('phase: PLACING');
  });

  it('prompts to place the pending card on the human turn', () => {
    const out = formatOpenfacechineseState(baseState);
    expect(out).toContain('Place this card');
    expect(out).toContain('You');
    expect(out).toContain('CPU 1');
  });

  it('hides round scores during placing', () => {
    const out = formatOpenfacechineseState(baseState);
    expect(out).not.toContain('total:');
  });

  it('shows scores, royalty, foul, and fantasyland at round end', () => {
    const out = formatOpenfacechineseState({
      ...baseState,
      phase: OpenFaceChinesePhase.ROUND_END,
      isHumanTurn: false,
      currentCard: undefined,
      players: [
        player({ id: 0, isHuman: true, roundScore: 6, royalty: 4, totalScore: 6, fantasyland: true }),
        player({ id: 1, isHuman: false, roundScore: -6, totalScore: -6, fouled: true }),
      ],
    });
    expect(out).toContain('total: 6');
    expect(out).toContain('royalty +4');
    expect(out).toContain('FANTASYLAND');
    expect(out).toContain('FOULED');
    expect(out).not.toContain('Place this card');
  });

  it('renders UNKNOWN for an unexpected phase', () => {
    expect(formatOpenfacechineseState({ ...baseState, phase: 99 })).toContain('phase: UNKNOWN');
  });
});
