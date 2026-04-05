import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { describe, expect, it, vi } from 'vitest';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import { useGameApi } from './useGameApi';

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

describe('useGameApi', () => {
  it('has correct initial state', () => {
    const apiFn = vi.fn();
    const { result } = renderHook(() => useGameApi(apiFn), { wrapper: createWrapper() });
    expect(result.current.state).toBeNull();
    expect(result.current.loading).toBe(false);
    expect(result.current.error).toBeNull();
  });

  it('sets state on success', async () => {
    const apiFn = vi.fn().mockResolvedValue({ value: 42 });
    const { result } = renderHook(() => useGameApi(apiFn), { wrapper: createWrapper() });

    await act(async () => {
      await result.current.exec('arg1', 'arg2');
    });

    expect(apiFn).toHaveBeenCalledWith('arg1', 'arg2');
    expect(result.current.state).toEqual({ value: 42 });
    expect(result.current.loading).toBe(false);
    expect(result.current.error).toBeNull();
  });

  it('sets loading during exec', async () => {
    let resolve!: (v: unknown) => void;
    const apiFn = vi.fn().mockReturnValue(
      new Promise((r) => {
        resolve = r;
      }),
    );
    const { result } = renderHook(() => useGameApi(apiFn), { wrapper: createWrapper() });

    act(() => {
      result.current.exec();
    });

    await waitFor(() => expect(result.current.loading).toBe(true));

    await act(async () => {
      resolve({ done: true });
    });

    expect(result.current.loading).toBe(false);
  });

  it('sets error on failure', async () => {
    const apiFn = vi.fn().mockRejectedValue(new Error('network'));
    const { result } = renderHook(() => useGameApi(apiFn), { wrapper: createWrapper() });

    await act(async () => {
      await result.current.exec();
    });

    expect(result.current.error).toBe(NETWORK_ERROR_MESSAGE());
    expect(result.current.state).toBeNull();
    expect(result.current.loading).toBe(false);
  });

  it('clears error on successful retry', async () => {
    const apiFn = vi.fn().mockRejectedValueOnce(new Error('fail')).mockResolvedValueOnce({ ok: true });
    const { result } = renderHook(() => useGameApi(apiFn), { wrapper: createWrapper() });

    await act(async () => {
      await result.current.exec();
    });
    expect(result.current.error).not.toBeNull();

    await act(async () => {
      await result.current.exec();
    });
    expect(result.current.error).toBeNull();
    expect(result.current.state).toEqual({ ok: true });
  });

  it('calls onSuccess with response', async () => {
    const apiFn = vi.fn().mockResolvedValue({ data: 'hello' });
    const onSuccess = vi.fn();
    const { result } = renderHook(() => useGameApi(apiFn, { onSuccess }), { wrapper: createWrapper() });

    await act(async () => {
      await result.current.exec();
    });

    expect(onSuccess).toHaveBeenCalledWith({ data: 'hello' });
  });

  it('works without onSuccess option', async () => {
    const apiFn = vi.fn().mockResolvedValue({ v: 1 });
    const { result } = renderHook(() => useGameApi(apiFn), { wrapper: createWrapper() });

    await act(async () => {
      await result.current.exec();
    });

    expect(result.current.state).toEqual({ v: 1 });
  });

  it('does not call onSuccess on error', async () => {
    const apiFn = vi.fn().mockRejectedValue(new Error('err'));
    const onSuccess = vi.fn();
    const { result } = renderHook(() => useGameApi(apiFn, { onSuccess }), { wrapper: createWrapper() });

    await act(async () => {
      await result.current.exec();
    });

    expect(onSuccess).not.toHaveBeenCalled();
  });

  it('allows direct setState usage', async () => {
    const apiFn = vi.fn().mockResolvedValue({ init: true });
    const { result } = renderHook(() => useGameApi(apiFn), { wrapper: createWrapper() });

    await act(async () => {
      await result.current.exec();
    });
    expect(result.current.state).toEqual({ init: true });

    act(() => {
      result.current.setState({ manual: true });
    });
    expect(result.current.state).toEqual({ manual: true });
  });

  it('passes all arguments through to the api function', async () => {
    const apiFn = vi.fn().mockResolvedValue(null);
    const { result } = renderHook(() => useGameApi(apiFn), { wrapper: createWrapper() });

    await act(async () => {
      await result.current.exec('cmd', [1, 2], 99);
    });

    expect(apiFn).toHaveBeenCalledWith('cmd', [1, 2], 99);
  });

  it('preserves state on error', async () => {
    const apiFn = vi.fn().mockResolvedValueOnce({ initial: true }).mockRejectedValueOnce(new Error('fail'));
    const { result } = renderHook(() => useGameApi(apiFn), { wrapper: createWrapper() });

    await act(async () => {
      await result.current.exec();
    });
    expect(result.current.state).toEqual({ initial: true });

    await act(async () => {
      await result.current.exec();
    });
    expect(result.current.state).toEqual({ initial: true });
    expect(result.current.error).not.toBeNull();
  });

  it('uses latest apiFn via ref', async () => {
    const apiFn1 = vi.fn().mockResolvedValue({ v: 1 });
    const apiFn2 = vi.fn().mockResolvedValue({ v: 2 });
    const wrapper = createWrapper();
    const { result, rerender } = renderHook(({ fn }) => useGameApi(fn), {
      initialProps: { fn: apiFn1 },
      wrapper,
    });

    await act(async () => {
      await result.current.exec();
    });
    expect(result.current.state).toEqual({ v: 1 });

    rerender({ fn: apiFn2 });

    await act(async () => {
      await result.current.exec();
    });
    expect(apiFn2).toHaveBeenCalled();
    expect(result.current.state).toEqual({ v: 2 });
  });

  it('keeps loading true until async onSuccess resolves', async () => {
    let resolveCallback!: () => void;
    const apiFn = vi.fn().mockResolvedValue({ v: 1 });
    const onSuccess = vi.fn().mockReturnValue(
      new Promise<void>((r) => {
        resolveCallback = r;
      }),
    );
    const { result } = renderHook(() => useGameApi(apiFn, { onSuccess }), { wrapper: createWrapper() });

    act(() => {
      result.current.exec();
    });

    await waitFor(() => expect(onSuccess).toHaveBeenCalled());
    expect(result.current.loading).toBe(true);

    await act(async () => {
      resolveCallback();
    });

    expect(result.current.loading).toBe(false);
  });

  it('retries the last command on retry()', async () => {
    const apiFn = vi.fn().mockRejectedValueOnce(new Error('fail')).mockResolvedValueOnce({ ok: true });
    const { result } = renderHook(() => useGameApi(apiFn), { wrapper: createWrapper() });

    await act(async () => {
      await result.current.exec('cmd1', 'arg2');
    });
    expect(result.current.error).not.toBeNull();

    await act(async () => {
      await result.current.retry();
    });
    expect(apiFn).toHaveBeenLastCalledWith('cmd1', 'arg2');
    expect(result.current.error).toBeNull();
    expect(result.current.state).toEqual({ ok: true });
  });

  it('retry is a no-op when no previous call', async () => {
    const apiFn = vi.fn();
    const { result } = renderHook(() => useGameApi(apiFn), { wrapper: createWrapper() });

    await act(async () => {
      await result.current.retry();
    });
    expect(apiFn).not.toHaveBeenCalled();
  });

  it('uses latest onSuccess via ref', async () => {
    const apiFn = vi.fn().mockResolvedValue({ v: 1 });
    const onSuccess1 = vi.fn();
    const onSuccess2 = vi.fn();
    const wrapper = createWrapper();
    const { result, rerender } = renderHook(({ cb }) => useGameApi(apiFn, { onSuccess: cb }), {
      initialProps: { cb: onSuccess1 },
      wrapper,
    });

    await act(async () => {
      await result.current.exec();
    });
    expect(onSuccess1).toHaveBeenCalled();

    rerender({ cb: onSuccess2 });

    await act(async () => {
      await result.current.exec();
    });
    expect(onSuccess2).toHaveBeenCalled();
  });
});
