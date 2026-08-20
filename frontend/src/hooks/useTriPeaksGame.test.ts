import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { tripeaksApi } from '../api/gameApi';
import type { TriPeaksResponse } from '../types/card';
import { useTriPeaksGame } from './useTriPeaksGame';

vi.mock('../api/gameApi', () => ({
  tripeaksApi: { exec: vi.fn() },
}));

const mockExec = vi.mocked(tripeaksApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

const defaultState: TriPeaksResponse = {
  layout: [[{ card: null, removed: true, exposed: false }]],
  stockCount: 20,
  waste: [],
  phase: 0,
  moveCount: 0,
  score: 0,
  combo: 0,
  canUndo: false,
  isStalemate: false,
  message: '',
};

beforeEach(() => {
  mockExec.mockResolvedValue(defaultState);
});

describe('useTriPeaksGame', () => {
  it('calls reset on mount', async () => {
    renderHook(() => useTriPeaksGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('returns initial state after mount', async () => {
    const { result } = renderHook(() => useTriPeaksGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).toEqual(defaultState));
  });

  it('handleDraw dispatches draw and clears hint', async () => {
    const { result } = renderHook(() => useTriPeaksGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleDraw();
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
    expect(result.current.hint).toBeNull();
  });

  it('handleReset dispatches reset and clears hint', async () => {
    const { result } = renderHook(() => useTriPeaksGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleReset();
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(result.current.hint).toBeNull();
  });

  it('handleGiveUp dispatches giveup', async () => {
    const { result } = renderHook(() => useTriPeaksGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleGiveUp();
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('handleUndo dispatches undo', async () => {
    const { result } = renderHook(() => useTriPeaksGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleUndo();
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo'));
  });

  it('handleUndoEscape dispatches undo_n with count and clears hint', async () => {
    const { result } = renderHook(() => useTriPeaksGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleUndoEscape(5);
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo_n', undefined, undefined, 5));
    expect(result.current.hint).toBeNull();
  });

  it('handleHint calls exec with hint and sets hint state', async () => {
    const hintResponse: TriPeaksResponse = {
      ...defaultState,
      hint: { type: 'remove', row: 3, col: 0 },
    };
    const { result } = renderHook(() => useTriPeaksGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockResolvedValue(hintResponse);
    await act(async () => {
      await result.current.handleHint();
    });

    expect(result.current.hint).toEqual(hintResponse.hint);
  });

  it('handleHint sets hint to null when no hint returned', async () => {
    const { result } = renderHook(() => useTriPeaksGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockResolvedValue({ ...defaultState, hint: undefined });
    await act(async () => {
      await result.current.handleHint();
    });

    expect(result.current.hint).toBeNull();
  });

  it('handleHint sets hintError on failure', async () => {
    const { result } = renderHook(() => useTriPeaksGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockRejectedValue(new Error('Network error'));
    await act(async () => {
      await result.current.handleHint();
    });

    expect(result.current.hintError).toBeTruthy();
    expect(result.current.hint).toBeNull();
  });

  it('handleSelectCard forwards row/col to the remove command and clears hint', async () => {
    const { result } = renderHook(() => useTriPeaksGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    // Seed hint state so we can confirm it is cleared on selection.
    mockExec.mockResolvedValue({ ...defaultState, hint: { type: 'remove', row: 1, col: 2 } });
    await act(async () => {
      await result.current.handleHint();
    });
    expect(result.current.hint).not.toBeNull();

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleSelectCard(0, 2);
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('remove', 0, 2));
    await waitFor(() => expect(result.current.hint).toBeNull());
  });
});
