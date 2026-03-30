import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { speedApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { SpeedResponse } from '../types/card';
import { SpeedPage } from './SpeedPage';

vi.mock('../api/gameApi', () => ({
  speedApi: { exec: vi.fn() },
  actionLogApi: { speed: vi.fn() },
}));

const mockExec = vi.mocked(speedApi.exec);

const playState: SpeedResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 4,
      cards: [
        { design: 'SPADE', value: 4 },
        { design: 'HEART', value: 8 },
        { design: 'CLOVER', value: 11 },
        { design: 'DIAMOND', value: 2 },
      ],
      drawPileSize: 21,
    },
    { id: 1, isHuman: false, cardCount: 4, cards: [], drawPileSize: 21 },
  ],
  centerPiles: [
    { design: 'DIAMOND', value: 5 },
    { design: 'SPADE', value: 9 },
  ],
  phase: 0,
  gameEndFlag: false,
  winnerIdx: -1,
  config: { cpuDifficulty: 1 },
  message: '',
};

const stuckState: SpeedResponse = {
  ...playState,
  phase: 1,
};

const gameEndState: SpeedResponse = {
  ...playState,
  phase: 2,
  gameEndFlag: true,
  winnerIdx: 0,
};

describe('SpeedPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(playState);
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<SpeedPage />);
    await waitFor(() => {
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, { cpuDifficulty: 1 });
    });
  });

  it('renders the page heading', () => {
    renderWithProviders(<SpeedPage />);
    expect(screen.getByText('スピード')).toBeInTheDocument();
  });

  it('renders player hand after API resolves', async () => {
    renderWithProviders(<SpeedPage />);
    // Wait for the exec call to happen (state loaded)
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    // Then verify state-dependent content renders
    await waitFor(() => expect(screen.getByText('手札')).toBeInTheDocument());
  });

  it('shows stuck phase flip button', async () => {
    mockExec.mockResolvedValue(stuckState);
    renderWithProviders(<SpeedPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    await waitFor(() => expect(screen.getByText('めくる')).toBeInTheDocument());
  });

  it('calls flip on button click', async () => {
    mockExec.mockResolvedValue(stuckState);
    renderWithProviders(<SpeedPage />);
    await waitFor(() => expect(screen.getByText('めくる')).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'めくる' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('flip'));
  });

  it('shows game end phase', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<SpeedPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    await waitFor(() => expect(screen.getByText('ゲーム終了')).toBeInTheDocument());
  });
});
