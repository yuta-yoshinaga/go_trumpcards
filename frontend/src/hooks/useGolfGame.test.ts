import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { golfApi } from '../api/gameApi';
import type { GolfResponse } from '../types/card';
import { useGolfGame } from './useGolfGame';

vi.mock('../api/gameApi', () => ({
  golfApi: { exec: vi.fn() },
}));

const mockExec = vi.mocked(golfApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

const defaultState: GolfResponse = {
  layout: [],
  stockCount: 16,
  waste: [{ design: 'CLOVER', value: 4 }],
  phase: 0,
  moveCount: 0,
  canUndo: false,
  isStalemate: false,
  message: '',
};

beforeEach(() => {
  mockExec.mockResolvedValue(defaultState);
});

describe('useGolfGame', () => {
  it('calls reset on mount', async () => {
    renderHook(() => useGolfGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('returns initial state after mount', async () => {
    const { result } = renderHook(() => useGolfGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).toEqual(defaultState));
  });

  it('handleDraw dispatches draw command', async () => {
    const { result } = renderHook(() => useGolfGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleDraw();
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
  });

  it('handleReset dispatches reset command', async () => {
    const { result } = renderHook(() => useGolfGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleReset();
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('handleGiveUp dispatches giveup command', async () => {
    const { result } = renderHook(() => useGolfGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleGiveUp();
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('handleUndo dispatches undo command', async () => {
    const { result } = renderHook(() => useGolfGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleUndo();
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo'));
  });

  it('handleSelectCard dispatches remove command with col', async () => {
    const { result } = renderHook(() => useGolfGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleSelectCard(3);
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('remove', 3));
  });

  it('handleHint calls exec with hint and sets hint state', async () => {
    const hintResponse: GolfResponse = {
      ...defaultState,
      hint: { type: 'remove', col: 2 },
    };
    const { result } = renderHook(() => useGolfGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockResolvedValue(hintResponse);
    await act(async () => {
      await result.current.handleHint();
    });

    expect(result.current.hint).toEqual(hintResponse.hint);
  });

  it('handleHint sets hint to null when no hint returned', async () => {
    const { result } = renderHook(() => useGolfGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockResolvedValue({ ...defaultState, hint: undefined });
    await act(async () => {
      await result.current.handleHint();
    });

    expect(result.current.hint).toBeNull();
  });

  it('handleHint sets hintError on failure', async () => {
    const { result } = renderHook(() => useGolfGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockRejectedValue(new Error('Network error'));
    await act(async () => {
      await result.current.handleHint();
    });

    expect(result.current.hintError).toBeTruthy();
    expect(result.current.hint).toBeNull();
  });

  it('handleDraw clears hint', async () => {
    const hintResponse: GolfResponse = {
      ...defaultState,
      hint: { type: 'remove', col: 2 },
    };
    const { result } = renderHook(() => useGolfGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockResolvedValue(hintResponse);
    await act(async () => {
      await result.current.handleHint();
    });
    expect(result.current.hint).not.toBeNull();

    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleDraw();
    });
    expect(result.current.hint).toBeNull();
  });
});
