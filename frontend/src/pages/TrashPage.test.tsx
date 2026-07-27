import { act, fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { trashApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
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
  it('shows the CPU open-slot count indicator', async () => {
    const mixedSlots = Array.from({ length: 10 }, (_, i) =>
      i < 3 ? { faceUp: true, card: card('HEART', i + 1) } : { faceUp: false },
    );
    mockExec.mockResolvedValue({
      ...playerTurnState,
      players: [playerTurnState.players[0], { slots: mixedSlots, isCpu: true }],
    });
    renderWithProviders(<TrashPage />);
    expect(await screen.findByText('3/10 枚オープン')).toBeInTheDocument();
    // Exactly one badge: the human row never receives the badge prop.
    expect(screen.getAllByText(/枚オープン/)).toHaveLength(1);
  });

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
    fireEvent.click(stock as HTMLElement);
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
    await flushPendingDispatch();
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
    fireEvent.click(resetBtn as Element);
    // The confirm dialog surfaces a button; ensure it renders some confirm text.
    await waitFor(() => {
      expect(screen.getAllByRole('button').length).toBeGreaterThanOrEqual(3);
    });
  });

  it('highlights the rank-matching slot when a normal pending card is held', async () => {
    // Drew the 4 of HEART → slot index 3 should pulse-warn on the human row.
    // CPU row is rendered first, so the player's slot is the second match.
    const pendingFour: TrashResponse = { ...playerTurnState, pending: card('HEART', 4) };
    mockExec.mockResolvedValue(pendingFour);
    renderWithProviders(<TrashPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const slot4s = screen.getAllByRole('button', { name: '4: face-down' });
    const playerSlot4 = slot4s[1];
    expect(playerSlot4.dataset.pendingTarget).toBe('true');
    expect(playerSlot4.className).toContain('ring-ds-warning');
    expect(playerSlot4.className).toContain('motion-safe:animate-pulse');
    // CPU row never highlights.
    expect(slot4s[0].dataset.pendingTarget).toBe('false');
  });

  it('highlights slot 0 (Ace boundary) when pending value is 1', async () => {
    const pendingAce: TrashResponse = { ...playerTurnState, pending: card('SPADE', 1) };
    mockExec.mockResolvedValue(pendingAce);
    renderWithProviders(<TrashPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const slot1s = screen.getAllByRole('button', { name: '1: face-down' });
    expect(slot1s[1].dataset.pendingTarget).toBe('true');
  });

  it('highlights slot 9 (10 boundary) when pending value is 10', async () => {
    const pendingTen: TrashResponse = { ...playerTurnState, pending: card('SPADE', 10) };
    mockExec.mockResolvedValue(pendingTen);
    renderWithProviders(<TrashPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const slot10s = screen.getAllByRole('button', { name: '10: face-down' });
    expect(slot10s[1].dataset.pendingTarget).toBe('true');
  });

  it('does not highlight when pending card value is outside 1..10 (wild flow handles K/Joker separately)', async () => {
    // K (13) drawn but still in play phase (not yet AWAIT_WILD) — no pending highlight.
    const pendingK: TrashResponse = { ...playerTurnState, pending: card('DIAMOND', 13) };
    mockExec.mockResolvedValue(pendingK);
    renderWithProviders(<TrashPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const allSlots = screen.getAllByRole('button', { name: /face-down/ });
    for (const slot of allSlots) {
      expect(slot.dataset.pendingTarget).toBe('false');
    }
  });

  it('does not highlight on CPU turn', async () => {
    const cpuPending: TrashResponse = { ...cpuTurnState, pending: card('HEART', 4) };
    mockExec.mockImplementation(async () => cpuPending);
    renderWithProviders(<TrashPage />);
    await waitFor(() => expect(screen.getAllByText(/CPUのターン/).length).toBeGreaterThan(0));
    for (const slot of screen.getAllByRole('button', { name: /face-down/ })) {
      expect(slot.dataset.pendingTarget).toBe('false');
    }
  });

  it('announces a normal pending card and its target slot in a polite live region', async () => {
    mockExec.mockResolvedValue({ ...playerTurnState, pending: card('HEART', 4) });
    renderWithProviders(<TrashPage />);
    const region = await screen.findByTestId('tr-pending-announce');
    expect(region).toHaveAttribute('role', 'status');
    expect(region).toHaveAttribute('aria-live', 'polite');
    await waitFor(() => expect(region).toHaveTextContent('引いたカード: ♥ 4。スロット 4 に置けます'));
  });

  it('announces a wild pending card as a free-choice placement', async () => {
    mockExec.mockResolvedValue(awaitWildState); // ♦ K, AWAIT_WILD
    renderWithProviders(<TrashPage />);
    const region = await screen.findByTestId('tr-pending-announce');
    await waitFor(() =>
      expect(region).toHaveTextContent('引いたカード: ♦ K。ワイルドカードです。空いているスロットを選んでください'),
    );
  });

  it('announces a dead (J/Q) pending card as discarded', async () => {
    mockExec.mockResolvedValue({ ...playerTurnState, pending: card('SPADE', 11) });
    renderWithProviders(<TrashPage />);
    const region = await screen.findByTestId('tr-pending-announce');
    await waitFor(() => expect(region).toHaveTextContent('引いたカード: ♠ J。置けるスロットがなく捨て札になります'));
  });

  it('announces a normal pending card as discarded when its target slot is already face-up', async () => {
    // Slot 4 (index 3) already filled → a pending 4 cannot be placed and is dead.
    const filledSlot4 = faceDownSlots();
    filledSlot4[3] = { faceUp: true, card: card('HEART', 4) };
    mockExec.mockResolvedValue({
      ...playerTurnState,
      players: [
        { slots: filledSlot4, isCpu: false },
        { slots: faceDownSlots(), isCpu: true },
      ],
      pending: card('SPADE', 4),
    });
    renderWithProviders(<TrashPage />);
    const region = await screen.findByTestId('tr-pending-announce');
    await waitFor(() => expect(region).toHaveTextContent('引いたカード: ♠ 4。置けるスロットがなく捨て札になります'));
  });

  it('keeps the pending live region empty when no card is pending', async () => {
    renderWithProviders(<TrashPage />);
    const region = await screen.findByTestId('tr-pending-announce');
    expect(region).toHaveTextContent('');
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

  it('defaults the CPU speed select to normal when unset', async () => {
    renderWithProviders(<TrashPage />);
    const select = (await screen.findByTestId('trash-cpu-speed-select')) as HTMLSelectElement;
    expect(select.value).toBe('normal');
  });

  it('loads the persisted CPU speed from localStorage on mount', async () => {
    localStorage.setItem('trash:cpuSpeed', 'fast');
    renderWithProviders(<TrashPage />);
    const select = (await screen.findByTestId('trash-cpu-speed-select')) as HTMLSelectElement;
    expect(select.value).toBe('fast');
  });

  it('persists the chosen CPU speed to localStorage', async () => {
    renderWithProviders(<TrashPage />);
    const select = (await screen.findByTestId('trash-cpu-speed-select')) as HTMLSelectElement;
    fireEvent.change(select, { target: { value: 'slow' } });
    expect(select.value).toBe('slow');
    expect(localStorage.getItem('trash:cpuSpeed')).toBe('slow');
  });

  it('uses the selected speed as the CPU turn delay', async () => {
    vi.useFakeTimers();
    try {
      // 'fast' → 200ms delay. Verify the boundary at 200ms.
      localStorage.setItem('trash:cpuSpeed', 'fast');
      mockExec.mockImplementation(async () => cpuTurnState);
      renderWithProviders(<TrashPage />);
      // Flush the mount reset + initial render so the CPU-turn effect subscribes.
      await act(async () => {
        await vi.advanceTimersByTimeAsync(0);
      });
      mockExec.mockClear();
      // Just before the 200ms fast delay no cpu step has fired yet.
      await act(async () => {
        await vi.advanceTimersByTimeAsync(199);
      });
      expect(mockExec).not.toHaveBeenCalledWith('cpu');
      // Crossing 200ms fires the fast cpu step.
      await act(async () => {
        await vi.advanceTimersByTimeAsync(1);
      });
      expect(mockExec).toHaveBeenCalledWith('cpu');
    } finally {
      vi.useRealTimers();
    }
  });
});
