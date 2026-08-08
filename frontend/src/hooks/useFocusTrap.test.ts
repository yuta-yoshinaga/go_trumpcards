import { act, renderHook } from '@testing-library/react';
import type { RefObject } from 'react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { useFocusTrap } from './useFocusTrap';

/** Builds a container with `n` buttons, mounts it, and returns a ref to it. */
function mountContainer(n: number, tabIndex?: number): { ref: RefObject<HTMLDivElement>; cleanup: () => void } {
  const container = document.createElement('div');
  if (tabIndex !== undefined) container.tabIndex = tabIndex;
  for (let i = 0; i < n; i++) {
    const b = document.createElement('button');
    b.textContent = `b${i}`;
    container.appendChild(b);
  }
  document.body.appendChild(container);
  return { ref: { current: container }, cleanup: () => container.remove() };
}

describe('useFocusTrap', () => {
  afterEach(() => {
    document.body.innerHTML = '';
  });

  it('focuses the first focusable child on open', () => {
    const { ref, cleanup } = mountContainer(2);
    renderHook(() => useFocusTrap(ref, true, () => {}));
    expect(document.activeElement).toBe(ref.current.querySelector('button'));
    cleanup();
  });

  it('falls back to the container when it has no focusable children', () => {
    const { ref, cleanup } = mountContainer(0, -1);
    renderHook(() => useFocusTrap(ref, true, () => {}));
    expect(document.activeElement).toBe(ref.current);
    cleanup();
  });

  it('calls onClose on Escape', () => {
    const { ref, cleanup } = mountContainer(1);
    const onClose = vi.fn();
    renderHook(() => useFocusTrap(ref, true, onClose));
    act(() => document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' })));
    expect(onClose).toHaveBeenCalledTimes(1);
    cleanup();
  });

  it('wraps Tab from last to first and Shift+Tab from first to last', () => {
    const { ref, cleanup } = mountContainer(3);
    renderHook(() => useFocusTrap(ref, true, () => {}));
    const buttons = ref.current.querySelectorAll('button');
    const last = buttons[2] as HTMLElement;
    const first = buttons[0] as HTMLElement;

    last.focus();
    act(() => document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Tab' })));
    expect(document.activeElement).toBe(first);

    first.focus();
    act(() => document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Tab', shiftKey: true })));
    expect(document.activeElement).toBe(last);
    cleanup();
  });

  it('pulls focus back inside when it has escaped the container', () => {
    const outside = document.createElement('button');
    document.body.appendChild(outside);
    const { ref, cleanup } = mountContainer(2);
    renderHook(() => useFocusTrap(ref, true, () => {}));

    // Focus escaped to an element outside the trap (e.g. via a mouse click).
    outside.focus();
    act(() => document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Tab' })));
    expect(document.activeElement).toBe(ref.current.querySelector('button'));
    outside.remove();
    cleanup();
  });

  it('restores focus to the previously-focused element on close', () => {
    const trigger = document.createElement('button');
    document.body.appendChild(trigger);
    trigger.focus();
    const { ref, cleanup } = mountContainer(1);
    const { rerender } = renderHook(({ open }) => useFocusTrap(ref, open, () => {}), {
      initialProps: { open: true },
    });
    expect(document.activeElement).not.toBe(trigger);
    rerender({ open: false });
    expect(document.activeElement).toBe(trigger);
    trigger.remove();
    cleanup();
  });

  it('does not trap while closed', () => {
    const { ref, cleanup } = mountContainer(1);
    const onClose = vi.fn();
    renderHook(() => useFocusTrap(ref, false, onClose));
    act(() => document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' })));
    expect(onClose).not.toHaveBeenCalled();
    cleanup();
  });
  // Non-modal panels (a landmark `role="region"`, say) still want the open /
  // Escape / restore half of this hook, but cycling Tab inside them is a
  // WCAG 2.1.2 keyboard trap. `trap: false` keeps everything except the
  // cycling. See issues #5182 and #5183.
  describe('trap: false', () => {
    it('does not cycle Tab at either boundary', () => {
      const { ref, cleanup } = mountContainer(2);
      renderHook(() => useFocusTrap(ref, true, () => {}, { trap: false }));
      const [first, last] = Array.from(ref.current.querySelectorAll('button'));

      last.focus();
      const fwd = new KeyboardEvent('keydown', { key: 'Tab', bubbles: true });
      const fwdSpy = vi.spyOn(fwd, 'preventDefault');
      act(() => {
        document.dispatchEvent(fwd);
      });
      expect(fwdSpy).not.toHaveBeenCalled();
      expect(document.activeElement).toBe(last);

      first.focus();
      const back = new KeyboardEvent('keydown', { key: 'Tab', shiftKey: true, bubbles: true });
      const backSpy = vi.spyOn(back, 'preventDefault');
      act(() => {
        document.dispatchEvent(back);
      });
      expect(backSpy).not.toHaveBeenCalled();
      expect(document.activeElement).toBe(first);

      cleanup();
    });

    it('still focuses on open, closes on Escape, and restores on close', () => {
      const outside = document.createElement('button');
      document.body.appendChild(outside);
      outside.focus();

      const { ref, cleanup } = mountContainer(2);
      const onClose = vi.fn();
      const { unmount } = renderHook(() => useFocusTrap(ref, true, onClose, { trap: false }));

      expect(document.activeElement).toBe(ref.current.querySelector('button'));

      act(() => {
        document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }));
      });
      expect(onClose).toHaveBeenCalledTimes(1);

      unmount();
      expect(document.activeElement).toBe(outside);

      outside.remove();
      cleanup();
    });
  });

  // Negative control: the default must still trap, or `trap: false` proves nothing.
  it('traps by default when no options are passed', () => {
    const { ref, cleanup } = mountContainer(2);
    renderHook(() => useFocusTrap(ref, true, () => {}));
    const [first, last] = Array.from(ref.current.querySelectorAll('button'));

    last.focus();
    act(() => {
      document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Tab', bubbles: true }));
    });
    expect(document.activeElement).toBe(first);

    cleanup();
  });
});
