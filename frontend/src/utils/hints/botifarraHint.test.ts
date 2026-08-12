import { describe, expect, it } from 'vitest';
import type { BotifarraResponse } from '../../types/card';
import { BOTIFARRA_NO_TRUMP } from '../../types/games/botifarra';
import { BotifarraPhase } from '../../types/phases';
import { getBotifarraHint } from './botifarraHint';

const base: BotifarraResponse = {
  players: [],
  phase: BotifarraPhase.DECLARE,
  validPlays: [],
  dealerIdx: 0,
  declarerIdx: -1,
  trumpSuit: BOTIFARRA_NO_TRUMP,
  multiplier: 1,
  currentTurn: 0,
  isHumanTurn: true,
  currentTrick: [],
  lastTrick: [],
  lastTrickWinner: -1,
  trickCount: 0,
  roundPoints: [0, 0],
  scores: [0, 0],
  gameEndFlag: false,
  winnerTeam: -1,
  message: '',
};

describe('getBotifarraHint', () => {
  it('advises the longest suit while declaring', () => {
    expect(getBotifarraHint(base)?.reason).toBe('frontendHint.botifarraDeclareLongest');
    expect(getBotifarraHint({ ...base, phase: BotifarraPhase.DELEGATED })?.reason).toBe(
      'frontendHint.botifarraDeclareLongest',
    );
  });

  it('names the obligation while playing', () => {
    expect(getBotifarraHint({ ...base, phase: BotifarraPhase.PLAY })?.reason).toBe('frontendHint.botifarraMustWin');
  });

  it('stays quiet when it is not your turn or the game is over', () => {
    expect(getBotifarraHint({ ...base, phase: BotifarraPhase.PLAY, isHumanTurn: false })).toBeNull();
    expect(getBotifarraHint({ ...base, phase: BotifarraPhase.DOUBLE })).toBeNull();
    expect(getBotifarraHint({ ...base, gameEndFlag: true })).toBeNull();
  });
});
