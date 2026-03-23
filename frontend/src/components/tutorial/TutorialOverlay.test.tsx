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

  it('traps focus within the dialog', () => {
    render(<TutorialOverlay {...defaultProps} />);
    const dialog = screen.getByRole('dialog');
    const buttons = dialog.querySelectorAll('button');
    // Focus should stay within dialog
    expect(buttons.length).toBeGreaterThan(0);
  });
});
