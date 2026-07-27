import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { describe, expect, it, vi } from 'vitest';
import { bridgeApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import type { BridgeResponse } from '../types/card';
import { DEFAULT_BRIDGE_CONFIG, useBridgeGame } from './useBridgeGame';

vi.mock('../api/gameApi', () => ({
  bridgeApi: { exec: vi.fn() },
}));

const mockExec = vi.mocked(bridgeApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

const baseBridgeState: BridgeResponse = {
  players: [
    { id: 0, isHuman: true, cardCount: 13, cards: [], team: 0, trickCount: 0 },
    { id: 1, isHuman: false, cardCount: 13, cards: [], team: 1, trickCount: 0 },
    { id: 2, isHuman: false, cardCount: 13, cards: [], team: 0, trickCount: 0 },
    { id: 3, isHuman: false, cardCount: 13, cards: [], team: 1, trickCount: 0 },
  ],
  phase: 0,
  roundNumber: 1,
  trickNumber: 0,
  currentPlayerIdx: 0,
  bidPlayerIdx: 0,
  dealerIdx: 0,
  trumpSuit: 0,
  contractLevel: 0,
  contractSuit: 0,
  doubled: 0,
  declarerIdx: -1,
  dummyIdx: -1,
  bidHistory: [],
  vulnerability: [false, false],
  currentTrick: [],
  teamScores: [0, 0],
  gamesWon: [0, 0],
  belowLine: [0, 0],
  gameEndFlag: false,
  winnerTeam: -1,
  leadPlayerIdx: -1,
  openingLeadDone: false,
  dummyHand: null,
  message: '',
  config: { cpuDifficulty: 1 },
};

describe('useBridgeGame', () => {
  it('calls reset on mount', async () => {
    mockExec.mockResolvedValue(baseBridgeState);
    renderHook(() => useBridgeGame(), { wrapper: createWrapper() });
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, undefined, undefined, DEFAULT_BRIDGE_CONFIG),
    );
  });

  it('handleBid calls apiExec with bid command', async () => {
    mockExec.mockResolvedValue(baseBridgeState);
    const { result } = renderHook(() => useBridgeGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(baseBridgeState);
    act(() => result.current.handleBid(1, 1, 1));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', undefined, 1, 1, 1));
  });

  it('handleBid with pass calls apiExec correctly', async () => {
    mockExec.mockResolvedValue(baseBridgeState);
    const { result } = renderHook(() => useBridgeGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(baseBridgeState);
    act(() => result.current.handleBid(0));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', undefined, 0, undefined, undefined));
  });

  it('handlePlay does nothing if no card selected', async () => {
    mockExec.mockResolvedValue(baseBridgeState);
    const { result } = renderHook(() => useBridgeGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    act(() => result.current.handlePlay());
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('handlePlay calls apiExec with selected card', async () => {
    mockExec.mockResolvedValue(baseBridgeState);
    const { result } = renderHook(() => useBridgeGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => result.current.toggleCard(3));
    mockExec.mockClear();
    mockExec.mockResolvedValue(baseBridgeState);
    act(() => result.current.handlePlay());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 3));
  });

  it('handleNextTrick calls apiExec with next', async () => {
    mockExec.mockResolvedValue(baseBridgeState);
    const { result } = renderHook(() => useBridgeGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(baseBridgeState);
    act(() => result.current.handleNextTrick());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('handleNextRound calls apiExec with nextround', async () => {
    mockExec.mockResolvedValue(baseBridgeState);
    const { result } = renderHook(() => useBridgeGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(baseBridgeState);
    act(() => result.current.handleNextRound());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('handleHint fetches hint', async () => {
    const hintState = { ...baseBridgeState, hint: { cardIndex: 2, reason: 'follow_suit' } };
    mockExec.mockResolvedValue(baseBridgeState);
    const { result } = renderHook(() => useBridgeGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockResolvedValue(hintState);
    await act(async () => result.current.handleHint());
    expect(result.current.hint).toEqual({ cardIndex: 2, reason: 'follow_suit' });
  });

  it('handleHint sets error on failure', async () => {
    mockExec.mockResolvedValue(baseBridgeState);
    const { result } = renderHook(() => useBridgeGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockRejectedValue(new Error('network'));
    await act(async () => result.current.handleHint());
    expect(result.current.hintError).toBeTruthy();
  });

  it('clearSelection clears hint on success', async () => {
    const hintState = { ...baseBridgeState, hint: { cardIndex: 2, reason: 'follow_suit' } };
    mockExec.mockResolvedValue(baseBridgeState);
    const { result } = renderHook(() => useBridgeGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockResolvedValue(hintState);
    await act(async () => result.current.handleHint());
    expect(result.current.hint).not.toBeNull();

    mockExec.mockResolvedValue(baseBridgeState);
    act(() => result.current.handleBid(0));
    await waitFor(() => expect(result.current.hint).toBeNull());
  });
});
