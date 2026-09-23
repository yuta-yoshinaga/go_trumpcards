import { describe, expect, it } from 'vitest';
import type { HoldemResponse } from '../../../types/card';
import { HoldemPhase } from '../../../types/phases';
import { formatHoldemState } from './holdemFormatter';

const baseState: HoldemResponse = {
  players: [],
  communityCards: [],
  pot: 0,
  sidePots: [],
  dealerIdx: 0,
  currentTurn: 0,
  phase: HoldemPhase.INIT,
  gameEndFlag: false,
  lastBet: 0,
  minRaise: 20,
  roundResults: [],
  cpuActions: [],
  message: '',
  handCount: 0,
  smallBlind: 5,
  bigBlind: 10,
  tournamentMode: false,
  blindLevelHands: 10,
  blindMultiplier: 200,
  bettingLimit: 0,
  raiseCount: 0,
  maxBetAmount: 0,
  tableSize: 4,
  rebuyPhaseType: 0,
  rebuyChips: 0,
  rebuyMaxCount: 0,
  rebuyCounts: [],
  addonChips: 0,
  rebuyAvailable: false,
  addonAvailable: false,
  rebuyEnabled: false,
  addonEnabled: false,
  rebuyPeriodHands: 0,
  addonAfterHand: 0,
  addonUsed: [],
  muckAvailable: false,
};

describe('formatHoldemState', () => {
  it.each([
    [0, 'limit: Fixed'],
    [2, 'limit: No Limit'],
  ] as const)('shows the betting limit for limit %s', (bettingLimit, expectedLine) => {
    const output = formatHoldemState({ ...baseState, bettingLimit });

    expect(output).toContain(expectedLine);
  });
});
