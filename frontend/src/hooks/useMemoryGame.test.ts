import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { memoryApi } from '../api/gameApi';
import type { MemoryResponse } from '../types/card';
import { useMemoryGame } from './useMemoryGame';

vi.mock('../api/gameApi', () => ({
  memoryApi: { exec: vi.fn() },
}));

const mockExec = vi.mocked(memoryApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

const defaultState: MemoryResponse = {
  players: [
    { id: 0, isHuman: true, pairCount: 0, pairs: [] },
    { id: 1, isHuman: false, pairCount: 0, pairs: [] },
    { id: 2, isHuman: false, pairCount: 0, pairs: [] },
    { id: 3, isHuman: false, pairCount: 0, pairs: [] },
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
  mockExec.mockResolvedValue(defaultState);
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

  it('auto-next fires handleNext after delay during result phase', async () => {
    vi.useFakeTimers();
    localStorage.setItem('memory_auto_next_delay_ms', '1000');
    const resultState: MemoryResponse = { ...defaultState, phase: 2 };
    mockExec.mockResolvedValue(resultState);

    const { result } = renderHook(() => useMemoryGame(), { wrapper: createWrapper() });
    await vi.waitFor(() => expect(result.current.state).toEqual(resultState));

    mockExec.mockClear();
    mockExec.mockResolvedValue(resultState);
    await act(async () => {
      await vi.advanceTimersByTimeAsync(1000);
    });
    expect(mockExec).toHaveBeenCalledWith('next');

    localStorage.removeItem('memory_auto_next_delay_ms');
    vi.useRealTimers();
  });

  it('auto-next is disabled when delay is 0 (manual)', async () => {
    vi.useFakeTimers();
    localStorage.setItem('memory_auto_next_delay_ms', '0');
    const resultState: MemoryResponse = { ...defaultState, phase: 2 };
    mockExec.mockResolvedValue(resultState);

    const { result } = renderHook(() => useMemoryGame(), { wrapper: createWrapper() });
    await vi.waitFor(() => expect(result.current.state).toEqual(resultState));

    mockExec.mockClear();
    await act(async () => {
      await vi.advanceTimersByTimeAsync(3000);
    });
    expect(mockExec).not.toHaveBeenCalled();

    localStorage.removeItem('memory_auto_next_delay_ms');
    vi.useRealTimers();
  });

  it('auto-next does not fire when game has ended', async () => {
    vi.useFakeTimers();
    localStorage.setItem('memory_auto_next_delay_ms', '1000');
    const gameEndState: MemoryResponse = { ...defaultState, phase: 2, gameEndFlag: true };
    mockExec.mockResolvedValue(gameEndState);

    const { result } = renderHook(() => useMemoryGame(), { wrapper: createWrapper() });
    await vi.waitFor(() => expect(result.current.state).toEqual(gameEndState));

    mockExec.mockClear();
    await act(async () => {
      await vi.advanceTimersByTimeAsync(3000);
    });
    expect(mockExec).not.toHaveBeenCalled();

    localStorage.removeItem('memory_auto_next_delay_ms');
    vi.useRealTimers();
  });

  it('setAutoNextDelayMs persists the chosen delay', async () => {
    const { result } = renderHook(() => useMemoryGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.setAutoNextDelayMs(2000);
    });

    expect(result.current.autoNextDelayMs).toBe(2000);
    expect(localStorage.getItem('memory_auto_next_delay_ms')).toBe('2000');
    localStorage.removeItem('memory_auto_next_delay_ms');
  });
});
