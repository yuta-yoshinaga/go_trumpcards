import { waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { crazyPineappleApi, pineappleApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { PineappleResponse } from '../types/card';
import { CrazyPineapplePage } from './CrazyPineapplePage';

vi.mock('../api/gameApi', () => ({
  pineappleApi: { exec: vi.fn() },
  crazyPineappleApi: { exec: vi.fn() },
  actionLogApi: { crazypineapple: vi.fn() },
}));

const mockCrazyExec = vi.mocked(crazyPineappleApi.exec);
const mockPineappleExec = vi.mocked(pineappleApi.exec);

const initState: PineappleResponse = {
  players: [],
  communityCards: [],
  pot: 0,
  sidePots: [],
  dealerIdx: 0,
  currentTurn: 0,
  phase: 0,
  gameEndFlag: false,
  lastBet: 0,
  minRaise: 20,
  roundResults: [],
  cpuActions: [],
  message: '',
  handCount: 0,
  smallBlind: 5,
  bigBlind: 10,
  bettingLimit: 0,
  tournamentMode: false,
  blindLevelHands: 10,
  blindMultiplier: 150,
  rebuyEnabled: false,
  rebuyMaxCount: 2,
  rebuyChips: 1000,
  rebuyPeriodHands: 5,
  rebuyCounts: [],
  addonEnabled: false,
  addonChips: 500,
  addonAfterHand: 5,
  addonUsed: [],
  rebuyPhaseType: 0,
  muckAvailable: false,
  muckDone: false,
  shownCards: [],
  tableSize: 4,
  isDiscardPhase: false,
  discardDone: [],
};

describe('CrazyPineapplePage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockCrazyExec.mockResolvedValue(initState);
    mockPineappleExec.mockResolvedValue(initState);
  });

  it('calls crazyPineappleApi.exec on mount (NOT pineappleApi)', async () => {
    renderWithProviders(<CrazyPineapplePage />);
    await waitFor(() => {
      expect(mockCrazyExec).toHaveBeenCalledWith('reset');
    });
    expect(mockPineappleExec).not.toHaveBeenCalled();
  });
});
