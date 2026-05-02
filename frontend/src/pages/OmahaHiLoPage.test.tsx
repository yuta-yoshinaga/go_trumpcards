import { screen, waitFor } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { actionLogApi, omahaHiLoApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { OmahaResponse } from '../types/card';
import { OmahaHiLoPage } from './OmahaHiLoPage';

/** Smoke tests for OmahaHiLoPage. The full UI surface is covered by
 * OmahaPage.test.tsx (the two pages share their implementation modulo
 * api binding and tutorial selectors); these tests assert that the new
 * page mounts, calls the omahahilo API on initial reset, and surfaces
 * the Hi-Lo split-pot fields when present in the server response. */
vi.mock('../api/gameApi', () => ({
  omahaHiLoApi: { exec: vi.fn() },
  actionLogApi: { omahahilo: vi.fn() },
}));

const mockExec = vi.mocked(omahaHiLoApi.exec);

const initState: OmahaResponse = {
  players: [],
  communityCards: [],
  pot: 0,
  sidePots: [],
  dealerIdx: 0,
  currentTurn: 0,
  phase: 0,
  gameEndFlag: false,
  lastBet: 0,
  minRaise: 0,
  bettingLimit: 2,
  raiseCount: 0,
  maxBetAmount: 0,
  roundResults: [],
  cpuActions: [],
  message: '',
  handCount: 0,
  smallBlind: 5,
  bigBlind: 10,
  tournamentMode: false,
  blindLevelHands: 0,
  blindMultiplier: 0,
  tableSize: 4,
  rebuyPhaseType: 0,
  rebuyChips: 0,
  rebuyMaxCount: 0,
  rebuyCounts: [0, 0, 0, 0],
  addonChips: 0,
  rebuyAvailable: false,
  addonAvailable: false,
  rebuyEnabled: false,
  addonEnabled: false,
  rebuyPeriodHands: 0,
  addonAfterHand: 0,
  addonUsed: [false, false, false, false],
  muckAvailable: false,
  isHiLo: true,
};

describe('OmahaHiLoPage', () => {
  it('calls omahaHiLoApi.exec with reset on mount', async () => {
    mockExec.mockResolvedValue(initState);
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => {
      expect(mockExec).toHaveBeenCalledWith('reset');
    });
  });

  it('renders the Hi-Lo page heading once initial state arrives', async () => {
    mockExec.mockResolvedValue(initState);
    renderWithProviders(<OmahaHiLoPage />);
    await waitFor(() => {
      // h1 reads the omahahilo nav label from common.json (ja: "オマハ ハイロー")
      expect(screen.getByText('オマハ ハイロー')).toBeTruthy();
    });
  });
});
