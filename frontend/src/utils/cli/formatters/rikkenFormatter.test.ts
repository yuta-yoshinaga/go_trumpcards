import { describe, expect, it } from 'vitest';
import type { Card, RikkenResponse } from '../../../types/card';
import { RIKKEN_NO_TRUMP } from '../../../types/games/rikken';
import { RikkenContract, RikkenPhase } from '../../../types/phases';
import { formatRikkenState } from './rikkenFormatter';

const cards: Card[] = [
  { design: 'SPADE', value: 1 },
  { design: 'HEART', value: 9 },
  { design: 'CLOVER', value: 4 },
];

const base: RikkenResponse = {
  players: [
    { id: 0, isHuman: true, cardCount: 3, cards, trickCount: 2, score: 9, isDeclarerSide: true, hasPassed: false },
    {
      id: 1,
      isHuman: false,
      cardCount: 3,
      cards: [],
      trickCount: 1,
      score: -3,
      isDeclarerSide: false,
      hasPassed: true,
    },
  ],
  phase: RikkenPhase.PLAY,
  validPlays: [1],
  dealerIdx: 0,
  contract: RikkenContract.SOLO,
  declarerIdx: 0,
  partnerIdx: -1,
  trumpSuit: 3,
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

describe('formatRikkenState', () => {
  it('shows the phase, round, contract and trump', () => {
    const out = formatRikkenState(base);
    expect(out).toContain('phase: PLAY');
    expect(out).toContain('round: 2 / 8');
    expect(out).toContain('contract: Solo (trump: Hearts)');
    expect(out).toContain('declarer: seat 0 (2 tricks)');
    // **負のコントロール: Rik 以外にパートナーはいない** (受け入れ条件3)。
    expect(out).not.toContain('partner:');
  });

  // **Rik のパートナーは指名札で決まる秘密の相方** (#5772)。カード表示は判明前も
  // "hidden" と出しているので、CLI モードだけ黙っているのは情報量の差になる。
  it('names the Rik partner, and says hidden until they are known', () => {
    const known = formatRikkenState({ ...base, contract: RikkenContract.RIK, partnerIdx: 2 });
    expect(known).toContain('partner: seat 2');

    const unknown = formatRikkenState({ ...base, contract: RikkenContract.RIK, partnerIdx: -1 });
    expect(unknown).toContain('partner: hidden');
  });

  // **切り札なしは名前で出す。** -1 をそのまま表示しません。
  it('names no trump rather than printing -1', () => {
    const out = formatRikkenState({ ...base, contract: RikkenContract.MISERE, trumpSuit: RIKKEN_NO_TRUMP });
    expect(out).toContain('contract: Misere (trump: none)');
  });

  // **得点は負も出す。** ゼロサムなので当然そうなります。
  it('shows negative scores', () => {
    expect(formatRikkenState(base)).toContain('#1:-3');
  });

  it('marks only the legal cards in hand', () => {
    const out = formatRikkenState(base);
    expect(out).toContain('[0 ]');
    expect(out).toContain('[1*]');
    expect(out).toContain('[2 ]');
  });

  it('shows the current trick and the winner at the end', () => {
    expect(formatRikkenState(base)).toContain('seat 1:');
    expect(formatRikkenState({ ...base, gameEndFlag: true, winnerIdx: 2 })).toContain('winner: seat 2');
  });

  it('handles an unknown phase and a missing config', () => {
    const out = formatRikkenState({ ...base, phase: 99, config: undefined, declarerIdx: -1 });
    expect(out).toContain('phase: UNKNOWN');
    expect(out).not.toContain(' / 8');
    expect(out).not.toContain('declarer:');
  });

  it('shows the server message when present', () => {
    expect(formatRikkenState({ ...base, message: 'その札は出せません' })).toContain('その札は出せません');
  });
});
