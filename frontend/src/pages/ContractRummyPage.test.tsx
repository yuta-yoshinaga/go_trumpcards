import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { contractrummyApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, ContractRummyResponse } from '../types/card';
import { ContractRummyPage } from './ContractRummyPage';

vi.mock('../api/gameApi', () => ({
  contractrummyApi: { exec: vi.fn() },
  actionLogApi: { contractrummy: vi.fn() },
}));

const mockExec = vi.mocked(contractrummyApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

const baseHand: Card[] = [
  card('SPADE', 5),
  card('HEART', 5),
  card('DIAMOND', 5),
  card('CLOVER', 13),
  card('SPADE', 13),
  card('HEART', 13),
  card('SPADE', 2),
  card('SPADE', 3),
  card('SPADE', 4),
  card('SPADE', 6),
  card('SPADE', 7),
];

const drawState: ContractRummyResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 11,
      cards: baseHand,
      melds: [],
      contractMet: false,
      roundScore: 0,
      cumulativeScore: 0,
    },
    {
      id: 1,
      isHuman: false,
      cardCount: 11,
      cards: [],
      melds: [],
      contractMet: false,
      roundScore: 0,
      cumulativeScore: 0,
    },
    {
      id: 2,
      isHuman: false,
      cardCount: 11,
      cards: [],
      melds: [],
      contractMet: false,
      roundScore: 0,
      cumulativeScore: 0,
    },
  ],
  phase: 0,
  roundNumber: 1,
  totalRounds: 7,
  currentPlayerIdx: 0,
  discardTop: card('HEART', 7),
  drawPileCount: 60,
  gameEndFlag: false,
  winnerIdx: -1,
  roundWinnerIdx: -1,
  contractSlots: [
    { kind: 0, size: 3 },
    { kind: 0, size: 3 },
  ],
  config: { cpuDifficulty: 1, failContractPenalty: 25 },
  message: '',
};

const playState: ContractRummyResponse = { ...drawState, phase: 1 };

const roundEndState: ContractRummyResponse = {
  ...drawState,
  phase: 2,
  roundWinnerIdx: 0,
};

beforeEach(() => {
  vi.clearAllMocks();
  localStorage.clear();
  mockExec.mockResolvedValue(drawState);
});

describe('ContractRummyPage', () => {
  it('calls reset on mount', async () => {
    renderWithProviders(<ContractRummyPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows the round and contract banner', async () => {
    renderWithProviders(<ContractRummyPage />);
    await waitFor(() => expect(screen.getByText(/Round 1 \/ 7|ラウンド 1 \/ 7/)).toBeInTheDocument());
  });

  it('shows draw-phase action buttons during human draw turn', async () => {
    renderWithProviders(<ContractRummyPage />);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: /Draw from stock|山札から引く/ })).toBeInTheDocument(),
    );
    expect(screen.getByRole('button', { name: /Take discard|捨て札を取る/ })).toBeInTheDocument();
  });

  it('invokes drawstock when the stock button is clicked', async () => {
    renderWithProviders(<ContractRummyPage />);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: /Draw from stock|山札から引く/ })).toBeInTheDocument(),
    );
    fireEvent.click(screen.getByRole('button', { name: /Draw from stock|山札から引く/ }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('drawstock'));
  });

  it('shows play-phase action buttons after drawing', async () => {
    mockExec.mockResolvedValue(playState);
    renderWithProviders(<ContractRummyPage />);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: /Submit contract|コントラクトを場に出す/ })).toBeInTheDocument(),
    );
  });

  it('shows next-round button at round end', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<ContractRummyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /Next round|次のラウンドへ/ })).toBeInTheDocument());
  });
});
