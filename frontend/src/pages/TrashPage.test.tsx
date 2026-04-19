import { screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { trashApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, TrashResponse, TrashSlot } from '../types/card';
import { TrashPage } from './TrashPage';

vi.mock('../api/gameApi', () => ({
  trashApi: { exec: vi.fn() },
  actionLogApi: { trash: vi.fn() },
}));

const mockExec = vi.mocked(trashApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

function faceDownSlots(): TrashSlot[] {
  return Array.from({ length: 10 }, () => ({ faceUp: false }));
}

function flippedSlots(): TrashSlot[] {
  return Array.from({ length: 10 }, (_, i) => ({ faceUp: true, card: card('SPADE', i + 1) }));
}

const playerTurnState: TrashResponse = {
  phase: 0,
  current: 0,
  players: [
    { slots: faceDownSlots(), isCpu: false },
    { slots: faceDownSlots(), isCpu: true },
  ],
  stockSize: 34,
  discardSize: 0,
  moveCount: 0,
  winner: -1,
  message: '',
  messageCode: 'trash.playerTurn',
};

const awaitWildState: TrashResponse = {
  ...playerTurnState,
  phase: 1,
  pending: card('DIAMOND', 13),
  messageCode: 'trash.awaitWild',
};

const gameOverWinState: TrashResponse = {
  ...playerTurnState,
  phase: 2,
  players: [
    { slots: flippedSlots(), isCpu: false },
    { slots: faceDownSlots(), isCpu: true },
  ],
  winner: 0,
  moveCount: 12,
  messageCode: 'trash.gameOverWin',
  messageParams: { moveCount: '12' },
};

beforeEach(() => {
  vi.clearAllMocks();
  localStorage.clear();
  mockExec.mockResolvedValue(playerTurnState);
});

describe('TrashPage', () => {
  it('renders the page heading', async () => {
    renderWithProviders(<TrashPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });

  it('dispatches reset on mount', async () => {
    renderWithProviders(<TrashPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows the draw-turn phase name', async () => {
    renderWithProviders(<TrashPage />);
    await waitFor(() => expect(screen.getAllByText(/あなたのターン/).length).toBeGreaterThan(0));
  });

  it('shows the await-wild phase and pending card label', async () => {
    mockExec.mockResolvedValue(awaitWildState);
    renderWithProviders(<TrashPage />);
    await waitFor(() => expect(screen.getAllByText(/ワイルド配置/).length).toBeGreaterThan(0));
    expect(screen.getByText(/手札/)).toBeInTheDocument();
  });

  it('renders 10 slots per player', async () => {
    renderWithProviders(<TrashPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    // 20 slot buttons total (10 opponent + 10 self)
    const slotButtons = screen.getAllByRole('button', { name: /face-down|\d+:/ });
    expect(slotButtons.length).toBeGreaterThanOrEqual(20);
  });

  it('shows the win banner on gameOverWin', async () => {
    mockExec.mockResolvedValue(gameOverWinState);
    renderWithProviders(<TrashPage />);
    await waitFor(() => expect(screen.getAllByText(/ゲーム終了/).length).toBeGreaterThan(0));
  });

  it('surfaces an error alert when the API fails on mount', async () => {
    mockExec.mockRejectedValue(new Error('network error'));
    renderWithProviders(<TrashPage />);
    await waitFor(() => expect(screen.getByRole('alert')).toBeInTheDocument());
  });
});
