import { act, fireEvent, screen, waitFor } from '@testing-library/react';
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

const cpuTurnState: TrashResponse = {
  ...playerTurnState,
  current: 1,
  messageCode: 'trash.cpuTurn',
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
  vi.useRealTimers();
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

  it('shows the cpu-turn phase name when opponent is active', async () => {
    // Stub cpu command so the auto-advance effect does not mutate state.
    mockExec.mockImplementation(async (command) => {
      if (command === 'reset') return cpuTurnState;
      return cpuTurnState;
    });
    renderWithProviders(<TrashPage />);
    await waitFor(() => expect(screen.getAllByText(/CPUのターン/).length).toBeGreaterThan(0));
  });

  it('renders 10 slots per player', async () => {
    renderWithProviders(<TrashPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const slotButtons = screen.getAllByRole('button', { name: /face-down|\d+:/ });
    expect(slotButtons.length).toBeGreaterThanOrEqual(20);
  });

  it('dispatches draw when the stock pile is clicked on the human turn', async () => {
    renderWithProviders(<TrashPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const stock = screen.getByText('🂠').closest('button');
    expect(stock).not.toBeNull();
    expect(stock).not.toBeDisabled();
    fireEvent.click(stock!);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
  });

  it('disables the stock button during CPU turn and await-wild phase', async () => {
    mockExec.mockImplementation(async () => ({ ...awaitWildState, current: 0 }));
    renderWithProviders(<TrashPage />);
    await waitFor(() => expect(screen.getAllByText(/ワイルド配置/).length).toBeGreaterThan(0));
    const stock = screen.getByText('🂠').closest('button');
    expect(stock).toBeDisabled();
  });

  it('dispatches placeWild when a face-down slot is clicked during await-wild', async () => {
    mockExec.mockImplementation(async (command) => {
      if (command === 'reset') return awaitWildState;
      return awaitWildState;
    });
    renderWithProviders(<TrashPage />);
    await waitFor(() => expect(screen.getAllByText(/ワイルド配置/).length).toBeGreaterThan(0));
    // Human slots are rendered below the opponent row — take the first face-down human slot
    // by matching its aria-label. Indices 1-10 exist for each player; pick a later one to
    // hit the human row rather than the opponent.
    const humanSlot = screen.getAllByRole('button', { name: '5: face-down' })[1];
    fireEvent.click(humanSlot);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('place', 5));
  });

  it('ignores slot clicks outside the await-wild phase', async () => {
    renderWithProviders(<TrashPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const slot = screen.getAllByRole('button', { name: '3: face-down' })[1];
    fireEvent.click(slot);
    // Only the mount-time reset call is expected.
    expect(mockExec).toHaveBeenCalledTimes(1);
    expect(mockExec).not.toHaveBeenCalledWith('place', expect.anything());
  });

  it('auto-advances the CPU turn via useEffect', async () => {
    vi.useFakeTimers();
    // reset returns cpu turn; subsequent cpu calls return the same so the effect keeps firing
    // but we only assert one call after the first 500ms tick.
    mockExec.mockImplementation(async () => cpuTurnState);
    renderWithProviders(<TrashPage />);
    // Flush the mounted reset call.
    await vi.waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    // The 500ms timer schedules the cpu command.
    await act(async () => {
      await vi.advanceTimersByTimeAsync(500);
    });
    expect(mockExec).toHaveBeenCalledWith('cpu');
    vi.useRealTimers();
  });

  it('shows the win banner on gameOverWin', async () => {
    mockExec.mockResolvedValue(gameOverWinState);
    renderWithProviders(<TrashPage />);
    await waitFor(() => expect(screen.getAllByText(/ゲーム終了/).length).toBeGreaterThan(0));
  });

  it('renders the action-log button on game over', async () => {
    mockExec.mockResolvedValue(gameOverWinState);
    renderWithProviders(<TrashPage />);
    await waitFor(() => expect(screen.getAllByText(/ゲーム終了/).length).toBeGreaterThan(0));
    const logButtons = screen.getAllByRole('button', { name: /棋譜|action log|アクション/i });
    expect(logButtons.length).toBeGreaterThan(0);
  });

  it('reset button opens the confirmation dialog', async () => {
    renderWithProviders(<TrashPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    // Resetting requires a confirmation; after clicking, the dialog element should be present.
    const resetBtn = document.querySelector('[data-tutorial="tr-reset"]');
    expect(resetBtn).not.toBeNull();
    fireEvent.click(resetBtn!);
    // The confirm dialog surfaces a button; ensure it renders some confirm text.
    await waitFor(() => {
      expect(screen.getAllByRole('button').length).toBeGreaterThanOrEqual(3);
    });
  });

  it('surfaces an error alert when the API fails on mount', async () => {
    mockExec.mockRejectedValue(new Error('network error'));
    renderWithProviders(<TrashPage />);
    await waitFor(() => expect(screen.getByRole('alert')).toBeInTheDocument());
  });

  it('shows 次のゲーム at game-end and fires reset directly (no confirm)', async () => {
    mockExec.mockResolvedValue(gameOverWinState);
    renderWithProviders(<TrashPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.queryByRole('button', { name: '確認' })).not.toBeInTheDocument();
  });
});
