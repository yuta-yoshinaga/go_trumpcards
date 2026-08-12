import { describe, expect, it } from 'vitest';
import type { Card, ColourWhistResponse } from '../../../types/card';
import { COLOUR_WHIST_NO_TRUMP } from '../../../types/games/colourwhist';
import { ColourWhistContract, ColourWhistPhase } from '../../../types/phases';
import { formatColourWhistState } from './colourwhistFormatter';

const cards: Card[] = [
  { design: 'SPADE', value: 1 },
  { design: 'HEART', value: 9 },
  { design: 'CLOVER', value: 4 },
];

const base: ColourWhistResponse = {
  players: [
    { id: 0, isHuman: true, cardCount: 3, cards, trickCount: 2, score: 6, isDeclarerSide: true, hasPassed: false },
    {
      id: 1,
      isHuman: false,
      cardCount: 3,
      cards: [],
      trickCount: 1,
      score: -2,
      isDeclarerSide: false,
      hasPassed: true,
    },
  ],
  phase: ColourWhistPhase.PLAY,
  validPlays: [2],
  dealerIdx: 0,
  contract: ColourWhistContract.ALLEEN,
  declarerIdx: 0,
  partnerIdx: -1,
  trumpSuit: 3,
  troelForced: false,
  currentTurn: 0,
  isHumanTurn: true,
  currentTrick: [{ playerIdx: 1, card: { design: 'CLOVER', value: 13 } }],
  lastTrick: [],
  lastTrickWinner: -1,
  trickCount: 3,
  declarerTricks: 2,
  roundNumber: 2,
  gameEndFlag: false,
  winnerIdx: -1,
  config: { rounds: 8 },
  message: '',
};

describe('formatColourWhistState', () => {
  it('shows the phase, round, contract and trump', () => {
    const out = formatColourWhistState(base);
    expect(out).toContain('phase: PLAY');
    expect(out).toContain('round: 2 / 8');
    expect(out).toContain('contract: Alleen (trump: Hearts)');
    expect(out).toContain('declarer: seat 0 (2 tricks)');
  });

  // **競りが飛ばされた理由を出す。** 出さないと不具合に見えます。
  it('explains a troel deal', () => {
    const out = formatColourWhistState({
      ...base,
      troelForced: true,
      contract: ColourWhistContract.TROEL,
      declarerIdx: 1,
    });
    expect(out).toContain('contract: Troel');
    expect(out).toContain('seat 1 held three aces');
    expect(out).toContain('no auction');
  });

  // **競りで決めた契約では出さない。** 負のコントロールです。
  it('says nothing about troel for a bid contract', () => {
    expect(formatColourWhistState(base)).not.toContain('no auction');
  });

  it('names no trump rather than printing -1', () => {
    const out = formatColourWhistState({
      ...base,
      contract: ColourWhistContract.MISERIE,
      trumpSuit: COLOUR_WHIST_NO_TRUMP,
    });
    expect(out).toContain('contract: Miserie (trump: none)');
    expect(out).not.toContain('-1');
  });

  // **得点は負も出す。**
  it('shows negative scores', () => {
    expect(formatColourWhistState(base)).toContain('#1:-2');
  });

  it('marks only the legal cards in hand', () => {
    const out = formatColourWhistState(base);
    expect(out).toContain('[0 ]');
    expect(out).toContain('[1 ]');
    expect(out).toContain('[2*]');
  });

  it('shows the current trick and the winner at the end', () => {
    expect(formatColourWhistState(base)).toContain('seat 1:');
    expect(formatColourWhistState({ ...base, gameEndFlag: true, winnerIdx: 1 })).toContain('winner: seat 1');
  });

  it('handles an unknown phase and a missing config', () => {
    const out = formatColourWhistState({ ...base, phase: 99, config: undefined, declarerIdx: -1 });
    expect(out).toContain('phase: UNKNOWN');
    expect(out).not.toContain(' / 8');
    expect(out).not.toContain('declarer:');
  });

  it('shows the server message when present', () => {
    expect(formatColourWhistState({ ...base, message: 'その札は出せません' })).toContain('その札は出せません');
  });
});
