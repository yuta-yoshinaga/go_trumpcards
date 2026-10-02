import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { irishPokerApi, pineappleApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { HoldemPlayerData, PineappleResponse } from '../types/card';
import { IrishPokerPage } from './IrishPokerPage';

vi.mock('../api/gameApi', () => ({
  pineappleApi: { exec: vi.fn() },
  crazyPineappleApi: { exec: vi.fn() },
  irishPokerApi: { exec: vi.fn() },
  actionLogApi: { irishpoker: vi.fn() },
}));

const mockIrishExec = vi.mocked(irishPokerApi.exec);
const mockPineappleExec = vi.mocked(pineappleApi.exec);

const humanPlayer = (overrides: Partial<HoldemPlayerData> = {}): HoldemPlayerData => ({
  id: 0,
  isHuman: true,
  cards: [
    { design: 'SPADE', value: 1 },
    { design: 'HEART', value: 13 },
    { design: 'DIAMOND', value: 7 },
    { design: 'CLOVER', value: 4 },
  ],
  chips: 980,
  currentBet: 0,
  folded: false,
  allIn: false,
  handRank: 0,
  handName: '',
  bestHand: [],
  playStyleName: '',
  totalHands: 0,
  vpip: 0,
  pfr: 0,
  threeBet: 0,
  af: '-',
  ...overrides,
});

const cpuPlayer = (id: number, overrides: Partial<HoldemPlayerData> = {}): HoldemPlayerData => ({
  id,
  isHuman: false,
  cards: [
    { design: 'DIAMOND', value: 2 },
    { design: 'CLOVER', value: 7 },
  ],
  chips: 1000,
  currentBet: 0,
  folded: false,
  allIn: false,
  handRank: 0,
  handName: '',
  bestHand: [],
  playStyleName: '',
  totalHands: 0,
  vpip: 0,
  pfr: 0,
  threeBet: 0,
  af: '-',
  ...overrides,
});

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
  raiseCount: 0,
  maxBetAmount: 0,
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
  rebuyAvailable: false,
  addonEnabled: false,
  addonChips: 500,
  addonAfterHand: 5,
  addonUsed: [],
  addonAvailable: false,
  rebuyPhaseType: 0,
  muckAvailable: false,
  tableSize: 4,
  isDiscardPhase: false,
  discardDone: [],
  initialDealCount: 4,
  liveBestHand: '',
};

const discardState: PineappleResponse = {
  ...initState,
  players: [humanPlayer(), cpuPlayer(1), cpuPlayer(2), cpuPlayer(3)],
  pot: 60,
  dealerIdx: 3,
  currentTurn: 0,
  phase: 8,
  isDiscardPhase: true,
  discardDone: [false, true, true, true],
  communityCards: [
    { design: 'SPADE', value: 10 },
    { design: 'HEART', value: 5 },
    { design: 'DIAMOND', value: 8 },
  ],
  message: '',
  handCount: 1,
  initialDealCount: 4,
};

describe('IrishPokerPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockIrishExec.mockResolvedValue(initState);
    mockPineappleExec.mockResolvedValue(initState);
  });

  it('calls irishPokerApi.exec on mount (NOT pineappleApi)', async () => {
    renderWithProviders(<IrishPokerPage />);
    await waitFor(() => {
      expect(mockIrishExec).toHaveBeenCalledWith('reset');
    });
    expect(mockPineappleExec).not.toHaveBeenCalled();
  });

  it('renders discard controls during discard phase', async () => {
    mockIrishExec.mockResolvedValue(discardState);
    renderWithProviders(<IrishPokerPage />);
    await waitFor(() => expect(screen.getByTestId('discard-controls')).toBeInTheDocument());
  });

  it('requires selecting 2 cards then submits them together as cardIdxs', async () => {
    mockIrishExec.mockResolvedValue(discardState);
    renderWithProviders(<IrishPokerPage />);
    await waitFor(() => expect(screen.getByTestId('discard-controls')).toBeInTheDocument());

    const discardBtn = screen.getByRole('button', { name: 'カードを捨ててください。' });
    expect(discardBtn).toBeDisabled(); // 0 selected

    const cardButtons = screen.getAllByRole('button').filter((b) => b.getAttribute('aria-pressed') !== null);
    fireEvent.click(cardButtons[2]); // 1/2 selected → still disabled, count badge shows
    expect(screen.getByTestId('discard-count')).toHaveTextContent('1/2');
    expect(discardBtn).toBeDisabled();

    fireEvent.click(cardButtons[3]); // 2/2 selected → enabled
    await waitFor(() => expect(discardBtn).not.toBeDisabled());

    fireEvent.click(discardBtn);
    fireEvent.click(screen.getByRole('button', { name: '確定' }));
    await waitFor(() => expect(mockIrishExec).toHaveBeenCalledWith('discard', undefined, { cardIdxs: [2, 3] }));
  });

  it('caps the discard selection at 2 cards', async () => {
    mockIrishExec.mockResolvedValue(discardState);
    renderWithProviders(<IrishPokerPage />);
    await waitFor(() => expect(screen.getByTestId('discard-controls')).toBeInTheDocument());

    const cardButtons = screen.getAllByRole('button').filter((b) => b.getAttribute('aria-pressed') !== null);
    fireEvent.click(cardButtons[0]);
    fireEvent.click(cardButtons[1]);
    fireEvent.click(cardButtons[2]); // ignored — cap is 2
    expect(screen.getByTestId('discard-count')).toHaveTextContent('2/2');
    expect(cardButtons[2]).toHaveAttribute('aria-pressed', 'false');
  });
});
