import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { describe, expect, it, vi } from 'vitest';
import { binokelApi } from '../api/gameApi';
import type { BinokelResponse } from '../types/card';
import { DEFAULT_BINOKEL_CONFIG, useBinokelGame } from './useBinokelGame';

vi.mock('../api/gameApi', () => ({
  binokelApi: { exec: vi.fn() },
}));

const mockExec = vi.mocked(binokelApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

const baseBinokelState: BinokelResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 15,
      cards: [],
      score: 0,
      trickCount: 0,
      bid: 0,
      hasPassed: false,
      meldScore: 0,
      trickPoints: 0,
    },
    {
      id: 1,
      isHuman: false,
      cardCount: 15,
      cards: [],
      score: 0,
      trickCount: 0,
      bid: 0,
      hasPassed: false,
      meldScore: 0,
      trickPoints: 0,
    },
    {
      id: 2,
      isHuman: false,
      cardCount: 15,
      cards: [],
      score: 0,
      trickCount: 0,
      bid: 0,
      hasPassed: false,
      meldScore: 0,
      trickPoints: 0,
    },
  ],
  phase: 0,
  roundNumber: 1,
  trickNumber: 0,
  currentPlayerIdx: 0,
  bidPlayerIdx: 0,
  dealerIdx: 2,
  trumpSuit: 0,
  highestBid: 0,
  highestBidder: -1,
  currentTrick: [],
  scores: [0, 0, 0],
  gameEndFlag: false,
  winnerPlayer: -1,
  leadPlayerIdx: -1,
  playerMelds: [[], [], []],
  dabb: [],
  dabbDiscarded: [],
  message: '',
  config: { cpuDifficulty: 1, pointLimit: 1500 },
};

describe('useBinokelGame', () => {
  it('calls reset with default config on mount', async () => {
    mockExec.mockResolvedValue(baseBinokelState);
    renderHook(() => useBinokelGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, DEFAULT_BINOKEL_CONFIG));
  });

  it('handleBid calls exec with bid command', async () => {
    mockExec.mockResolvedValue(baseBinokelState);
    const { result } = renderHook(() => useBinokelGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(baseBinokelState);
    act(() => result.current.handleBid(150));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', undefined, undefined, 150));
  });

  it('handlePass calls exec with pass command', async () => {
    mockExec.mockResolvedValue(baseBinokelState);
    const { result } = renderHook(() => useBinokelGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(baseBinokelState);
    act(() => result.current.handlePass());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('pass'));
  });

  it('handleDiscard calls exec with discard command and indices', async () => {
    mockExec.mockResolvedValue(baseBinokelState);
    const { result } = renderHook(() => useBinokelGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(baseBinokelState);
    act(() => result.current.handleDiscard([0, 1, 2]));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('discard', undefined, undefined, undefined, undefined, [0, 1, 2]),
    );
  });

  it('handleCallTrump calls exec with trump command', async () => {
    mockExec.mockResolvedValue(baseBinokelState);
    const { result } = renderHook(() => useBinokelGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(baseBinokelState);
    act(() => result.current.handleCallTrump(2));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('trump', undefined, undefined, undefined, 2));
  });

  it('handleConfirmMelds calls exec with meld command', async () => {
    mockExec.mockResolvedValue(baseBinokelState);
    const { result } = renderHook(() => useBinokelGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(baseBinokelState);
    act(() => result.current.handleConfirmMelds());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('meld'));
  });

  it('handlePlay calls exec with play command and card index', async () => {
    mockExec.mockResolvedValue(baseBinokelState);
    const { result } = renderHook(() => useBinokelGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(baseBinokelState);
    act(() => result.current.handlePlay(3));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 3));
  });

  it('handleNextTrick calls exec with next command', async () => {
    mockExec.mockResolvedValue(baseBinokelState);
    const { result } = renderHook(() => useBinokelGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(baseBinokelState);
    act(() => result.current.handleNextTrick());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('handleNextRound calls exec with nextround command', async () => {
    mockExec.mockResolvedValue(baseBinokelState);
    const { result } = renderHook(() => useBinokelGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(baseBinokelState);
    act(() => result.current.handleNextRound());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('handleHint fetches the server hint and stores it without touching game state', async () => {
    mockExec.mockResolvedValue(baseBinokelState);
    const { result } = renderHook(() => useBinokelGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    // The hint command returns only the bare hint object (top-level fields).
    mockExec.mockResolvedValue({ hint: { bidAmount: 25, reason: 'hint_bid' } } as unknown as BinokelResponse);
    await act(async () => {
      await result.current.handleHint();
    });

    expect(mockExec).toHaveBeenCalledWith('hint');
    expect(result.current.hint).toEqual({ bidAmount: 25, reason: 'hint_bid' });
    expect(result.current.hintError).toBeNull();
    // Game state is unaffected by a hint fetch.
    expect(result.current.state).toEqual(baseBinokelState);
  });

  it('handleHint ignores an empty (no-reason) hint response', async () => {
    mockExec.mockResolvedValue(baseBinokelState);
    const { result } = renderHook(() => useBinokelGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue({ message: 'ヒントなし' } as unknown as BinokelResponse);
    await act(async () => {
      await result.current.handleHint();
    });
    expect(result.current.hint).toBeNull();
  });

  it('handleHint sets hintError when the request fails', async () => {
    mockExec.mockResolvedValue(baseBinokelState);
    const { result } = renderHook(() => useBinokelGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockRejectedValueOnce(new Error('network'));
    await act(async () => {
      await result.current.handleHint();
    });
    expect(result.current.hintError).not.toBeNull();
    expect(result.current.hint).toBeNull();
  });

  it('clears the hint after a subsequent game action succeeds', async () => {
    mockExec.mockResolvedValue(baseBinokelState);
    const { result } = renderHook(() => useBinokelGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockResolvedValue({ hint: { suit: 1, reason: 'hint_trump' } } as unknown as BinokelResponse);
    await act(async () => {
      await result.current.handleHint();
    });
    expect(result.current.hint).not.toBeNull();

    mockExec.mockResolvedValue(baseBinokelState);
    act(() => result.current.handleCallTrump(1));
    await waitFor(() => expect(result.current.hint).toBeNull());
  });

  it('handleReset calls exec with reset command and current config', async () => {
    mockExec.mockResolvedValue(baseBinokelState);
    const { result } = renderHook(() => useBinokelGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(baseBinokelState);
    act(() => result.current.handleReset());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, DEFAULT_BINOKEL_CONFIG));
  });

  it('does not set state after unmount in handleHint on success', async () => {
    mockExec.mockResolvedValue(baseBinokelState);
    const { result, unmount } = renderHook(() => useBinokelGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    let resolvePromise: ((value: BinokelResponse) => void) | undefined;
    mockExec.mockReturnValue(
      new Promise((resolve) => {
        resolvePromise = resolve;
      }),
    );
    const hintPromise = result.current.handleHint();
    unmount();
    resolvePromise?.(baseBinokelState);
    await hintPromise;
  });

  it('does not set error after unmount in handleHint on failure', async () => {
    mockExec.mockResolvedValue(baseBinokelState);
    const { result, unmount } = renderHook(() => useBinokelGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    let rejectPromise: ((reason?: unknown) => void) | undefined;
    mockExec.mockReturnValue(
      new Promise((_, reject) => {
        rejectPromise = reject;
      }),
    );
    const hintPromise = result.current.handleHint();
    unmount();
    rejectPromise?.(new Error('network'));
    await hintPromise;
  });
});
