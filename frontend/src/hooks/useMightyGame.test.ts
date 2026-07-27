import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { mightyApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import type { MightyResponse } from '../types/card';
import { DEFAULT_MIGHTY_CONFIG, useMightyGame } from './useMightyGame';

vi.mock('../api/gameApi', () => ({
  mightyApi: { exec: vi.fn() },
}));

const mockExec = vi.mocked(mightyApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

/** Minimal response — only the fields these tests read. */
const baseState = { players: [], phase: 0, message: '' } as unknown as MightyResponse;

async function mounted() {
  const rendered = renderHook(() => useMightyGame(), { wrapper: createWrapper() });
  await waitFor(() => expect(mockExec).toHaveBeenCalled());
  mockExec.mockClear();
  mockExec.mockResolvedValue(baseState);
  return rendered;
}

describe('useMightyGame', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(baseState);
  });

  // Mighty's command signature is wide, so the config rides in the 10th slot. Pinning
  // the position matters: a signature change that silently shifted it would leave the
  // game resetting with server defaults instead of the configured ones.
  it('resets with the default config on mount', async () => {
    renderHook(() => useMightyGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    const resetCall = mockExec.mock.calls.find((call) => call[0] === 'reset');
    expect(resetCall).toBeDefined();
    expect(resetCall?.[9]).toEqual(DEFAULT_MIGHTY_CONFIG);
  });

  it('submits a bid, carrying the no-trump flag', async () => {
    const { result } = await mounted();
    act(() => result.current.handleBid(15, true));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', 15, true));
  });

  it('passes by bidding zero', async () => {
    const { result } = await mounted();
    act(() => result.current.handlePass());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', 0, false));
  });

  it('declares trump together with the friend card', async () => {
    const { result } = await mounted();
    act(() => result.current.handleTrumpAndFriend(1, 3, 14));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('trump', undefined, undefined, undefined, 1, 3, 14));
  });

  // Mighty's declarer discards exactly three cards; anything else must not be sent.
  it('exchanges only when exactly three discards are given', async () => {
    const { result } = await mounted();
    act(() => result.current.handleExchange([1, 2]));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();

    act(() => result.current.handleExchange([1, 2, 3]));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith(
        'exchange',
        undefined,
        undefined,
        undefined,
        undefined,
        undefined,
        undefined,
        [1, 2, 3],
      ),
    );
  });

  it('plays only when exactly one card is selected', async () => {
    const { result } = await mounted();
    act(() => result.current.handlePlay());
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();

    act(() => result.current.toggleCard(0));
    act(() => result.current.handlePlay());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', undefined, undefined, 0));
  });

  // Leading the joker names the suit everyone must follow, so the suit has to reach
  // the server alongside the card.
  it('leads the joker with the nominated suit, and only with a card selected', async () => {
    const { result } = await mounted();
    act(() => result.current.handleJokerLead(2));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();

    act(() => result.current.toggleCard(6));
    act(() => result.current.handleJokerLead(2));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith(
        'jokerlead',
        undefined,
        undefined,
        6,
        undefined,
        undefined,
        undefined,
        undefined,
        2,
      ),
    );
  });

  it('stores the hint the server returns and clears any error', async () => {
    const hint = { cardIndices: [1], reason: 'follow suit' };
    const { result } = await mounted();
    mockExec.mockResolvedValue({ ...baseState, hint } as unknown as MightyResponse);
    await act(async () => {
      await result.current.handleHint();
    });
    expect(mockExec).toHaveBeenCalledWith('hint');
    expect(result.current.hint).toEqual(hint);
    expect(result.current.hintError).toBeNull();
    expect(result.current.hintLoading).toBe(false);
  });

  it('normalises a response with no hint to null', async () => {
    const { result } = await mounted();
    await act(async () => {
      await result.current.handleHint();
    });
    expect(result.current.hint).toBeNull();
  });

  it('reports a failed hint request without stranding hintLoading', async () => {
    const { result } = await mounted();
    mockExec.mockRejectedValue(new Error('network'));
    await act(async () => {
      await result.current.handleHint();
    });
    expect(result.current.hintError).toEqual(expect.any(String));
    expect(result.current.hintLoading).toBe(false);
  });

  it('clears a stale hint once a command succeeds', async () => {
    const { result } = await mounted();
    mockExec.mockResolvedValue({ ...baseState, hint: { cardIndices: [0], reason: 'x' } } as unknown as MightyResponse);
    await act(async () => {
      await result.current.handleHint();
    });
    expect(result.current.hint).not.toBeNull();
    mockExec.mockResolvedValue(baseState);
    await act(async () => {
      await result.current.apiCall('next');
    });
    expect(result.current.hint).toBeNull();
  });
});
