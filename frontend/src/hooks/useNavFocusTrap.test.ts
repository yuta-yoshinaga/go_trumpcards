import { renderHook } from '@testing-library/react';
import { useRef } from 'react';
import { afterEach, describe, expect, it } from 'vitest';
import { useNavFocusTrap } from './useNavFocusTrap';

interface SetupOptions {
  isOpen: boolean;
  isActive: boolean;
  /** Pre-set the toggle button as the document's active element before the effect runs. */
  toggleStartsFocused?: boolean;
}

interface SetupResult {
  container: HTMLElement;
  toggle: HTMLButtonElement;
  buttons: HTMLButtonElement[];
  rerender: (props: SetupOptions) => void;
}

/** Sets up a nav container with three buttons + an out-of-nav toggle button. */
function setup(initial: SetupOptions): SetupResult {
  const toggle = document.createElement('button');
  toggle.textContent = 'toggle';
  document.body.appendChild(toggle);

  const container = document.createElement('nav');
  const buttons: HTMLButtonElement[] = [];
  for (const label of ['first', 'second', 'last']) {
    const btn = document.createElement('button');
    btn.textContent = label;
    container.appendChild(btn);
    buttons.push(btn);
  }
  document.body.appendChild(container);

  if (initial.toggleStartsFocused) {
    toggle.focus();
  }

  const { rerender } = renderHook(
    ({ isOpen, isActive }: SetupOptions) => {
      const containerRef = useRef(container);
      const restoreRef = useRef(toggle);
      useNavFocusTrap(containerRef, restoreRef, isOpen, isActive);
    },
    { initialProps: initial },
  );

  return { container, toggle, buttons, rerender };
}

function pressTab(shift = false): void {
  document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Tab', shiftKey: shift, bubbles: true }));
}

describe('useNavFocusTrap', () => {
  afterEach(() => {
    while (document.body.firstChild) {
      document.body.removeChild(document.body.firstChild);
    }
  });

  it('moves focus to the first focusable element when isOpen flips true', () => {
    const { buttons, rerender } = setup({ isOpen: false, isActive: true, toggleStartsFocused: true });
    rerender({ isOpen: true, isActive: true });
    expect(document.activeElement).toBe(buttons[0]);
  });

  it('restores focus to the toggle button when isOpen flips false', () => {
    const { toggle, rerender } = setup({ isOpen: true, isActive: true });
    // Sanity: open transition focused the first button (not the toggle).
    expect(document.activeElement).not.toBe(toggle);
    rerender({ isOpen: false, isActive: true });
    expect(document.activeElement).toBe(toggle);
  });

  it('does not install the Tab handler when isActive is false (tablet+)', () => {
    const { buttons } = setup({ isOpen: true, isActive: false });
    // Open-transition focus still moves to the first button…
    expect(document.activeElement).toBe(buttons[0]);
    // …but Tab from the last button does NOT wrap (no trap installed).
    buttons[buttons.length - 1].focus();
    pressTab();
    // No preventDefault means the browser would advance focus naturally;
    // since happy-dom does not implement Tab navigation, the active element
    // simply stays where it was. The key assertion is that focus did NOT
    // jump back to the first button (which would prove a trap is installed).
    expect(document.activeElement).toBe(buttons[buttons.length - 1]);
  });

  it('wraps Tab from the last focusable element back to the first', () => {
    const { buttons } = setup({ isOpen: true, isActive: true });
    buttons[buttons.length - 1].focus();
    pressTab();
    expect(document.activeElement).toBe(buttons[0]);
  });

  it('wraps Shift+Tab from the first focusable element back to the last', () => {
    const { buttons } = setup({ isOpen: true, isActive: true });
    buttons[0].focus();
    pressTab(true);
    expect(document.activeElement).toBe(buttons[buttons.length - 1]);
  });

  it('wraps Tab forward when active element has escaped the container', () => {
    // Reproduces the "focus escaped the nav" path from useNavFocusTrap.ts:56
    // (`!container.contains(active)`). The toggle button lives outside the
    // container, so focusing it before the Tab keydown forces that branch.
    const { buttons, toggle } = setup({ isOpen: true, isActive: true });
    toggle.focus();
    pressTab();
    expect(document.activeElement).toBe(buttons[0]);
  });

  it('wraps Shift+Tab to the last focusable when active element has escaped', () => {
    const { buttons, toggle } = setup({ isOpen: true, isActive: true });
    toggle.focus();
    pressTab(true);
    expect(document.activeElement).toBe(buttons[buttons.length - 1]);
  });

  it('ignores non-Tab keydown events', () => {
    const { buttons } = setup({ isOpen: true, isActive: true });
    buttons[1].focus();
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter', bubbles: true }));
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'a', bubbles: true }));
    // No trap re-targeting should happen for non-Tab keys.
    expect(document.activeElement).toBe(buttons[1]);
  });

  it('does not move focus on an isActive flip while staying open', () => {
    // Re-running the effect because of an isActive flip (viewport resize)
    // must not steal focus from the user mid-interaction.
    const { buttons, rerender } = setup({ isOpen: true, isActive: true });
    buttons[1].focus();
    rerender({ isOpen: true, isActive: false });
    expect(document.activeElement).toBe(buttons[1]);
  });
});
