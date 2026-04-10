import { act, renderHook } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { useCardSwipeSelection } from './useCardSwipeSelection';

/**
 * Tests for the swipe/drag bulk card selection hook. We stub
 * `document.elementFromPoint` so each x-coordinate (10 px wide) resolves to a
 * synthetic card element whose `data-card-index` the hook can read.
 */
describe('useCardSwipeSelection', () => {
  let cardElements: Map<number, HTMLElement>;

  beforeEach(() => {
    cardElements = new Map();
    for (let i = 0; i < 5; i++) {
      const el = document.createElement('button');
      el.setAttribute('data-card-index', String(i));
      document.body.appendChild(el);
      cardElements.set(i, el);
    }
    vi.spyOn(document, 'elementFromPoint').mockImplementation((x: number) => {
      const idx = Math.floor(x / 10);
      return cardElements.get(idx) ?? null;
    });
  });

  afterEach(() => {
    vi.restoreAllMocks();
    cardElements.forEach((el) => {
      el.remove();
    });
    cardElements.clear();
  });

  const firePointer = (type: 'pointermove' | 'pointerup' | 'pointercancel', x: number, y = 0) => {
    const ev = new Event(type, { bubbles: true });
    Object.defineProperty(ev, 'clientX', { value: x });
    Object.defineProperty(ev, 'clientY', { value: y });
    window.dispatchEvent(ev);
  };

  it('does not toggle on a plain tap (no pointer movement)', () => {
    const toggle = vi.fn();
    const { result } = renderHook(() => useCardSwipeSelection({ selected: [], toggle }));
    act(() => {
      result.current.onPointerDown(0);
      firePointer('pointerup', 0);
    });
    expect(toggle).not.toHaveBeenCalled();
  });

  it('bulk-selects cards traversed when starting from an unselected card', () => {
    const toggle = vi.fn();
    const { result } = renderHook(() => useCardSwipeSelection({ selected: [], toggle }));
    act(() => {
      result.current.onPointerDown(0);
      firePointer('pointermove', 15);
      firePointer('pointermove', 25);
      firePointer('pointerup', 25);
    });
    expect(toggle).toHaveBeenCalledWith(0);
    expect(toggle).toHaveBeenCalledWith(1);
    expect(toggle).toHaveBeenCalledWith(2);
    expect(toggle).toHaveBeenCalledTimes(3);
  });

  it('bulk-deselects cards traversed when starting from a selected card', () => {
    const toggle = vi.fn();
    const { result } = renderHook(() => useCardSwipeSelection({ selected: [0, 1, 2], toggle }));
    act(() => {
      result.current.onPointerDown(0);
      firePointer('pointermove', 15);
      firePointer('pointermove', 25);
      firePointer('pointerup', 25);
    });
    expect(toggle).toHaveBeenCalledWith(0);
    expect(toggle).toHaveBeenCalledWith(1);
    expect(toggle).toHaveBeenCalledWith(2);
    expect(toggle).toHaveBeenCalledTimes(3);
  });

  it('leaves cards alone that already match the swipe mode', () => {
    const toggle = vi.fn();
    // Start on card 0 (unselected) → mode = select. Card 1 is already selected
    // so it should be skipped. Card 2 is unselected so it gets toggled.
    const { result } = renderHook(() => useCardSwipeSelection({ selected: [1], toggle }));
    act(() => {
      result.current.onPointerDown(0);
      firePointer('pointermove', 15);
      firePointer('pointermove', 25);
      firePointer('pointerup', 25);
    });
    expect(toggle).toHaveBeenCalledWith(0);
    expect(toggle).not.toHaveBeenCalledWith(1);
    expect(toggle).toHaveBeenCalledWith(2);
    expect(toggle).toHaveBeenCalledTimes(2);
  });

  it('toggles each card only once, even if the pointer re-enters it', () => {
    const toggle = vi.fn();
    const { result } = renderHook(() => useCardSwipeSelection({ selected: [], toggle }));
    act(() => {
      result.current.onPointerDown(0);
      firePointer('pointermove', 15);
      firePointer('pointermove', 25);
      firePointer('pointermove', 15); // back to card 1
      firePointer('pointermove', 5); // back to card 0
      firePointer('pointerup', 5);
    });
    expect(toggle).toHaveBeenCalledTimes(3);
  });

  it('suppresses the trailing click after a real swipe', () => {
    const toggle = vi.fn();
    const clickHandler = vi.fn();
    const card1 = cardElements.get(1);
    if (!card1) throw new Error('card1 missing');
    card1.addEventListener('click', clickHandler);
    const { result } = renderHook(() => useCardSwipeSelection({ selected: [], toggle }));
    act(() => {
      result.current.onPointerDown(0);
      firePointer('pointermove', 15);
      firePointer('pointerup', 15);
      card1.dispatchEvent(new MouseEvent('click', { bubbles: true, cancelable: true }));
    });
    expect(clickHandler).not.toHaveBeenCalled();
    card1.removeEventListener('click', clickHandler);
  });

  it('does not suppress clicks after a plain tap', () => {
    const toggle = vi.fn();
    const clickHandler = vi.fn();
    const card0 = cardElements.get(0);
    if (!card0) throw new Error('card0 missing');
    card0.addEventListener('click', clickHandler);
    const { result } = renderHook(() => useCardSwipeSelection({ selected: [], toggle }));
    act(() => {
      result.current.onPointerDown(0);
      firePointer('pointerup', 0);
      card0.dispatchEvent(new MouseEvent('click', { bubbles: true, cancelable: true }));
    });
    expect(clickHandler).toHaveBeenCalledTimes(1);
    card0.removeEventListener('click', clickHandler);
  });

  it('cancels an in-progress swipe when a native drag starts', () => {
    const toggle = vi.fn();
    const { result } = renderHook(() => useCardSwipeSelection({ selected: [], toggle }));
    act(() => {
      result.current.onPointerDown(0);
      window.dispatchEvent(new Event('dragstart'));
      firePointer('pointermove', 15);
      firePointer('pointerup', 15);
    });
    expect(toggle).not.toHaveBeenCalled();
  });

  it('is inert when disabled', () => {
    const toggle = vi.fn();
    const { result } = renderHook(() => useCardSwipeSelection({ selected: [], toggle, enabled: false }));
    act(() => {
      result.current.onPointerDown(0);
      firePointer('pointermove', 15);
      firePointer('pointerup', 15);
    });
    expect(toggle).not.toHaveBeenCalled();
  });
});
