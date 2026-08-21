import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { narcoticApi } from '../api/gameApi';
import type { NarcoticResponse } from '../types/card';
import { useNarcoticGame } from './useNarcoticGame';

vi.mock('../api/gameApi', () => ({
  narcoticApi: { exec: vi.fn() },
}));

const mockExec = vi.mocked(narcoticApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

const defaultState: NarcoticResponse = {
  columns: [[], [], [], []],
  stockCount: 44,
  discardCount: 4,
  redealCount: 0,
  phase: 0,
  moveCount: 0,
  canUndo: false,
  isStalemate: false,
  message: '',
};

beforeEach(() => {
  mockExec.mockResolvedValue(defaultState);
});

describe('useNarcoticGame', () => {
  it('calls reset on mount', async () => {
    renderHook(() => useNarcoticGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('returns initial state after mount', async () => {
    const { result } = renderHook(() => useNarcoticGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).toEqual(defaultState));
  });

  it('handleDraw dispatches draw command', async () => {
    const { result } = renderHook(() => useNarcoticGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => result.current.handleDraw());

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
  });

  it('handleReset dispatches reset command', async () => {
    const { result } = renderHook(() => useNarcoticGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => result.current.handleReset());

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('handleGiveUp dispatches giveup command', async () => {
    const { result } = renderHook(() => useNarcoticGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => result.current.handleGiveUp());

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('handleUndo dispatches undo command', async () => {
    const { result } = renderHook(() => useNarcoticGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => result.current.handleUndo());

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo'));
  });

  it('handleUndoEscape dispatches undo_n with count', async () => {
    const { result } = renderHook(() => useNarcoticGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => result.current.handleUndoEscape(3));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo_n', undefined, 3));
  });

  // **列を取らない。**揃った4枚をまとめて捨てるので、選ぶ余地が無い。
  // クローン元 (Aces Up) は列ごとに捨てるので `remove` に col が要った。
  it('handleRemove dispatches remove with no column', async () => {
    const { result } = renderHook(() => useNarcoticGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => result.current.handleRemove());

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('remove'));
  });

  // **クローン元には無い手。**山札が尽きても場を集めれば続けられる。
  it('handleRedeal dispatches redeal', async () => {
    const { result } = renderHook(() => useNarcoticGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => result.current.handleRedeal());

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('redeal'));
  });

  it('handleMove dispatches move command with col', async () => {
    const { result } = renderHook(() => useNarcoticGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => result.current.handleMove(1));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('move', 1));
  });

  it('handleHint sets hint state', async () => {
    const hintResponse: NarcoticResponse = { ...defaultState, hint: { type: 'move', col: 2 } };
    const { result } = renderHook(() => useNarcoticGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockResolvedValue(hintResponse);
    await act(async () => {
      await result.current.handleHint();
    });

    expect(result.current.hint).toEqual(hintResponse.hint);
  });

  it('handleHint sets hint to null when no hint returned', async () => {
    const { result } = renderHook(() => useNarcoticGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockResolvedValue({ ...defaultState, hint: undefined });
    await act(async () => {
      await result.current.handleHint();
    });

    expect(result.current.hint).toBeNull();
  });

  it('handleHint sets hintError on failure', async () => {
    const { result } = renderHook(() => useNarcoticGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockRejectedValue(new Error('Network error'));
    await act(async () => {
      await result.current.handleHint();
    });

    expect(result.current.hintError).toBeTruthy();
    expect(result.current.hint).toBeNull();
  });
});
