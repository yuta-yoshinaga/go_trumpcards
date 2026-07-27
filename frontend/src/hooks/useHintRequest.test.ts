import { act, renderHook, waitFor } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { useHintRequest } from './useHintRequest';

interface Res {
  hint: { kind: string } | null;
}

function setup(overrides: Partial<Parameters<typeof useHintRequest<Res, { kind: string }>>[0]> = {}) {
  const setHint = vi.fn();
  const setHintError = vi.fn();
  const setHintLoading = vi.fn();
  const fetchHint = vi.fn(async () => ({ hint: { kind: 'play' } }) as Res);
  const { result, unmount, rerender } = renderHook(() =>
    useHintRequest<Res, { kind: string }>({
      fetchHint,
      selectHint: (res) => res.hint,
      setHint,
      setHintError,
      setHintLoading,
      ...overrides,
    }),
  );
  return { result, unmount, rerender, setHint, setHintError, setHintLoading, fetchHint };
}

describe('useHintRequest', () => {
  it('stores the selected hint and clears any previous error', async () => {
    const { result, setHint, setHintError } = setup();
    await act(async () => {
      await result.current();
    });
    expect(setHint).toHaveBeenCalledWith({ kind: 'play' });
    expect(setHintError).toHaveBeenCalledWith(null);
  });

  it('toggles the loading flag around the request when one is supplied', async () => {
    const { result, setHintLoading } = setup();
    await act(async () => {
      await result.current();
    });
    expect(setHintLoading).toHaveBeenNthCalledWith(1, true);
    expect(setHintLoading).toHaveBeenNthCalledWith(2, false);
  });

  it('works for games that do not track a loading flag', async () => {
    const { result, setHint } = setup({ setHintLoading: undefined });
    await act(async () => {
      await result.current();
    });
    expect(setHint).toHaveBeenCalledWith({ kind: 'play' });
  });

  it('normalises a missing hint to null', async () => {
    const { result, setHint } = setup({ fetchHint: async () => ({ hint: null }) as Res });
    await act(async () => {
      await result.current();
    });
    expect(setHint).toHaveBeenCalledWith(null);
  });

  it('reports a network error and leaves the hint alone', async () => {
    const { result, setHint, setHintError, setHintLoading } = setup({
      fetchHint: async () => {
        throw new Error('boom');
      },
    });
    await act(async () => {
      await result.current();
    });
    expect(setHint).not.toHaveBeenCalled();
    expect(setHintError).toHaveBeenCalledWith(expect.any(String));
    expect(setHintLoading).toHaveBeenLastCalledWith(false);
  });

  // The reason this hook exists rather than 20 hand-written copies: the response
  // lands after an await, so it can outlive the page. Writing state then is a silent
  // no-op until the environment is torn down, where React throws from
  // dispatchSetState. See #4444 / #4447.
  it('writes nothing once the component has unmounted', async () => {
    let resolveHint: ((value: Res) => void) | undefined;
    const { result, unmount, setHint, setHintError, setHintLoading } = setup({
      fetchHint: () =>
        new Promise<Res>((resolve) => {
          resolveHint = resolve;
        }),
    });
    act(() => {
      void result.current();
    });
    await waitFor(() => expect(setHintLoading).toHaveBeenCalledWith(true));
    setHintLoading.mockClear();

    unmount();
    resolveHint?.({ hint: { kind: 'too-late' } });
    await new Promise((resolve) => {
      setTimeout(resolve, 0);
    });

    expect(setHint).not.toHaveBeenCalled();
    expect(setHintError).not.toHaveBeenCalled();
    expect(setHintLoading).not.toHaveBeenCalled();
  });

  it('keeps a stable identity across renders so dependents are not invalidated', () => {
    const { result, rerender } = setup();
    const first = result.current;
    rerender();
    expect(result.current).toBe(first);
  });
});
