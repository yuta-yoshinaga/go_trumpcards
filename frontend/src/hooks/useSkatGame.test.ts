import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { describe, expect, it, vi } from 'vitest';
import { skatApi } from '../api/gameApi';
import type { SkatResponse } from '../types/card';
import { SkatGameType } from '../types/phases';
import { DEFAULT_SKAT_CONFIG, useSkatGame } from './useSkatGame';

vi.mock('../api/gameApi', () => ({
  skatApi: { exec: vi.fn() },
}));

const mockExec = vi.mocked(skatApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

const baseSkatState: SkatResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 10,
      cards: [],
      bid: 0,
      isDeclarer: false,
      cardPoints: 0,
      roundsWon: 0,
      roundsLost: 0,
      roundScore: 0,
      cumulativeScore: 0,
      trickCount: 0,
    },
    {
      id: 1,
      isHuman: false,
      cardCount: 10,
      cards: [],
      bid: 0,
      isDeclarer: false,
      cardPoints: 0,
      roundsWon: 0,
      roundsLost: 0,
      roundScore: 0,
      cumulativeScore: 0,
      trickCount: 0,
    },
    {
      id: 2,
      isHuman: false,
      cardCount: 10,
      cards: [],
      bid: 0,
      isDeclarer: false,
      cardPoints: 0,
      roundsWon: 0,
      roundsLost: 0,
      roundScore: 0,
      cumulativeScore: 0,
      trickCount: 0,
    },
  ],
  phase: 0,
  roundNumber: 1,
  trickNumber: 0,
  currentPlayerIdx: -1,
  currentTrick: [],
  forehandIdx: 0,
  middlehandIdx: 1,
  rearhandIdx: 2,
  dealerIdx: 0,
  declarerIdx: -1,
  currentBid: 0,
  activeBidActorIdx: 0,
  gameType: 0,
  trumpSuit: 0,
  pickedSkat: false,
  declarerCardPoints: 0,
  defendersCardPoints: 0,
  winnerSide: -1,
  gameValue: 0,
  gameEndFlag: false,
  leadPlayerIdx: -1,
  message: '',
  config: { cpuDifficulty: 1, targetScore: 500 },
};

describe('useSkatGame', () => {
  it('calls reset on mount with default config', async () => {
    mockExec.mockResolvedValue(baseSkatState);
    renderHook(() => useSkatGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', { config: DEFAULT_SKAT_CONFIG }));
  });

  it('handleBid dispatches bid command', async () => {
    mockExec.mockResolvedValue(baseSkatState);
    const { result } = renderHook(() => useSkatGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(baseSkatState);
    act(() => result.current.handleBid(true));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { accept: true }));
  });

  it('handlePickSkat dispatches pickskat command', async () => {
    mockExec.mockResolvedValue(baseSkatState);
    const { result } = renderHook(() => useSkatGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(baseSkatState);
    act(() => result.current.handlePickSkat(false));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('pickskat', { pickup: false }));
  });

  it('handleDiscard does nothing if not exactly two cards selected', async () => {
    mockExec.mockResolvedValue(baseSkatState);
    const { result } = renderHook(() => useSkatGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    act(() => result.current.handleDiscard()); // 0 selected
    expect(mockExec).not.toHaveBeenCalled();

    act(() => result.current.toggleCard(0));
    mockExec.mockClear();
    act(() => result.current.handleDiscard()); // 1 selected
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('handleDiscard sends both indices when two cards are selected', async () => {
    mockExec.mockResolvedValue(baseSkatState);
    const { result } = renderHook(() => useSkatGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => result.current.toggleCard(0));
    act(() => result.current.toggleCard(2));
    mockExec.mockClear();
    mockExec.mockResolvedValue(baseSkatState);
    act(() => result.current.handleDiscard());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('discard', { discardA: 0, discardB: 2 }));
  });

  it('handleDeclareGame dispatches game with type and trumpSuit', async () => {
    mockExec.mockResolvedValue(baseSkatState);
    const { result } = renderHook(() => useSkatGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(baseSkatState);
    act(() => result.current.handleDeclareGame(SkatGameType.SUIT, 1));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('game', { gameType: SkatGameType.SUIT, trumpSuit: 1 }));
  });

  it('handleDeclareGame dispatches without trumpSuit for grand/null', async () => {
    mockExec.mockResolvedValue(baseSkatState);
    const { result } = renderHook(() => useSkatGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(baseSkatState);
    act(() => result.current.handleDeclareGame(SkatGameType.GRAND));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('game', { gameType: SkatGameType.GRAND, trumpSuit: undefined }),
    );
  });

  it('handlePlay does nothing if no card selected', async () => {
    mockExec.mockResolvedValue(baseSkatState);
    const { result } = renderHook(() => useSkatGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    act(() => result.current.handlePlay());
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('handlePlay sends selected cardIndex', async () => {
    mockExec.mockResolvedValue(baseSkatState);
    const { result } = renderHook(() => useSkatGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => result.current.toggleCard(3));
    mockExec.mockClear();
    mockExec.mockResolvedValue(baseSkatState);
    act(() => result.current.handlePlay());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 3 }));
  });

  it('handleNextTrick dispatches next', async () => {
    mockExec.mockResolvedValue(baseSkatState);
    const { result } = renderHook(() => useSkatGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(baseSkatState);
    act(() => result.current.handleNextTrick());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('handleNextRound dispatches nextround', async () => {
    mockExec.mockResolvedValue(baseSkatState);
    const { result } = renderHook(() => useSkatGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(baseSkatState);
    act(() => result.current.handleNextRound());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('handleHint stores returned hint and clears error', async () => {
    mockExec.mockResolvedValue(baseSkatState);
    const { result } = renderHook(() => useSkatGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    const accept = 1;
    mockExec.mockResolvedValueOnce({ ...baseSkatState, hint: { bid: accept, reason: 'strategic_bid' } });
    await act(async () => {
      await result.current.handleHint();
    });
    expect(result.current.hint).toEqual({ bid: accept, reason: 'strategic_bid' });
    expect(result.current.hintError).toBeNull();
  });

  it('handleHint sets hintError on rejection', async () => {
    mockExec.mockResolvedValue(baseSkatState);
    const { result } = renderHook(() => useSkatGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockRejectedValueOnce(new Error('boom'));
    await act(async () => {
      await result.current.handleHint();
    });
    expect(result.current.hintError).not.toBeNull();
  });

  it('handleConfigChange updates the config', async () => {
    mockExec.mockResolvedValue(baseSkatState);
    const { result } = renderHook(() => useSkatGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => result.current.handleConfigChange('cpuDifficulty', '2'));
    expect(result.current.skatConfig.cpuDifficulty).toBe(2);
  });

  it('reset dispatches reset with the current config', async () => {
    mockExec.mockResolvedValue(baseSkatState);
    const { result } = renderHook(() => useSkatGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => result.current.handleConfigChange('cpuDifficulty', '2'));
    mockExec.mockClear();
    act(() => result.current.reset());
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', { config: { cpuDifficulty: 2, targetScore: 500 } }),
    );
  });
});
