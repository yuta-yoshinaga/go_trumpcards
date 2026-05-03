import { screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { pitchApi } from '../api/gameApi';
import { useCliMode } from '../hooks/useCliMode';
import { renderWithProviders } from '../test/renderWithProviders';
import type { CardDesign, PitchResponse } from '../types/card';
import { PitchPhase } from '../types/phases';
import { PitchPage } from './PitchPage';

vi.mock('../api/gameApi', () => ({
  pitchApi: { exec: vi.fn() },
  actionLogApi: { pitch: vi.fn() },
}));

vi.mock('../hooks/useCliMode', () => ({
  useCliMode: vi.fn(() => ({
    cliEnabled: false,
    toggleCli: vi.fn(),
    logEntries: [],
    addInput: vi.fn(),
    addOutput: vi.fn(),
    addError: vi.fn(),
    clearLog: vi.fn(),
  })),
}));

const mockApi = vi.mocked(pitchApi.exec);
const _mockUseCliMode = vi.mocked(useCliMode);

const makeCard = (design: CardDesign, value: number) => ({ design, value });

const baseConfig = { cpuDifficulty: 1, pointLimit: 7 };

const bidState: PitchResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 6,
      cards: [makeCard('SPADE', 5), makeCard('HEART', 9)],
      bid: -1,
      roundScore: 0,
      cumulativeScore: 0,
      trickCount: 0,
    },
    { id: 1, isHuman: false, cardCount: 6, cards: [], bid: -1, roundScore: 0, cumulativeScore: 0, trickCount: 0 },
    { id: 2, isHuman: false, cardCount: 6, cards: [], bid: -1, roundScore: 0, cumulativeScore: 0, trickCount: 0 },
    { id: 3, isHuman: false, cardCount: 6, cards: [], bid: -1, roundScore: 0, cumulativeScore: 0, trickCount: 0 },
  ],
  phase: PitchPhase.BID,
  roundNumber: 1,
  trickNumber: 0,
  dealerIdx: 3,
  currentPlayerIdx: -1,
  bidPlayerIdx: 0,
  currentBid: 0,
  bidWinnerIdx: -1,
  trumpSuit: 0,
  currentTrick: [],
  gameEndFlag: false,
  winnerIdx: -1,
  leadPlayerIdx: -1,
  validPlayIndices: [],
  message: '',
  config: baseConfig,
};

const playState: PitchResponse = {
  ...bidState,
  phase: PitchPhase.PLAY,
  trickNumber: 1,
  currentPlayerIdx: 0,
  bidPlayerIdx: 0,
  currentBid: 3,
  bidWinnerIdx: 0,
  trumpSuit: 1,
  validPlayIndices: [0, 1],
  leadPlayerIdx: 0,
  players: bidState.players.map((p, i) => (i === 0 ? { ...p, bid: 3 } : { ...p, bid: 0 })),
};

const gameEndState: PitchResponse = {
  ...bidState,
  phase: PitchPhase.GAME_END,
  gameEndFlag: true,
  winnerIdx: 0,
};

beforeEach(() => {
  vi.clearAllMocks();
  mockApi.mockResolvedValue(bidState);
});

describe('PitchPage', () => {
  it('calls reset on mount', async () => {
    renderWithProviders(<PitchPage />);
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('reset'));
  });

  it('renders bid phase with pass + bid buttons', async () => {
    mockApi.mockResolvedValue(bidState);
    renderWithProviders(<PitchPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /パス/ })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: /ビッド 2/ })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /ビッド 3/ })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /ビッド 4/ })).toBeInTheDocument();
  });

  it('renders play phase with player hand', async () => {
    mockApi.mockResolvedValue(playState);
    renderWithProviders(<PitchPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /出す/ })).toBeInTheDocument());
  });

  it('shows score table with players', async () => {
    mockApi.mockResolvedValue(bidState);
    renderWithProviders(<PitchPage />);
    await waitFor(() => expect(screen.getAllByText(/あなた/).length).toBeGreaterThan(0));
    expect(screen.getAllByText(/CPU 1/).length).toBeGreaterThan(0);
  });

  it('shows winner banner on game end', async () => {
    mockApi.mockResolvedValue(gameEndState);
    renderWithProviders(<PitchPage />);
    await waitFor(() => expect(screen.getByText(/あなたの勝利/)).toBeInTheDocument());
  });
});
