import { act, fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { montecarloApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, MonteCarloBoardCell, MonteCarloResponse } from '../types/card';
import { MonteCarloPage } from './MonteCarloPage';

vi.mock('../api/gameApi', () => ({
  montecarloApi: { exec: vi.fn() },
  actionLogApi: { montecarlo: vi.fn() },
}));

const mockExec = vi.mocked(montecarloApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });
const empty: MonteCarloBoardCell = { card: null };

function emptyBoard(): MonteCarloBoardCell[][] {
  return Array.from({ length: 5 }, () => Array.from({ length: 5 }, () => ({ ...empty })));
}

function boardWithPair(): MonteCarloBoardCell[][] {
  const b = emptyBoard();
  b[0][0] = { card: card('SPADE', 7) };
  b[0][1] = { card: card('HEART', 7) };
  b[2][2] = { card: card('CLOVER', 5) };
  return b;
}

const playingState: MonteCarloResponse = {
  board: boardWithPair(),
  phase: 0,
  stockCount: 27,
  removedCount: 0,
  dealCount: 0,
  canUndo: false,
  isStalemate: false,
  message: '',
  messageCode: 'montecarlo.playing',
};

const gameClearState: MonteCarloResponse = {
  ...playingState,
  phase: 1,
  stockCount: 0,
  removedCount: 52,
  dealCount: 5,
  messageCode: 'montecarlo.gameClear',
  messageParams: { dealCount: '5', removedCount: '52' },
};

const gameOverState: MonteCarloResponse = {
  ...playingState,
  phase: 2,
  messageCode: 'montecarlo.gameOver',
};

beforeEach(() => {
  vi.clearAllMocks();
  localStorage.clear();
  mockExec.mockResolvedValue(playingState);
});

describe('MonteCarloPage', () => {
  it('renders heading', async () => {
    renderWithProviders(<MonteCarloPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<MonteCarloPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows stock and removed counts in the header', async () => {
    renderWithProviders(<MonteCarloPage />);
    await waitFor(() => expect(screen.getAllByText(/27/).length).toBeGreaterThan(0));
    expect(screen.getByText(/0\/52/)).toBeInTheDocument();
  });

  it('renders 5x5 grid with filled cells as buttons', async () => {
    renderWithProviders(<MonteCarloPage />);
    await waitFor(() => expect(screen.getByTestId('mc-cell-0-0')).toBeInTheDocument());
    expect(screen.getByTestId('mc-cell-4-4')).toBeInTheDocument();
  });

  it('clicking a filled cell selects it; clicking again deselects', async () => {
    renderWithProviders(<MonteCarloPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    const cell00 = screen.getByTestId('mc-cell-0-0');
    fireEvent.click(cell00);
    expect(cell00.className).toContain('ring-ds-accent');
    fireEvent.click(cell00);
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('clicking two adjacent filled cells fires remove command', async () => {
    renderWithProviders(<MonteCarloPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('mc-cell-0-0'));
    fireEvent.click(screen.getByTestId('mc-cell-0-1'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('remove', 0, 0, 0, 1));
  });

  it('deal button fires deal command', async () => {
    renderWithProviders(<MonteCarloPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('mc-deal-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('deal'));
  });

  it('giveup button opens a confirm dialog and only dispatches giveup after confirm', async () => {
    renderWithProviders(<MonteCarloPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'ギブアップ' }));
    expect(mockExec).not.toHaveBeenCalledWith('giveup');
    expect(screen.getByText('投了確認')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('undo button is disabled when canUndo is false', async () => {
    renderWithProviders(<MonteCarloPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByRole('button', { name: '元に戻す' })).toBeDisabled();
  });

  it('undo button fires undo when canUndo is true', async () => {
    mockExec.mockResolvedValue({ ...playingState, canUndo: true });
    renderWithProviders(<MonteCarloPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    screen.getByRole('button', { name: '元に戻す' }).click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo'));
  });

  it('shows game clear celebration', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<MonteCarloPage />);
    await waitFor(() => expect(screen.getAllByText(/ゲームクリア/).length).toBeGreaterThan(0));
  });

  it('shows game over phase', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<MonteCarloPage />);
    await waitFor(() => expect(screen.getAllByText(/ゲームオーバー/).length).toBeGreaterThan(0));
  });

  it('shows error alert when API fails on mount', async () => {
    mockExec.mockRejectedValue(new Error('network error'));
    renderWithProviders(<MonteCarloPage />);
    await waitFor(() => expect(screen.getByRole('alert')).toBeInTheDocument());
  });

  it('flags adjacent same-rank cell with pulsing success ring after selecting', async () => {
    // boardWithPair: [0][0] = ♠7, [0][1] = ♥7 — clicking [0][0] should mark [0][1] as a pair match.
    renderWithProviders(<MonteCarloPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    fireEvent.click(screen.getByTestId('mc-cell-0-0'));

    const partner = screen.getByTestId('mc-cell-0-1');
    expect(partner).toHaveAttribute('data-pair-match', 'true');
    expect(partner.className).toContain('ring-ds-success');
    expect(partner.className).toContain('animate-pulse');
  });

  it('dims adjacent cells with a different rank instead of ringing them', async () => {
    // Build a board where two adjacent cells (0,0) and (1,0) hold different ranks.
    const board = emptyBoard();
    board[0][0] = { card: card('SPADE', 7) };
    board[1][0] = { card: card('HEART', 9) };
    mockExec.mockResolvedValue({ ...playingState, board });
    renderWithProviders(<MonteCarloPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    fireEvent.click(screen.getByTestId('mc-cell-0-0'));

    const neighbor = screen.getByTestId('mc-cell-1-0');
    expect(neighbor).not.toHaveAttribute('data-pair-match');
    expect(neighbor.className).toContain('opacity-60');
    expect(neighbor.className).not.toContain('animate-pulse');
  });

  it('lifts a matching adjacent pair candidate and shows a transient success toast on removal', async () => {
    vi.useFakeTimers();
    try {
      const board = emptyBoard();
      board[0][0] = { card: card('SPADE', 7) };
      board[1][0] = { card: card('HEART', 7) }; // adjacent, same rank
      mockExec.mockResolvedValue({ ...playingState, board });
      renderWithProviders(<MonteCarloPage />);
      // Use vi.waitFor (not RTL waitFor) here: RTL's waitFor relies on real timers
      // and would hang under vi.useFakeTimers().
      await vi.waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

      fireEvent.click(screen.getByTestId('mc-cell-0-0'));
      const match = screen.getByTestId('mc-cell-1-0');
      expect(match).toHaveAttribute('data-pair-match', 'true');
      expect(match.className).toContain('-translate-y-1');

      // Removing the valid pair flashes the toast, which auto-dismisses after 1s.
      fireEvent.click(match);
      expect(screen.getByTestId('mc-pair-toast')).toBeInTheDocument();
      act(() => {
        vi.advanceTimersByTime(1000);
      });
      expect(screen.queryByTestId('mc-pair-toast')).not.toBeInTheDocument();
    } finally {
      vi.useRealTimers();
    }
  });
});
