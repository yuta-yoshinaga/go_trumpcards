import { renderHook } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { registerOpenModal } from './keyboardNavUtils';
import { useCardKeyboardNav } from './useCardKeyboardNav';

let unregisterModal: (() => void) | undefined;

function fire(key: string, target?: Partial<HTMLElement>) {
  const event = new KeyboardEvent('keydown', { key, bubbles: true });
  if (target) {
    Object.defineProperty(event, 'target', { value: target });
  }
  document.dispatchEvent(event);
}

describe('useCardKeyboardNav', () => {
  afterEach(() => {
    unregisterModal?.();
    unregisterModal = undefined;
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

  it('moves the selection position within the hand and Enter toggles the focused card', () => {
    const onToggle = vi.fn();
    const onConfirm = vi.fn();
    const onFocusIndexChange = vi.fn();
    renderHook(() =>
      useCardKeyboardNav({
        cardCount: 13,
        onToggle,
        onConfirm,
        onClear: vi.fn(),
        enabled: true,
        arrowSelection: true,
        onFocusIndexChange,
      }),
    );

    fire('ArrowLeft');
    expect(onFocusIndexChange).toHaveBeenLastCalledWith(0);
    for (let i = 0; i < 12; i++) fire('ArrowRight');
    expect(onFocusIndexChange).toHaveBeenLastCalledWith(12);
    fire('ArrowRight');
    expect(onFocusIndexChange).toHaveBeenCalledTimes(14);
    fire('Enter');
    expect(onToggle).toHaveBeenCalledWith(12);
    expect(onConfirm).not.toHaveBeenCalled();
  });

  it('keeps the arrow position after Enter causes a callback change and rerender', () => {
    const onToggleBefore = vi.fn();
    const onToggleAfter = vi.fn();
    const onConfirm = vi.fn();
    const { rerender } = renderHook(
      ({ onToggle }) =>
        useCardKeyboardNav({
          cardCount: 5,
          onToggle,
          onConfirm,
          onClear: vi.fn(),
          enabled: true,
          arrowSelection: true,
        }),
      { initialProps: { onToggle: onToggleBefore } },
    );

    fire('ArrowRight');
    fire('ArrowRight');
    fire('Enter');
    expect(onToggleBefore).toHaveBeenCalledWith(2);

    rerender({ onToggle: onToggleAfter });
    fire('ArrowRight');
    fire('Enter');
    expect(onToggleAfter).toHaveBeenCalledWith(3);
    expect(onConfirm).not.toHaveBeenCalled();
  });

  it('clamps the focused position when the hand gets smaller', () => {
    const onFocusIndexChange = vi.fn();
    const { rerender } = renderHook(
      ({ cardCount }) =>
        useCardKeyboardNav({
          cardCount,
          onToggle: vi.fn(),
          onConfirm: vi.fn(),
          onClear: vi.fn(),
          enabled: true,
          arrowSelection: true,
          onFocusIndexChange,
        }),
      { initialProps: { cardCount: 5 } },
    );

    fire('ArrowRight');
    fire('ArrowRight');
    fire('ArrowRight');
    fire('ArrowRight');
    rerender({ cardCount: 2 });
    fire('ArrowRight');
    expect(onFocusIndexChange).toHaveBeenLastCalledWith(1);
  });

  it('does not create a focused selection when the hand is empty', () => {
    const onFocusIndexChange = vi.fn();
    renderHook(() =>
      useCardKeyboardNav({
        cardCount: 0,
        onToggle: vi.fn(),
        onConfirm: vi.fn(),
        onClear: vi.fn(),
        enabled: true,
        arrowSelection: true,
        onFocusIndexChange,
      }),
    );
    fire('ArrowRight');
    expect(onFocusIndexChange).not.toHaveBeenCalled();
  });

  it('blocks board keys while a modal is open, then resumes them', () => {
    const onToggle = vi.fn();
    const onConfirm = vi.fn();
    const onClear = vi.fn();
    renderHook(() => useCardKeyboardNav({ cardCount: 5, onToggle, onConfirm, onClear, enabled: true }));
    unregisterModal = registerOpenModal();
    fire('1');
    fire('Enter');
    fire('Escape');
    expect(onToggle).not.toHaveBeenCalled();
    expect(onConfirm).not.toHaveBeenCalled();
    expect(onClear).not.toHaveBeenCalled();
    unregisterModal();
    fire('1');
    expect(onToggle).toHaveBeenCalledWith(0);
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

  it('does not direct-play a card rejected by the legality predicate', () => {
    const onDirectPlay = vi.fn();
    renderHook(() =>
      useCardKeyboardNav({
        cardCount: 3,
        onToggle: vi.fn(),
        onConfirm: vi.fn(),
        onClear: vi.fn(),
        enabled: true,
        onDirectPlay,
        canDirectPlay: (index) => index === 1,
      }),
    );

    fire('1');
    fire('2');
    expect(onDirectPlay).toHaveBeenCalledTimes(1);
    expect(onDirectPlay).toHaveBeenCalledWith(1);
  });

  it('direct-plays a card accepted by the legality predicate', () => {
    const onDirectPlay = vi.fn();
    renderHook(() =>
      useCardKeyboardNav({
        cardCount: 3,
        onToggle: vi.fn(),
        onConfirm: vi.fn(),
        onClear: vi.fn(),
        enabled: true,
        onDirectPlay,
        canDirectPlay: () => true,
      }),
    );

    fire('2');
    expect(onDirectPlay).toHaveBeenCalledTimes(1);
    expect(onDirectPlay).toHaveBeenCalledWith(1);
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
