import { describe, expect, it } from 'vitest';
import type { BotifarraResponse } from '../../../types/card';
import { BOTIFARRA_NO_TRUMP } from '../../../types/games/botifarra';
import { BotifarraPhase } from '../../../types/phases';
import { formatBotifarraState } from './botifarraFormatter';

const base: BotifarraResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      team: 0,
      cardCount: 3,
      cards: [
        { design: 'SPADE', value: 9 },
        { design: 'CLOVER', value: 1 },
        { design: 'HEART', value: 13 },
      ],
      trickCount: 0,
    },
  ],
  phase: BotifarraPhase.PLAY,
  validPlays: [1],
  dealerIdx: 0,
  declarerIdx: 0,
  trumpSuit: 3,
  multiplier: 2,
  currentTurn: 0,
  isHumanTurn: true,
  currentTrick: [{ playerIdx: 1, card: { design: 'CLOVER', value: 12 } }],
  lastTrick: [],
  lastTrickWinner: -1,
  trickCount: 4,
  roundPoints: [20, 12],
  scores: [40, 31],
  gameEndFlag: false,
  winnerTeam: -1,
  config: { targetScore: 101, allowDoubling: true },
  message: '',
};

describe('formatBotifarraState', () => {
  it('shows the phase, score, trump and the round total', () => {
    const out = formatBotifarraState(base);
    expect(out).toContain('phase: PLAY');
    expect(out).toContain('score: your team 40 / theirs 31');
    expect(out).toContain('game to 101');
    expect(out).toContain('trump: Hearts (x2)');
    expect(out).toContain('this round: 20 / 12 of 72');
  });

  // **切り札なしは名前で出す。** -1 をそのまま表示しません。
  it('names no trump rather than printing -1', () => {
    const out = formatBotifarraState({ ...base, trumpSuit: BOTIFARRA_NO_TRUMP });
    expect(out).toContain('trump: No trump');
    expect(out).not.toContain('-1');
  });

  // **出せる札に印が付く。** 勝つ義務があるので大半が出せない場面があります。
  it('marks only the legal cards in hand', () => {
    const out = formatBotifarraState(base);
    expect(out).toContain('[0 ]');
    expect(out).toContain('[1*]');
    expect(out).toContain('[2 ]');
    expect(out).toContain('only cards marked * are legal');
  });

  it('shows the current trick and the winner at the end', () => {
    expect(formatBotifarraState(base)).toContain('seat 1:');
    const done = formatBotifarraState({ ...base, gameEndFlag: true, winnerTeam: 0 });
    expect(done).toContain('winner: team 0');
  });

  it('handles an unknown phase and a missing config', () => {
    const out = formatBotifarraState({ ...base, phase: 99, config: undefined });
    expect(out).toContain('phase: UNKNOWN');
    expect(out).not.toContain('game to');
  });

  it('shows the server message when present', () => {
    expect(formatBotifarraState({ ...base, message: 'その札は出せません' })).toContain('その札は出せません');
  });
});
