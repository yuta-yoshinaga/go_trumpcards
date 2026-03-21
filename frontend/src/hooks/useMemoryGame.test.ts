import { afterEach, beforeEach, describe, expect, it, vi } from 'bun:test';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { memoryApi } from '../api/gameApi';
import { asMocked } from '../test/viCompat';
import type { MemoryResponse } from '../types/card';
import { useMemoryGame } from './useMemoryGame';

let mockExec: ReturnType<typeof vi.fn>;
function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

const defaultState: MemoryResponse = {
  players: [
    { id: 0, isHuman: true, pairCount: 0 },
    { id: 1, isHuman: false, pairCount: 0 },
    { id: 2, isHuman: false, pairCount: 0 },
    { id: 3, isHuman: false, pairCount: 0 },
  ],
  board: Array.from({ length: 52 }, () => ({ card: null, faceUp: false, taken: false })),
  phase: 0,
  currentPlayerIdx: 0,
  firstFlipPos: -1,
  secondFlipPos: -1,
  lastMatchResult: false,
  gameEndFlag: false,
  winnerIdx: -1,
  turnNumber: 0,
  message: '',
  config: { cpuDifficulty: 1 },
};

beforeEach(() => {
  vi.spyOn(memoryApi, 'exec').mockImplementation(vi.fn());
  mockExec = asMocked(memoryApi.exec);
  mockExec.mockResolvedValue(defaultState);
});

afterEach(() => {
  vi.restoreAllMocks();
});

describe('useMemoryGame', () => {
  it('calls reset on mount', async () => {
    renderHook(() => useMemoryGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, { cpuDifficulty: 1 }));
  });

  it('returns initial state after mount', async () => {
    const { result } = renderHook(() => useMemoryGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).toEqual(defaultState));
  });

  it('handleFlip dispatches flip with position', async () => {
    const { result } = renderHook(() => useMemoryGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleFlip(5);
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('flip', 5));
  });

  it('handleNext dispatches next command', async () => {
    const { result } = renderHook(() => useMemoryGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleNext();
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('handleConfigChange updates config with valid number', async () => {
    const { result } = renderHook(() => useMemoryGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.handleConfigChange('cpuDifficulty', '2');
    });

    expect(result.current.memoryConfig.cpuDifficulty).toBe(2);
  });

  it('handleConfigChange ignores NaN values', async () => {
    const { result } = renderHook(() => useMemoryGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.handleConfigChange('cpuDifficulty', 'abc');
    });

    expect(result.current.memoryConfig.cpuDifficulty).toBe(1);
  });
});
