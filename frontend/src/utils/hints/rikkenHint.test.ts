import { describe, expect, it } from 'vitest';
import type { RikkenResponse } from '../../types/card';
import { RIKKEN_NO_TRUMP } from '../../types/games/rikken';
import { RikkenPhase } from '../../types/phases';
import { getRikkenHint } from './rikkenHint';

const base: RikkenResponse = {
  players: [],
  phase: RikkenPhase.BID,
  validPlays: [],
  dealerIdx: 0,
  contract: 0,
  declarerIdx: -1,
  partnerIdx: -1,
  trumpSuit: RIKKEN_NO_TRUMP,
  currentTurn: 0,
  isHumanTurn: true,
  currentTrick: [],
  lastTrick: [],
  lastTrickWinner: -1,
  trickCount: 0,
  declarerTricks: 0,
  roundNumber: 1,
  gameEndFlag: false,
  winnerIdx: -1,
  message: '',
};

describe('getRikkenHint', () => {
  it('advises on the contract while bidding', () => {
    expect(getRikkenHint(base)?.reason).toBe('frontendHint.rikkenBidStrength');
  });

  it('names the follow-suit rule while playing', () => {
    expect(getRikkenHint({ ...base, phase: RikkenPhase.PLAY })?.reason).toBe('frontendHint.rikkenFollowSuit');
  });

  it('stays quiet when it is not your turn or the game is over', () => {
    expect(getRikkenHint({ ...base, isHumanTurn: false })).toBeNull();
    expect(getRikkenHint({ ...base, phase: RikkenPhase.CALL })).toBeNull();
    expect(getRikkenHint({ ...base, gameEndFlag: true })).toBeNull();
  });
});
