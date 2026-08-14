import { describe, expect, it } from 'vitest';
import type { ColourWhistResponse } from '../../types/card';
import { COLOUR_WHIST_NO_TRUMP } from '../../types/games/colourwhist';
import { ColourWhistPhase } from '../../types/phases';
import { getColourwhistHint } from './colourwhistHint';

const base: ColourWhistResponse = {
  players: [],
  phase: ColourWhistPhase.BID,
  validPlays: [],
  dealerIdx: 0,
  contract: 0,
  declarerIdx: -1,
  partnerIdx: -1,
  trumpSuit: COLOUR_WHIST_NO_TRUMP,
  troelForced: false,
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

describe('getColourwhistHint', () => {
  it('advises on the contract while bidding', () => {
    expect(getColourwhistHint(base)?.reason).toBe('frontendHint.colourWhistBidStrength');
  });

  it('names the follow-suit rule while playing', () => {
    expect(getColourwhistHint({ ...base, phase: ColourWhistPhase.PLAY })?.reason).toBe(
      'frontendHint.colourWhistFollowSuit',
    );
  });

  it('stays quiet when it is not your turn or the game is over', () => {
    expect(getColourwhistHint({ ...base, isHumanTurn: false })).toBeNull();
    expect(getColourwhistHint({ ...base, phase: ColourWhistPhase.CALL })).toBeNull();
    expect(getColourwhistHint({ ...base, gameEndFlag: true })).toBeNull();
  });
});
