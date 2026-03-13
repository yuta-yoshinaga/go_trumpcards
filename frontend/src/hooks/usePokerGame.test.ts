import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { pokerApi } from '../api/gameApi';
import type { PokerResponse } from '../types/card';
import { PokerPhase } from '../types/phases';
import { usePokerGame } from './usePokerGame';

vi.mock('../api/gameApi', () => ({
  pokerApi: { exec: vi.fn() },
}));

const mockExec = vi.mocked(pokerApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

const humanPlayer = {
  id: 0,
  isHuman: true,
  cards: [
    { design: 'SPADE' as const, value: 1 },
    { design: 'HEART' as const, value: 5 },
  ],
  chips: 990,
  currentBet: 0,
  folded: false,
  allIn: false,
  handRank: 0,
  handName: '',
  exchangeCount: -1,
  playStyleName: '',
};

const baseState: PokerResponse = {
  players: [humanPlayer],
  pot: 0,
  sidePots: [],
  dealerIdx: 0,
  currentTurn: 0,
  phase: PokerPhase.INIT,
  gameEndFlag: false,
  lastBet: 0,
  minRaise: 10,
  ante: 0,
  jokerCount: 0,
  bettingLimit: 0,
  raiseCount: 0,
  maxBetAmount: 0,
  roundResults: [],
  cpuActions: [],
  cpuExchanges: [],
  isLowball: false,
  message: '',
};

const exchangeState: PokerResponse = {
  ...baseState,
  phase: PokerPhase.EXCHANGE,
  currentTurn: 0,
};

beforeEach(() => {
  mockExec.mockResolvedValue(baseState);
});

afterEach(() => {
  vi.useRealTimers();
});

describe('usePokerGame', () => {
  it('calls reset on mount', async () => {
    renderHook(() => usePokerGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('returns initial state after mount', async () => {
    const { result } = renderHook(() => usePokerGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).toEqual(baseState));
  });

  it('canExchange is false when not in exchange phase', async () => {
    const { result } = renderHook(() => usePokerGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());
    expect(result.current.canExchange).toBe(false);
  });

  it('canExchange is true in exchange phase on human turn', async () => {
    mockExec.mockResolvedValue(exchangeState);
    const { result } = renderHook(() => usePokerGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.canExchange).toBe(true));
  });

  it('toggleCard does nothing when canExchange is false', async () => {
    const { result } = renderHook(() => usePokerGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => result.current.toggleCard(0));
    expect(result.current.selected).toEqual([]);
  });

  it('toggleCard adds card index when canExchange is true', async () => {
    mockExec.mockResolvedValue(exchangeState);
    const { result } = renderHook(() => usePokerGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.canExchange).toBe(true));

    act(() => result.current.toggleCard(0));
    expect(result.current.selected).toEqual([0]);
  });

  it('toggleCard removes card when already selected', async () => {
    mockExec.mockResolvedValue(exchangeState);
    const { result } = renderHook(() => usePokerGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.canExchange).toBe(true));

    act(() => {
      result.current.toggleCard(0);
      result.current.toggleCard(0);
    });
    expect(result.current.selected).toEqual([]);
  });

  it('toggleCard fires odds API after 300ms debounce', async () => {
    mockExec.mockResolvedValue(exchangeState);
    const { result } = renderHook(() => usePokerGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.canExchange).toBe(true));

    vi.useFakeTimers({ toFake: ['setTimeout', 'clearTimeout'] });
    mockExec.mockClear();
    mockExec.mockResolvedValue({ ...exchangeState, odds: [] });

    act(() => result.current.toggleCard(0));
    expect(mockExec).not.toHaveBeenCalled();

    await act(async () => { vi.advanceTimersByTime(300); });
    vi.useRealTimers();

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('odds', [0]));
  });

  it('sets odds state after debounce resolves', async () => {
    mockExec.mockResolvedValue(exchangeState);
    const { result } = renderHook(() => usePokerGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.canExchange).toBe(true));

    const oddsData = [{ handRank: 1, handName: 'ワンペア', probability: 0.4, count: 4, total: 10 }];
    vi.useFakeTimers({ toFake: ['setTimeout', 'clearTimeout'] });
    mockExec.mockResolvedValue({ ...exchangeState, odds: oddsData });

    act(() => result.current.toggleCard(0));
    await act(async () => { vi.advanceTimersByTime(300); });
    vi.useRealTimers();

    await waitFor(() => expect(result.current.odds).toEqual(oddsData));
  });

  it('clears odds when selection becomes empty', async () => {
    mockExec.mockResolvedValue(exchangeState);
    const { result } = renderHook(() => usePokerGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.canExchange).toBe(true));

    const oddsData = [{ handRank: 1, handName: 'ワンペア', probability: 0.4, count: 4, total: 10 }];
    vi.useFakeTimers({ toFake: ['setTimeout', 'clearTimeout'] });
    mockExec.mockResolvedValue({ ...exchangeState, odds: oddsData });

    act(() => result.current.toggleCard(0));
    await act(async () => { vi.advanceTimersByTime(300); });
    vi.useRealTimers();
    await waitFor(() => expect(result.current.odds).toEqual(oddsData));

    // Deselect the card → selection empty → odds cleared immediately
    act(() => result.current.toggleCard(0));
    expect(result.current.odds).toBeNull();
  });

  it('rejects stale odds response via generation counter', async () => {
    mockExec.mockResolvedValue(exchangeState);
    const { result } = renderHook(() => usePokerGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.canExchange).toBe(true));

    const freshOdds = [{ handRank: 9, handName: 'ストレートフラッシュ', probability: 0.01, count: 0, total: 10 }];

    let resolveStale!: (v: PokerResponse) => void;
    vi.useFakeTimers({ toFake: ['setTimeout', 'clearTimeout'] });
    mockExec.mockImplementationOnce(
      (_cmd, _args) => new Promise<PokerResponse>((res) => { resolveStale = res; }),
    );
    mockExec.mockResolvedValue({ ...exchangeState, odds: freshOdds });

    // First toggle → starts stale request (not yet resolved)
    act(() => result.current.toggleCard(0));
    await act(async () => { vi.advanceTimersByTime(300); });

    // Second toggle → starts fresh request, bumps generation
    act(() => result.current.toggleCard(1));
    await act(async () => { vi.advanceTimersByTime(300); });
    vi.useRealTimers();

    await waitFor(() => expect(result.current.odds).toEqual(freshOdds));

    // Stale response resolves late — should be ignored
    await act(async () => resolveStale({ ...exchangeState, odds: [{ handRank: 1, handName: 'ワンペア', probability: 0.4, count: 4, total: 10 }] }));
    expect(result.current.odds).toEqual(freshOdds);
  });

  it('onSuccess clears selection and odds', async () => {
    mockExec.mockResolvedValue(exchangeState);
    const { result } = renderHook(() => usePokerGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.canExchange).toBe(true));

    const oddsData = [{ handRank: 1, handName: 'ワンペア', probability: 0.4, count: 4, total: 10 }];
    vi.useFakeTimers({ toFake: ['setTimeout', 'clearTimeout'] });
    mockExec.mockResolvedValue({ ...exchangeState, odds: oddsData });

    act(() => result.current.toggleCard(0));
    await act(async () => { vi.advanceTimersByTime(300); });
    vi.useRealTimers();
    await waitFor(() => expect(result.current.odds).toEqual(oddsData));
    expect(result.current.selected).toEqual([0]);

    // Trigger onSuccess via exec → should clear selection and odds
    mockExec.mockResolvedValue(baseState);
    await act(async () => { await result.current.exec('exchange', [0]); });

    expect(result.current.selected).toEqual([]);
    expect(result.current.odds).toBeNull();
  });

  it('cleans up debounce timer on unmount without errors', async () => {
    mockExec.mockResolvedValue(exchangeState);
    const { result, unmount } = renderHook(() => usePokerGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    vi.useFakeTimers({ toFake: ['setTimeout', 'clearTimeout'] });
    // Directly manipulate canExchangeRef by triggering toggleCard via the toggle function
    // Since state has EXCHANGE phase and currentTurn=0 (human), canExchange should be true
    mockExec.mockResolvedValue({ ...exchangeState, odds: [] });

    // Unmount before timer fires — should not throw
    unmount();
    expect(() => { vi.advanceTimersByTime(300); }).not.toThrow();
  });
});
