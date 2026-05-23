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
void useCliMode; // kept for vi.mock side-effect; unused locally

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

  it('renders the Game-pip badge with the human hand total', async () => {
    // SPADE 5 (0 pips) + HEART 9 (0 pips) = 0
    mockApi.mockResolvedValue(bidState);
    renderWithProviders(<PitchPage />);
    const badge = await screen.findByTestId('pitch-game-pips-badge');
    expect(badge.textContent).toMatch(/Game値: 0/);
  });

  it('Game-pip badge tooltip has no trailing newline when pips = 0', async () => {
    // SPADE 5 (0) + HEART 9 (0) → no breakdown line should be appended
    mockApi.mockResolvedValue(bidState);
    renderWithProviders(<PitchPage />);
    const badge = await screen.findByTestId('pitch-game-pips-badge');
    const title = badge.getAttribute('title') ?? '';
    expect(title).not.toMatch(/\n$/);
    expect(title).not.toContain('\n');
  });

  it('Game-pip badge sums A, K, Q, J, 10 correctly', async () => {
    const pipHand: PitchResponse = {
      ...bidState,
      players: [
        {
          ...bidState.players[0],
          // A(4) + 10(10) + J(1) + K(3) + Q(2) + 7(0) = 20
          cards: [
            makeCard('SPADE', 1),
            makeCard('SPADE', 10),
            makeCard('SPADE', 11),
            makeCard('SPADE', 13),
            makeCard('SPADE', 12),
            makeCard('HEART', 7),
          ],
          cardCount: 6,
        },
        ...bidState.players.slice(1),
      ],
    };
    mockApi.mockResolvedValue(pipHand);
    renderWithProviders(<PitchPage />);
    const badge = await screen.findByTestId('pitch-game-pips-badge');
    expect(badge.textContent).toMatch(/Game値: 20/);
    // Tooltip contains the breakdown of contributing cards only.
    expect(badge.getAttribute('title')).toContain('A(4)');
    expect(badge.getAttribute('title')).toContain('10(10)');
    expect(badge.getAttribute('title')).not.toContain('7(');
  });
});
