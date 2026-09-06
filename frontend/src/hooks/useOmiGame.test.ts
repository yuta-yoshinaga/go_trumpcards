import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { omiApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import type { OmiResponse } from '../types/card';
import { DEFAULT_OMI_CONFIG, useOmiGame } from './useOmiGame';

vi.mock('../api/gameApi', () => ({
  omiApi: { exec: vi.fn() },
}));

const mockExec = vi.mocked(omiApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

/** Minimal response — only the fields these tests read. */
const baseState = { players: [], phase: 0, message: '' } as unknown as OmiResponse;

async function mounted() {
  const rendered = renderHook(() => useOmiGame(), { wrapper: createWrapper() });
  await waitFor(() => expect(mockExec).toHaveBeenCalled());
  mockExec.mockClear();
  mockExec.mockResolvedValue(baseState);
  return rendered;
}

describe('useOmiGame', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(baseState);
  });

  it('resets with the default config on mount', async () => {
    renderHook(() => useOmiGame(), { wrapper: createWrapper() });
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, undefined, DEFAULT_OMI_CONFIG),
    );
  });

  // Omi: trump is called with `calltrump` + suit number (no goAlone)
  it('calls trump with the chosen suit', async () => {
    const { result } = await mounted();
    act(() => result.current.handleCallTrump(2));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('calltrump', undefined, 2));
  });

  it('plays only when exactly one card is selected', async () => {
    const { result } = await mounted();
    act(() => result.current.handlePlay());
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();

    act(() => result.current.toggleCard(1));
    act(() => result.current.handlePlay());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 1));
  });

  it('advances the trick and the round', async () => {
    const { result } = await mounted();
    act(() => result.current.handleNextTrick());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
    act(() => result.current.handleNextRound());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('stores the hint the server returns and clears any error', async () => {
    const hint = { cardIndex: 1, reason: 'call_strong_suit' };
    const { result } = await mounted();
    mockExec.mockResolvedValue({ ...baseState, hint } as unknown as OmiResponse);
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
      hint: { suit: 1, reason: 'call_strong_suit' },
    } as unknown as OmiResponse);
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
