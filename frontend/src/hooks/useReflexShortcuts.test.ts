import { renderHook } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { useReflexShortcuts } from './useReflexShortcuts';

function press(key: string, options: KeyboardEventInit = {}): boolean {
  const event = new KeyboardEvent('keydown', { key, bubbles: true, cancelable: true, ...options });
  return window.dispatchEvent(event);
}

describe('useReflexShortcuts', () => {
  it('triggers onStep on Enter / S / s', () => {
    const onStep = vi.fn();
    const onSlap = vi.fn();
    renderHook(() => useReflexShortcuts({ onStep, onSlap, enabled: true }));

    press('Enter');
    press('s');
    press('S');
    expect(onStep).toHaveBeenCalledTimes(3);
    expect(onSlap).not.toHaveBeenCalled();
  });

  it('triggers onSlap on Space', () => {
    const onStep = vi.fn();
    const onSlap = vi.fn();
    renderHook(() => useReflexShortcuts({ onStep, onSlap, enabled: true }));

    press(' ');
    expect(onSlap).toHaveBeenCalledTimes(1);
    expect(onStep).not.toHaveBeenCalled();
  });

  it('does nothing while disabled', () => {
    const onStep = vi.fn();
    const onSlap = vi.fn();
    renderHook(() => useReflexShortcuts({ onStep, onSlap, enabled: false }));

    press('Enter');
    press(' ');
    expect(onStep).not.toHaveBeenCalled();
    expect(onSlap).not.toHaveBeenCalled();
  });

  it('ignores keystrokes while a modifier is held (browser shortcuts)', () => {
    const onStep = vi.fn();
    const onSlap = vi.fn();
    renderHook(() => useReflexShortcuts({ onStep, onSlap, enabled: true }));

    press('Enter', { ctrlKey: true });
    press(' ', { metaKey: true });
    expect(onStep).not.toHaveBeenCalled();
    expect(onSlap).not.toHaveBeenCalled();
  });

  it('skips when the target is an editable element', () => {
    const onStep = vi.fn();
    const onSlap = vi.fn();
    renderHook(() => useReflexShortcuts({ onStep, onSlap, enabled: true }));

    const input = document.createElement('input');
    document.body.appendChild(input);
    input.focus();
    input.dispatchEvent(new KeyboardEvent('keydown', { key: ' ', bubbles: true }));
    input.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter', bubbles: true }));
    document.body.removeChild(input);

    expect(onStep).not.toHaveBeenCalled();
    expect(onSlap).not.toHaveBeenCalled();
  });

  it('honours the latest handler refs across re-renders', () => {
    const initialStep = vi.fn();
    const initialSlap = vi.fn();
    const { rerender } = renderHook(({ onStep, onSlap }) => useReflexShortcuts({ onStep, onSlap, enabled: true }), {
      initialProps: { onStep: initialStep, onSlap: initialSlap },
    });

    const nextStep = vi.fn();
    const nextSlap = vi.fn();
    rerender({ onStep: nextStep, onSlap: nextSlap });

    press('Enter');
    press(' ');
    expect(initialStep).not.toHaveBeenCalled();
    expect(initialSlap).not.toHaveBeenCalled();
    expect(nextStep).toHaveBeenCalledTimes(1);
    expect(nextSlap).toHaveBeenCalledTimes(1);
  });
});
