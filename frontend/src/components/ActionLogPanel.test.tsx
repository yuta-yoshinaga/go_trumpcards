import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { ActionLogEntry } from '../types/card';
import { ActionLogPanel } from './ActionLogPanel';

describe('ActionLogPanel', () => {
  const writeText = vi.fn().mockResolvedValue(undefined);

  beforeEach(() => {
    Object.defineProperty(navigator, 'clipboard', {
      value: { writeText },
      writable: true,
      configurable: true,
    });
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  const sampleEntries: ActionLogEntry[] = [
    { turnNumber: 1, playerIdx: 0, actionType: 'hit', detail: 'drew a card', cards: [{ design: 'SPADE', value: 1 }] },
    { turnNumber: 2, playerIdx: -1, actionType: 'deal', detail: 'dealt cards' },
    { turnNumber: 3, playerIdx: 1, actionType: 'stand', detail: 'stood', cards: [] },
  ];

  it('renders entries with correct formatting', () => {
    render(<ActionLogPanel entries={sampleEntries} onClose={vi.fn()} />);

    // Turn 1: Player 0 with cards
    expect(screen.getByText(/T1 \[Player 0\] hit: drew a card \[SPADE 1\]/)).toBeInTheDocument();
    // Turn 2: System (playerIdx < 0), no cards
    expect(screen.getByText(/T2 \[SYSTEM\] deal: dealt cards/)).toBeInTheDocument();
    // Turn 3: Player 1, empty cards array (no card bracket appended)
    expect(screen.getByText(/T3 \[Player 1\] stand: stood/)).toBeInTheDocument();
  });

  it('shows empty message when entries is empty', () => {
    render(<ActionLogPanel entries={[]} onClose={vi.fn()} />);
    expect(screen.getByText('棋譜はありません。')).toBeInTheDocument();
  });

  it('copy button calls navigator.clipboard.writeText', async () => {
    render(<ActionLogPanel entries={sampleEntries} onClose={vi.fn()} />);
    const copyButton = screen.getByRole('button', { name: 'コピー' });
    fireEvent.click(copyButton);

    await waitFor(() => {
      expect(writeText).toHaveBeenCalledTimes(1);
    });
    expect(writeText).toHaveBeenCalledWith(expect.stringContaining('T1 [Player 0] hit: drew a card'));
  });

  it('shows "コピーしました" after copy', async () => {
    render(<ActionLogPanel entries={sampleEntries} onClose={vi.fn()} />);
    const copyButton = screen.getByRole('button', { name: 'コピー' });
    fireEvent.click(copyButton);

    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'コピーしました' })).toBeInTheDocument();
    });
  });

  it('download button creates and clicks a download link', () => {
    const createObjectURL = vi.fn().mockReturnValue('blob:http://localhost/fake');
    const revokeObjectURL = vi.fn();
    vi.stubGlobal('URL', { createObjectURL, revokeObjectURL });

    const clickSpy = vi.fn();
    const originalCreateElement = document.createElement.bind(document);
    const createElementSpy = vi
      .spyOn(document, 'createElement')
      .mockImplementation((tag: string, options?: ElementCreationOptions) => {
        if (tag === 'a') {
          const el = originalCreateElement(tag, options) as HTMLAnchorElement;
          el.click = clickSpy;
          return el;
        }
        return originalCreateElement(tag, options);
      });

    render(<ActionLogPanel entries={sampleEntries} onClose={vi.fn()} />);
    const downloadButton = screen.getByRole('button', { name: 'ダウンロード' });
    fireEvent.click(downloadButton);

    expect(createObjectURL).toHaveBeenCalledTimes(1);
    expect(clickSpy).toHaveBeenCalledTimes(1);
    expect(revokeObjectURL).toHaveBeenCalledWith('blob:http://localhost/fake');

    createElementSpy.mockRestore();
    vi.unstubAllGlobals();
  });

  it('close button calls onClose', () => {
    const onClose = vi.fn();
    render(<ActionLogPanel entries={sampleEntries} onClose={onClose} />);
    const closeButton = screen.getByRole('button', { name: '閉じる' });
    fireEvent.click(closeButton);
    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it('has role="region" with aria-labelledby pointing to the title', () => {
    render(<ActionLogPanel entries={sampleEntries} onClose={vi.fn()} />);
    const region = screen.getByRole('region', { name: '棋譜' });
    expect(region).toBeInTheDocument();
    const labelledById = region.getAttribute('aria-labelledby');
    expect(labelledById).toBeTruthy();
    const title = document.getElementById(labelledById as string);
    expect(title?.textContent).toBe('棋譜');
  });

  it('announces copy confirmation via aria-live region', async () => {
    render(<ActionLogPanel entries={sampleEntries} onClose={vi.fn()} />);
    const liveRegion = screen.getByTestId('copy-announcer');
    expect(liveRegion).toHaveAttribute('aria-live', 'polite');
    expect(liveRegion).toHaveAttribute('aria-atomic', 'true');
    expect(liveRegion.textContent).toBe('');

    const copyButton = screen.getByRole('button', { name: 'コピー' });
    fireEvent.click(copyButton);

    await waitFor(() => {
      expect(liveRegion.textContent).toBe('コピーしました');
    });
  });

  it('focuses the first focusable element on mount', () => {
    render(<ActionLogPanel entries={sampleEntries} onClose={vi.fn()} />);
    const copyButton = screen.getByRole('button', { name: 'コピー' });
    expect(document.activeElement).toBe(copyButton);
  });

  it('does not focus when no focusable elements exist', () => {
    const spy = vi
      .spyOn(HTMLElement.prototype, 'querySelectorAll')
      .mockReturnValueOnce([] as unknown as NodeListOf<HTMLElement>);
    render(<ActionLogPanel entries={[]} onClose={vi.fn()} />);
    expect(document.activeElement).toBe(document.body);
    spy.mockRestore();
  });

  // The panel is a landmark `role="region"`, not a dialog: it has no
  // `aria-modal`, the game behind it stays live, and it is meant to be read
  // alongside the board. Cycling Tab inside it left keyboard users with no way
  // out except finding the close button by sight — WCAG 2.1.2 (Level A).
  // See issue #5183.
  it('does not trap focus: Tab from the last element leaves the panel', () => {
    render(<ActionLogPanel entries={sampleEntries} onClose={vi.fn()} />);
    const closeButton = screen.getByRole('button', { name: '閉じる' });
    const copyButton = screen.getByRole('button', { name: 'コピー' });

    closeButton.focus();
    // Fire on the focused element, not on `document`: a real Tab keydown
    // originates at the focused button and bubbles up through the panel, which
    // is what reaches a panel-level listener. Dispatching on `document`
    // instead never reaches one, so the assertion would hold even with the
    // trap still in place.
    fireEvent.keyDown(closeButton, { key: 'Tab', shiftKey: false });
    // jsdom does not move focus for Tab; what matters is that nothing wrapped
    // it back to the top of the panel.
    expect(document.activeElement).not.toBe(copyButton);
    expect(document.activeElement).toBe(closeButton);
  });

  it('does not trap focus: Shift+Tab from the first element leaves the panel', () => {
    render(<ActionLogPanel entries={sampleEntries} onClose={vi.fn()} />);
    const copyButton = screen.getByRole('button', { name: 'コピー' });
    const closeButton = screen.getByRole('button', { name: '閉じる' });

    copyButton.focus();
    fireEvent.keyDown(copyButton, { key: 'Tab', shiftKey: true });
    expect(document.activeElement).not.toBe(closeButton);
    expect(document.activeElement).toBe(copyButton);
  });

  it('never calls preventDefault on Tab', () => {
    render(<ActionLogPanel entries={sampleEntries} onClose={vi.fn()} />);
    const closeButton = screen.getByRole('button', { name: '閉じる' });
    const copyButton = screen.getByRole('button', { name: 'コピー' });

    // Both boundaries: the last element with Tab and the first with Shift+Tab
    // are exactly the two positions the old trap intercepted.
    for (const [el, shiftKey] of [
      [closeButton, false],
      [copyButton, true],
    ] as const) {
      el.focus();
      const event = new KeyboardEvent('keydown', { key: 'Tab', shiftKey, bubbles: true });
      const preventDefaultSpy = vi.spyOn(event, 'preventDefault');
      el.dispatchEvent(event);
      expect(preventDefaultSpy).not.toHaveBeenCalled();
    }
  });

  it('closes on Escape', () => {
    const onClose = vi.fn();
    render(<ActionLogPanel entries={sampleEntries} onClose={onClose} />);

    fireEvent.keyDown(document, { key: 'Escape' });
    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it('restores focus to the trigger when unmounted without the close button', () => {
    const triggerButton = document.createElement('button');
    document.body.appendChild(triggerButton);
    triggerButton.focus();

    const { unmount } = render(<ActionLogPanel entries={sampleEntries} onClose={vi.fn()} />);
    expect(document.activeElement).not.toBe(triggerButton);

    // The page unmounts the panel by clearing its state, which can happen
    // without the close button ever being clicked (game reset, route change).
    unmount();
    expect(document.activeElement).toBe(triggerButton);

    document.body.removeChild(triggerButton);
  });

  it('restores focus to trigger element on close', () => {
    const triggerButton = document.createElement('button');
    triggerButton.textContent = 'Trigger';
    document.body.appendChild(triggerButton);
    triggerButton.focus();

    const onClose = vi.fn();
    const { unmount } = render(<ActionLogPanel entries={sampleEntries} onClose={onClose} />);
    // Panel has stolen focus to its first focusable element
    expect(document.activeElement).not.toBe(triggerButton);

    const closeButton = screen.getByRole('button', { name: '閉じる' });
    fireEvent.click(closeButton);
    expect(onClose).toHaveBeenCalledTimes(1);
    // The page closes the panel by clearing the state onClose feeds, so the
    // real close path is "onClose then unmount"; restore happens on unmount.
    unmount();
    expect(document.activeElement).toBe(triggerButton);

    document.body.removeChild(triggerButton);
  });

  it('does not error when no element was focused before open', () => {
    // document.activeElement is document.body by default
    const onClose = vi.fn();
    render(<ActionLogPanel entries={sampleEntries} onClose={onClose} />);

    const closeButton = screen.getByRole('button', { name: '閉じる' });
    fireEvent.click(closeButton);
    expect(onClose).toHaveBeenCalledTimes(1);
    // Should not throw; body.focus() is a no-op so this is safe
  });

  it('ignores non-Tab keydown events', () => {
    render(<ActionLogPanel entries={sampleEntries} onClose={vi.fn()} />);
    const region = screen.getByRole('region', { name: '棋譜' });
    const copyButton = screen.getByRole('button', { name: 'コピー' });

    copyButton.focus();
    const event = new KeyboardEvent('keydown', { key: 'Escape', bubbles: true });
    const preventDefaultSpy = vi.spyOn(event, 'preventDefault');
    region.dispatchEvent(event);
    expect(preventDefaultSpy).not.toHaveBeenCalled();
  });
});
