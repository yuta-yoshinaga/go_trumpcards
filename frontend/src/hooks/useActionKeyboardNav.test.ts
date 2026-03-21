import { afterEach, describe, expect, it, vi } from 'bun:test';
import { renderHook } from '@testing-library/react';
import { useActionKeyboardNav } from './useActionKeyboardNav';

function fire(key: string, target?: Partial<HTMLElement>) {
  const event = new KeyboardEvent('keydown', { key, bubbles: true });
  if (target) {
    Object.defineProperty(event, 'target', { value: target });
  }
  document.dispatchEvent(event);
}

describe('useActionKeyboardNav', () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('fires action when matching key is pressed', () => {
    const hitAction = vi.fn();
    const standAction = vi.fn();
    renderHook(() =>
      useActionKeyboardNav({
        bindings: [
          { key: 'h', action: hitAction },
          { key: 's', action: standAction },
        ],
        enabled: true,
      }),
    );

    fire('h');
    expect(hitAction).toHaveBeenCalledTimes(1);
    expect(standAction).not.toHaveBeenCalled();

    fire('s');
    expect(standAction).toHaveBeenCalledTimes(1);
  });

  it('ignores non-matching keys', () => {
    const action = vi.fn();
    renderHook(() =>
      useActionKeyboardNav({
        bindings: [{ key: 'h', action }],
        enabled: true,
      }),
    );

    fire('x');
    fire('1');
    fire('Enter');
    expect(action).not.toHaveBeenCalled();
  });

  it('does nothing when globally disabled', () => {
    const action = vi.fn();
    renderHook(() =>
      useActionKeyboardNav({
        bindings: [{ key: 'h', action }],
        enabled: false,
      }),
    );

    fire('h');
    expect(action).not.toHaveBeenCalled();
  });

  it('skips per-binding disabled actions', () => {
    const action = vi.fn();
    renderHook(() =>
      useActionKeyboardNav({
        bindings: [{ key: 'h', action, enabled: false }],
        enabled: true,
      }),
    );

    fire('h');
    expect(action).not.toHaveBeenCalled();
  });

  it('fires when per-binding enabled is true', () => {
    const action = vi.fn();
    renderHook(() =>
      useActionKeyboardNav({
        bindings: [{ key: 'h', action, enabled: true }],
        enabled: true,
      }),
    );

    fire('h');
    expect(action).toHaveBeenCalledTimes(1);
  });

  it('ignores keys when target is an input element', () => {
    const action = vi.fn();
    renderHook(() =>
      useActionKeyboardNav({
        bindings: [{ key: 'h', action }],
        enabled: true,
      }),
    );

    fire('h', { tagName: 'INPUT' });
    expect(action).not.toHaveBeenCalled();
  });

  it('ignores keys when target is a textarea element', () => {
    const action = vi.fn();
    renderHook(() =>
      useActionKeyboardNav({
        bindings: [{ key: 'h', action }],
        enabled: true,
      }),
    );

    fire('h', { tagName: 'TEXTAREA' });
    expect(action).not.toHaveBeenCalled();
  });

  it('ignores keys when target is a select element', () => {
    const action = vi.fn();
    renderHook(() =>
      useActionKeyboardNav({
        bindings: [{ key: 'h', action }],
        enabled: true,
      }),
    );

    fire('h', { tagName: 'SELECT' });
    expect(action).not.toHaveBeenCalled();
  });

  it('cleans up listener on unmount', () => {
    const action = vi.fn();
    const { unmount } = renderHook(() =>
      useActionKeyboardNav({
        bindings: [{ key: 'h', action }],
        enabled: true,
      }),
    );

    unmount();
    fire('h');
    expect(action).not.toHaveBeenCalled();
  });

  it('handles per-binding enabled undefined as enabled', () => {
    const action = vi.fn();
    renderHook(() =>
      useActionKeyboardNav({
        bindings: [{ key: 'h', action, enabled: undefined }],
        enabled: true,
      }),
    );

    fire('h');
    expect(action).toHaveBeenCalledTimes(1);
  });
});
