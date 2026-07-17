import { describe, expect, it } from 'vitest';
import type { ContractRummyPlayer, ContractRummyResponse } from '../../../types/card';
import { formatContractRummyState } from './contractrummyFormatter';

const player = (over: Partial<ContractRummyPlayer> = {}): ContractRummyPlayer => ({
  id: 0,
  isHuman: true,
  cardCount: 2,
  cards: [
    { design: 'SPADE', value: 5 },
    { design: 'HEART', value: 13 },
  ],
  melds: [],
  contractMet: false,
  roundScore: 0,
  cumulativeScore: 0,
  ...over,
});

const baseState: ContractRummyResponse = {
  message: '',
  players: [player(), player({ id: 1, isHuman: false, cardCount: 4, cards: [] })],
  phase: 1,
  roundNumber: 1,
  totalRounds: 7,
  currentPlayerIdx: 0,
  discardTop: { design: 'CLOVER', value: 7 },
  drawPileCount: 30,
  gameEndFlag: false,
  winnerIdx: -1,
  roundWinnerIdx: -1,
  contractSlots: [
    { kind: 0, size: 3 },
    { kind: 0, size: 3 },
  ],
  config: { cpuDifficulty: 1, failContractPenalty: 25 },
};

describe('formatContractRummyState', () => {
  it('returns a loading placeholder for null state', () => {
    expect(formatContractRummyState(null)).toBe('Loading...');
  });

  it('renders round, phase, contract, stock, and discard', () => {
    const out = formatContractRummyState(baseState);
    expect(out).toContain('Contract Rummy');
    expect(out).toContain('round 1/7');
    expect(out).toContain('phase: PLAY');
    expect(out).toContain('contract: set(3) + set(3)');
    expect(out).toContain('stock: 30');
    expect(out).toContain('discard: ♣7');
  });

  it('shows an em dash when the discard pile is empty', () => {
    const out = formatContractRummyState({ ...baseState, discardTop: null });
    expect(out).toContain('discard: —');
  });

  it('marks the current player and renders melds', () => {
    const out = formatContractRummyState({
      ...baseState,
      players: [
        player({
          contractMet: true,
          melds: [
            {
              cards: [
                { design: 'SPADE', value: 3 },
                { design: 'HEART', value: 3 },
                { design: 'DIAMOND', value: 3 },
              ],
            },
          ],
        }),
        player({ id: 1, isHuman: false, cardCount: 4, cards: [] }),
      ],
    });
    expect(out).toContain('>');
    expect(out).toContain('[contract met]');
    expect(out).toContain('M0: ♠3 ♥3 ♦3');
  });

  it('renders the indexed human hand', () => {
    const out = formatContractRummyState(baseState);
    expect(out).toContain('[0]♠5');
    expect(out).toContain('[1]♥K');
  });

  it('announces the game winner', () => {
    const out = formatContractRummyState({ ...baseState, gameEndFlag: true, winnerIdx: 0 });
    expect(out).toContain('winner:');
  });

  it('announces the round winner at round end', () => {
    const out = formatContractRummyState({ ...baseState, phase: 2, roundWinnerIdx: 1 });
    expect(out).toContain('round winner:');
  });

  it('appends a server message when present', () => {
    const out = formatContractRummyState({ ...baseState, message: 'Draw a card' });
    expect(out).toContain('Draw a card');
  });
});
