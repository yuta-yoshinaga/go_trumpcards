import { renderHook } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { useCardKeyboardNav } from './useCardKeyboardNav';

function fire(key: string, target?: Partial<HTMLElement>) {
  const event = new KeyboardEvent('keydown', { key, bubbles: true });
  if (target) {
    Object.defineProperty(event, 'target', { value: target });
  }
  document.dispatchEvent(event);
}

describe('useCardKeyboardNav', () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('toggles card on number keys 1-9 and 0', () => {
    const onToggle = vi.fn();
    renderHook(() =>
      useCardKeyboardNav({
        cardCount: 10,
        onToggle,
        onConfirm: vi.fn(),
        onClear: vi.fn(),
        enabled: true,
      }),
    );

    fire('1');
    expect(onToggle).toHaveBeenCalledWith(0);

    fire('5');
    expect(onToggle).toHaveBeenCalledWith(4);

    fire('9');
    expect(onToggle).toHaveBeenCalledWith(8);

    fire('0');
    expect(onToggle).toHaveBeenCalledWith(9);
  });

  it('calls onConfirm on Enter', () => {
    const onConfirm = vi.fn();
    renderHook(() =>
      useCardKeyboardNav({
        cardCount: 5,
        onToggle: vi.fn(),
        onConfirm,
        onClear: vi.fn(),
        enabled: true,
      }),
    );

    fire('Enter');
    expect(onConfirm).toHaveBeenCalledTimes(1);
  });

  it('calls onClear on Escape', () => {
    const onClear = vi.fn();
    renderHook(() =>
      useCardKeyboardNav({
        cardCount: 5,
        onToggle: vi.fn(),
        onConfirm: vi.fn(),
        onClear,
        enabled: true,
      }),
    );

    fire('Escape');
    expect(onClear).toHaveBeenCalledTimes(1);
  });

  it('does nothing when enabled is false', () => {
    const onToggle = vi.fn();
    const onConfirm = vi.fn();
    const onClear = vi.fn();
    renderHook(() =>
      useCardKeyboardNav({
        cardCount: 5,
        onToggle,
        onConfirm,
        onClear,
        enabled: false,
      }),
    );

    fire('1');
    fire('Enter');
    fire('Escape');
    expect(onToggle).not.toHaveBeenCalled();
    expect(onConfirm).not.toHaveBeenCalled();
    expect(onClear).not.toHaveBeenCalled();
  });

  it('ignores keys when target is an input element', () => {
    const onToggle = vi.fn();
    renderHook(() =>
      useCardKeyboardNav({
        cardCount: 5,
        onToggle,
        onConfirm: vi.fn(),
        onClear: vi.fn(),
        enabled: true,
      }),
    );

    fire('1', { tagName: 'INPUT' });
    expect(onToggle).not.toHaveBeenCalled();
  });

  it('ignores keys when target is a textarea element', () => {
    const onToggle = vi.fn();
    renderHook(() =>
      useCardKeyboardNav({
        cardCount: 5,
        onToggle,
        onConfirm: vi.fn(),
        onClear: vi.fn(),
        enabled: true,
      }),
    );

    fire('1', { tagName: 'TEXTAREA' });
    expect(onToggle).not.toHaveBeenCalled();
  });

  it('ignores keys when target is a select element', () => {
    const onToggle = vi.fn();
    renderHook(() =>
      useCardKeyboardNav({
        cardCount: 5,
        onToggle,
        onConfirm: vi.fn(),
        onClear: vi.fn(),
        enabled: true,
      }),
    );

    fire('1', { tagName: 'SELECT' });
    expect(onToggle).not.toHaveBeenCalled();
  });

  it('ignores number keys beyond cardCount', () => {
    const onToggle = vi.fn();
    renderHook(() =>
      useCardKeyboardNav({
        cardCount: 3,
        onToggle,
        onConfirm: vi.fn(),
        onClear: vi.fn(),
        enabled: true,
      }),
    );

    fire('4'); // index 3, but only 3 cards (0-2)
    expect(onToggle).not.toHaveBeenCalled();

    fire('3'); // index 2, valid
    expect(onToggle).toHaveBeenCalledWith(2);
  });

  it('ignores 0 key when cardCount < 10', () => {
    const onToggle = vi.fn();
    renderHook(() =>
      useCardKeyboardNav({
        cardCount: 5,
        onToggle,
        onConfirm: vi.fn(),
        onClear: vi.fn(),
        enabled: true,
      }),
    );

    fire('0'); // index 9, but only 5 cards
    expect(onToggle).not.toHaveBeenCalled();
  });

  it('uses onDirectPlay instead of onToggle when provided', () => {
    const onToggle = vi.fn();
    const onDirectPlay = vi.fn();
    renderHook(() =>
      useCardKeyboardNav({
        cardCount: 5,
        onToggle,
        onConfirm: vi.fn(),
        onClear: vi.fn(),
        enabled: true,
        onDirectPlay,
      }),
    );

    fire('1');
    expect(onDirectPlay).toHaveBeenCalledWith(0);
    expect(onToggle).not.toHaveBeenCalled();
  });

  it('cleans up listener on unmount', () => {
    const onToggle = vi.fn();
    const { unmount } = renderHook(() =>
      useCardKeyboardNav({
        cardCount: 5,
        onToggle,
        onConfirm: vi.fn(),
        onClear: vi.fn(),
        enabled: true,
      }),
    );

    unmount();
    fire('1');
    expect(onToggle).not.toHaveBeenCalled();
  });

  it('ignores unrecognized keys', () => {
    const onToggle = vi.fn();
    const onConfirm = vi.fn();
    const onClear = vi.fn();
    renderHook(() =>
      useCardKeyboardNav({
        cardCount: 5,
        onToggle,
        onConfirm,
        onClear,
        enabled: true,
      }),
    );

    fire('a');
    fire('z');
    fire('Tab');
    expect(onToggle).not.toHaveBeenCalled();
    expect(onConfirm).not.toHaveBeenCalled();
    expect(onClear).not.toHaveBeenCalled();
  });
});
