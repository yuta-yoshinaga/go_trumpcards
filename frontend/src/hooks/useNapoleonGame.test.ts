import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { napoleonApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import type { NapoleonResponse } from '../types/card';
import { DEFAULT_NAPOLEON_CONFIG, useNapoleonGame } from './useNapoleonGame';

vi.mock('../api/gameApi', () => ({
  napoleonApi: { exec: vi.fn() },
}));

const mockExec = vi.mocked(napoleonApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

/** Minimal response — only the fields these tests read. */
const baseState = { players: [], phase: 0, message: '' } as unknown as NapoleonResponse;

async function mounted() {
  const rendered = renderHook(() => useNapoleonGame(), { wrapper: createWrapper() });
  await waitFor(() => expect(mockExec).toHaveBeenCalled());
  mockExec.mockClear();
  mockExec.mockResolvedValue(baseState);
  return rendered;
}

describe('useNapoleonGame', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(baseState);
  });

  it('resets with the default config on mount', async () => {
    renderHook(() => useNapoleonGame(), { wrapper: createWrapper() });
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith(
        'reset',
        undefined,
        undefined,
        undefined,
        undefined,
        undefined,
        undefined,
        DEFAULT_NAPOLEON_CONFIG,
      ),
    );
  });

  it('submits a bid', async () => {
    const { result } = await mounted();
    act(() => result.current.handleBid(14));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', 14));
  });

  // Passing is the same command with a zero bid, which is worth pinning: a future
  // refactor introducing a separate 'pass' command would break the server contract.
  it('passes by bidding zero', async () => {
    const { result } = await mounted();
    act(() => result.current.handlePass());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', 0));
  });

  it('declares trump together with the adjutant card', async () => {
    const { result } = await mounted();
    act(() => result.current.handleTrumpDeclaration(1, 2, 13));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('trump', undefined, 1, 2, 13));
  });

  it('exchanges a chosen discard', async () => {
    const { result } = await mounted();
    act(() => result.current.handleExchange(4));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('exchange', undefined, undefined, undefined, undefined, 4),
    );
  });

  it('plays only when exactly one card is selected', async () => {
    const { result } = await mounted();
    act(() => result.current.handlePlay());
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();

    act(() => result.current.toggleCard(5));
    act(() => result.current.handlePlay());
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('play', undefined, undefined, undefined, undefined, undefined, 5),
    );
  });

  it('stores the hint the server returns and clears any error', async () => {
    const hint = { cardIndices: [1], reason: 'follow suit' };
    const { result } = await mounted();
    mockExec.mockResolvedValue({ ...baseState, hint } as unknown as NapoleonResponse);
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
    mockExec.mockResolvedValue({
      ...baseState,
      hint: { cardIndices: [0], reason: 'x' },
    } as unknown as NapoleonResponse);
    await act(async () => {
      await result.current.handleHint();
    });
    expect(result.current.hint).not.toBeNull();
    mockExec.mockResolvedValue(baseState);
    await act(async () => {
      await result.current.apiExec('next');
    });
    expect(result.current.hint).toBeNull();
  });
});
