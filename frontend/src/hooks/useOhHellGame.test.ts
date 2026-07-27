import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { ohHellApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import type { OhHellResponse } from '../types/card';
import { DEFAULT_OH_HELL_CONFIG, useOhHellGame } from './useOhHellGame';

vi.mock('../api/gameApi', () => ({
  ohHellApi: { exec: vi.fn() },
}));

const mockExec = vi.mocked(ohHellApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

/** Minimal response — only the fields these tests read. */
const baseState = { players: [], phase: 0, message: '' } as unknown as OhHellResponse;

async function mounted() {
  const rendered = renderHook(() => useOhHellGame(), { wrapper: createWrapper() });
  await waitFor(() => expect(mockExec).toHaveBeenCalled());
  mockExec.mockClear();
  mockExec.mockResolvedValue(baseState);
  return rendered;
}

describe('useOhHellGame', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(baseState);
  });

  it('resets with the default config on mount', async () => {
    renderHook(() => useOhHellGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, DEFAULT_OH_HELL_CONFIG));
  });

  it('submits a bid', async () => {
    const { result } = await mounted();
    act(() => result.current.handleBid(3));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', 3));
  });

  it('submits a zero bid, which is a legal call rather than a no-op', async () => {
    const { result } = await mounted();
    act(() => result.current.handleBid(0));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', 0));
  });

  it('plays only when exactly one card is selected', async () => {
    const { result } = await mounted();
    act(() => result.current.handlePlay());
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();

    act(() => result.current.toggleCard(2));
    act(() => result.current.handlePlay());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', undefined, 2));
  });

  it('advances the trick and the round', async () => {
    const { result } = await mounted();
    act(() => result.current.handleNextTrick());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
    act(() => result.current.handleNextRound());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('stores the hint the server returns and clears any error', async () => {
    const hint = { cardIndices: [1], reason: 'follow suit' };
    const { result } = await mounted();
    mockExec.mockResolvedValue({ ...baseState, hint } as unknown as OhHellResponse);
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
    mockExec.mockResolvedValue({ ...baseState, hint: { cardIndices: [0], reason: 'x' } } as unknown as OhHellResponse);
    await act(async () => {
      await result.current.handleHint();
    });
    expect(result.current.hint).not.toBeNull();
    mockExec.mockResolvedValue(baseState);
    await act(async () => {
      await result.current.exec('next');
    });
    expect(result.current.hint).toBeNull();
  });
});
