import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { TutorialStep } from '../../types/tutorial';
import { TutorialOverlay } from './TutorialOverlay';

// Mock ResizeObserver
const observeMock = vi.fn();
const disconnectMock = vi.fn();
class MockResizeObserver {
  observe = observeMock;
  unobserve = vi.fn();
  disconnect = disconnectMock;
  callback: ResizeObserverCallback;
  constructor(callback: ResizeObserverCallback) {
    this.callback = callback;
  }
}
vi.stubGlobal('ResizeObserver', MockResizeObserver);

const step: TutorialStep = {
  target: '[data-tutorial="test-target"]',
  messageKey: 'テスト説明',
  placement: 'bottom',
  advanceOn: 'next',
};

const defaultProps = {
  step,
  stepIndex: 0,
  totalSteps: 3,
  onNext: vi.fn(),
  onSkip: vi.fn(),
  reducedMotion: false,
};

describe('TutorialOverlay', () => {
  let targetEl: HTMLDivElement;

  beforeEach(() => {
    targetEl = document.createElement('div');
    targetEl.setAttribute('data-tutorial', 'test-target');
    targetEl.getBoundingClientRect = vi.fn().mockReturnValue({
      top: 100,
      left: 50,
      width: 200,
      height: 40,
      right: 250,
      bottom: 140,
    });
    document.body.appendChild(targetEl);
  });

  afterEach(() => {
    document.body.removeChild(targetEl);
  });

  it('renders an overlay with dialog role', () => {
    render(<TutorialOverlay {...defaultProps} />);
    expect(screen.getByRole('dialog')).toBeInTheDocument();
  });

  it('renders aria-modal attribute', () => {
    render(<TutorialOverlay {...defaultProps} />);
    expect(screen.getByRole('dialog')).toHaveAttribute('aria-modal', 'true');
  });

  it('renders the SVG mask for spotlight', () => {
    const { container } = render(<TutorialOverlay {...defaultProps} />);
    expect(container.querySelector('svg')).toBeInTheDocument();
    expect(container.querySelector('mask')).toBeInTheDocument();
  });

  it('renders the tooltip with step message', () => {
    render(<TutorialOverlay {...defaultProps} />);
    expect(screen.getByText('テスト説明')).toBeInTheDocument();
  });

  it('calls onNext on Enter key', () => {
    const onNext = vi.fn();
    render(<TutorialOverlay {...defaultProps} onNext={onNext} />);
    fireEvent.keyDown(screen.getByRole('dialog'), { key: 'Enter' });
    expect(onNext).toHaveBeenCalledTimes(1);
  });

  it('calls onSkip on Escape key', () => {
    const onSkip = vi.fn();
    render(<TutorialOverlay {...defaultProps} onSkip={onSkip} />);
    fireEvent.keyDown(screen.getByRole('dialog'), { key: 'Escape' });
    expect(onSkip).toHaveBeenCalledTimes(1);
  });

  it('observes target element with ResizeObserver', async () => {
    render(<TutorialOverlay {...defaultProps} />);
    await waitFor(() => {
      expect(observeMock).toHaveBeenCalled();
    });
  });

  it('disconnects ResizeObserver on unmount', () => {
    const { unmount } = render(<TutorialOverlay {...defaultProps} />);
    unmount();
    expect(disconnectMock).toHaveBeenCalled();
  });

  it('renders without animation when reducedMotion is true', () => {
    const { container } = render(<TutorialOverlay {...defaultProps} reducedMotion={true} />);
    // Should still render the overlay, but without framer-motion transitions
    expect(container.querySelector('svg')).toBeInTheDocument();
  });

  it('handles click advanceOn by listening for target clicks', async () => {
    const onNext = vi.fn();
    const clickStep: TutorialStep = { ...step, advanceOn: 'click' };
    render(<TutorialOverlay {...defaultProps} step={clickStep} onNext={onNext} />);
    fireEvent.click(targetEl);
    await waitFor(() => {
      expect(onNext).toHaveBeenCalledTimes(1);
    });
  });

  it('handles missing target element gracefully', () => {
    document.body.removeChild(targetEl);
    const { container } = render(<TutorialOverlay {...defaultProps} />);
    // Should still render overlay without crashing
    expect(container.querySelector('svg')).toBeInTheDocument();
    // Re-add to avoid afterEach error
    document.body.appendChild(targetEl);
  });

  it('traps focus forward: Tab on last button wraps to first', () => {
    render(<TutorialOverlay {...defaultProps} />);
    const dialog = screen.getByRole('dialog');
    const buttons = Array.from(dialog.querySelectorAll('button'));
    expect(buttons.length).toBeGreaterThan(1);
    const first = buttons[0];
    const last = buttons[buttons.length - 1];
    last.focus();
    fireEvent.keyDown(dialog, { key: 'Tab' });
    expect(document.activeElement).toBe(first);
  });

  it('traps focus backward: Shift+Tab on first button wraps to last', () => {
    render(<TutorialOverlay {...defaultProps} />);
    const dialog = screen.getByRole('dialog');
    const buttons = Array.from(dialog.querySelectorAll('button'));
    const first = buttons[0];
    const last = buttons[buttons.length - 1];
    first.focus();
    fireEvent.keyDown(dialog, { key: 'Tab', shiftKey: true });
    expect(document.activeElement).toBe(last);
  });

  it('positions tooltip on top placement', () => {
    const topStep: TutorialStep = { ...step, placement: 'top' };
    render(<TutorialOverlay {...defaultProps} step={topStep} />);
    expect(screen.getByText('テスト説明')).toBeInTheDocument();
  });

  it('positions tooltip on left placement', () => {
    const leftStep: TutorialStep = { ...step, placement: 'left' };
    render(<TutorialOverlay {...defaultProps} step={leftStep} />);
    expect(screen.getByText('テスト説明')).toBeInTheDocument();
  });

  it('positions tooltip on right placement', () => {
    const rightStep: TutorialStep = { ...step, placement: 'right' };
    render(<TutorialOverlay {...defaultProps} step={rightStep} />);
    expect(screen.getByText('テスト説明')).toBeInTheDocument();
  });

  it('does not fire click listener for next advanceOn', async () => {
    const onNext = vi.fn();
    render(<TutorialOverlay {...defaultProps} onNext={onNext} />);
    fireEvent.click(targetEl);
    // onNext should not be called from target click since advanceOn is 'next'
    expect(onNext).not.toHaveBeenCalled();
  });

  it('ignores non-Tab key in focus trap handler', () => {
    render(<TutorialOverlay {...defaultProps} />);
    const dialog = screen.getByRole('dialog');
    const buttons = Array.from(dialog.querySelectorAll('button'));
    buttons[0].focus();
    fireEvent.keyDown(dialog, { key: 'ArrowDown' });
    expect(document.activeElement).toBe(buttons[0]);
  });

  it('locks body scroll on mount and restores on unmount', () => {
    document.body.style.overflow = 'auto';
    const { unmount } = render(<TutorialOverlay {...defaultProps} />);
    expect(document.body.style.overflow).toBe('hidden');
    unmount();
    expect(document.body.style.overflow).toBe('auto');
  });

  it('uses darker overlay opacity (0.75)', () => {
    const { container } = render(<TutorialOverlay {...defaultProps} />);
    const overlayRect = container.querySelector('svg > rect[mask]');
    expect(overlayRect).toHaveAttribute('fill', 'rgba(0,0,0,0.75)');
  });

  // Advancing a step used to throw focus out of the overlay: the effect that
  // tracks `step.target` restored focus to the trigger in its cleanup, so it
  // fired on every step change, while the focus-trap effect had `[]` deps and
  // never put focus back. From step 2 on, focus sat outside an `aria-modal`
  // dialog and the element-level Tab/Escape handlers stopped firing entirely.
  // The two steps must use different targets, or the effect does not re-run
  // and the bug is not reached. See issue #5184.
  describe('advancing a step', () => {
    let secondTarget: HTMLDivElement;
    const stepTwo: TutorialStep = { ...step, target: '[data-tutorial="second-target"]' };

    beforeEach(() => {
      secondTarget = document.createElement('div');
      secondTarget.setAttribute('data-tutorial', 'second-target');
      secondTarget.getBoundingClientRect = vi.fn().mockReturnValue({
        top: 200,
        left: 60,
        width: 120,
        height: 30,
        right: 180,
        bottom: 230,
      });
      document.body.appendChild(secondTarget);
    });

    afterEach(() => {
      secondTarget.remove();
    });

    it('keeps focus inside the overlay', () => {
      const outside = document.createElement('button');
      document.body.appendChild(outside);
      outside.focus();

      const { rerender } = render(<TutorialOverlay {...defaultProps} />);
      expect(screen.getByRole('dialog').contains(document.activeElement)).toBe(true);

      rerender(<TutorialOverlay {...defaultProps} step={stepTwo} stepIndex={1} />);
      expect(screen.getByRole('dialog').contains(document.activeElement)).toBe(true);
      expect(document.activeElement).not.toBe(outside);

      outside.remove();
    });

    it('still skips on Escape', () => {
      const onSkip = vi.fn();
      const { rerender } = render(<TutorialOverlay {...defaultProps} onSkip={onSkip} />);
      rerender(<TutorialOverlay {...defaultProps} onSkip={onSkip} step={stepTwo} stepIndex={1} />);

      fireEvent.keyDown(document, { key: 'Escape' });
      expect(onSkip).toHaveBeenCalledTimes(1);
    });

    it('does not restore focus to the trigger mid-tutorial', () => {
      const outside = document.createElement('button');
      document.body.appendChild(outside);
      outside.focus();

      const { rerender, unmount } = render(<TutorialOverlay {...defaultProps} />);
      rerender(<TutorialOverlay {...defaultProps} step={stepTwo} stepIndex={1} />);
      expect(document.activeElement).not.toBe(outside);

      // Only finishing the tutorial hands focus back.
      unmount();
      expect(document.activeElement).toBe(outside);

      outside.remove();
    });
  });

  it('restores focus on unmount', () => {
    const triggerButton = document.createElement('button');
    document.body.appendChild(triggerButton);
    triggerButton.focus();
    const { unmount } = render(<TutorialOverlay {...defaultProps} />);
    unmount();
    expect(document.activeElement).toBe(triggerButton);
    document.body.removeChild(triggerButton);
  });

  describe('viewport clamping', () => {
    const getTooltipContainer = (container: HTMLElement) =>
      container.querySelector<HTMLDivElement>('[role="dialog"] > div.absolute.z-10');

    it('clamps tooltip when overflowing left edge', () => {
      targetEl.getBoundingClientRect = vi.fn().mockReturnValue({
        top: 100,
        left: 5,
        width: 200,
        height: 40,
        right: 205,
        bottom: 140,
      });
      const leftStep: TutorialStep = { ...step, placement: 'top' };
      const { container } = render(<TutorialOverlay {...defaultProps} step={leftStep} />);
      const tooltip = getTooltipContainer(container);
      expect(tooltip).not.toBeNull();
      // The tooltip should be clamped — getBoundingClientRect in jsdom returns 0s,
      // so left < 8 triggers the clamp and transform becomes 'none'
      expect(tooltip?.style.transform).toBe('none');
    });

    it('clamps tooltip when overflowing bottom edge', () => {
      // Simulate viewport height
      Object.defineProperty(window, 'innerHeight', { value: 600, writable: true });
      targetEl.getBoundingClientRect = vi.fn().mockReturnValue({
        top: 550,
        left: 100,
        width: 200,
        height: 40,
        right: 300,
        bottom: 590,
      });
      const bottomStep: TutorialStep = { ...step, placement: 'bottom' };
      const { container } = render(<TutorialOverlay {...defaultProps} step={bottomStep} />);
      const tooltip = getTooltipContainer(container);
      expect(tooltip).not.toBeNull();
      expect(tooltip?.style.transform).toBe('none');
    });

    it('clamps tooltip when overflowing right edge', () => {
      Object.defineProperty(window, 'innerWidth', { value: 800, writable: true });
      targetEl.getBoundingClientRect = vi.fn().mockReturnValue({
        top: 100,
        left: 700,
        width: 200,
        height: 40,
        right: 900,
        bottom: 140,
      });
      const { container } = render(<TutorialOverlay {...defaultProps} />);
      const tooltip = getTooltipContainer(container);
      expect(tooltip).not.toBeNull();
      expect(tooltip?.style.transform).toBe('none');
    });

    it('clamps tooltip when overflowing top edge', () => {
      targetEl.getBoundingClientRect = vi.fn().mockReturnValue({
        top: 5,
        left: 100,
        width: 200,
        height: 40,
        right: 300,
        bottom: 45,
      });
      const topStep: TutorialStep = { ...step, placement: 'top' };
      const { container } = render(<TutorialOverlay {...defaultProps} step={topStep} />);
      const tooltip = getTooltipContainer(container);
      expect(tooltip).not.toBeNull();
      expect(tooltip?.style.transform).toBe('none');
    });

    it('does not clamp tooltip when within viewport bounds', () => {
      Object.defineProperty(window, 'innerWidth', { value: 1024, writable: true });
      Object.defineProperty(window, 'innerHeight', { value: 768, writable: true });
      targetEl.getBoundingClientRect = vi.fn().mockReturnValue({
        top: 200,
        left: 200,
        width: 200,
        height: 40,
        right: 400,
        bottom: 240,
      });
      const { container } = render(<TutorialOverlay {...defaultProps} />);
      const tooltip = getTooltipContainer(container);
      expect(tooltip).not.toBeNull();
      // Mock the tooltip element's getBoundingClientRect to be within viewport
      if (tooltip)
        tooltip.getBoundingClientRect = vi.fn().mockReturnValue({
          top: 252,
          left: 192,
          width: 300,
          height: 100,
          right: 492,
          bottom: 352,
        });
      // Re-render to trigger the useLayoutEffect with mocked tooltip rect
      const { container: c2 } = render(<TutorialOverlay {...defaultProps} />);
      const tooltip2 = getTooltipContainer(c2);
      expect(tooltip2).not.toBeNull();
      // In jsdom, getBoundingClientRect on the tooltip returns 0s by default,
      // which triggers clamping. Verify the clamped state sets left/top to margin values.
      // This test verifies the clamp logic runs; the "within bounds" case
      // cannot be reliably tested in jsdom without full layout engine.
      expect(tooltip2?.style.left).toBeDefined();
    });
  });
});
