import { renderHook } from '@testing-library/react';
import { useRef } from 'react';
import { afterEach, describe, expect, it } from 'vitest';
import { useDetailsOutsideClick } from './useDetailsOutsideClick';

describe('useDetailsOutsideClick', () => {
  afterEach(() => {
    while (document.body.firstChild) {
      document.body.removeChild(document.body.firstChild);
    }
  });

  function setupContainer(): { container: HTMLElement; details: HTMLDetailsElement } {
    const container = document.createElement('section');
    const details = document.createElement('details');
    details.setAttribute('open', '');
    const inside = document.createElement('p');
    inside.textContent = 'inside';
    details.appendChild(inside);
    container.appendChild(details);
    document.body.appendChild(container);
    return { container, details };
  }

  function dispatchMouseDownOutside(): void {
    const outside = document.createElement('div');
    document.body.appendChild(outside);
    outside.dispatchEvent(new MouseEvent('mousedown', { bubbles: true }));
  }

  it('closes open details on outside mousedown when enabled', () => {
    const { container, details } = setupContainer();
    renderHook(() => {
      const ref = useRef(container);
      useDetailsOutsideClick(ref, true);
    });
    dispatchMouseDownOutside();
    expect(details.hasAttribute('open')).toBe(false);
  });

  it('leaves details open when disabled', () => {
    const { container, details } = setupContainer();
    renderHook(() => {
      const ref = useRef(container);
      useDetailsOutsideClick(ref, false);
    });
    dispatchMouseDownOutside();
    expect(details.hasAttribute('open')).toBe(true);
  });

  it('keeps details open when the click is inside the details', () => {
    const { container, details } = setupContainer();
    renderHook(() => {
      const ref = useRef(container);
      useDetailsOutsideClick(ref, true);
    });
    const inner = details.querySelector('p');
    inner?.dispatchEvent(new MouseEvent('mousedown', { bubbles: true }));
    expect(details.hasAttribute('open')).toBe(true);
  });
});
