import { act, fireEvent, render, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { CpuActionToast } from './CpuActionToast';

const mockT = vi.fn((key: string) => key);
vi.mock('react-i18next', async () => ({
  ...(await vi.importActual('react-i18next')),
  useTranslation: () => ({ t: mockT }),
}));

const mockReduced = vi.fn(() => false);
vi.mock('../hooks/useReducedMotion', () => ({
  useReducedMotion: () => mockReduced(),
}));

describe('CpuActionToast', () => {
  beforeEach(() => {
    vi.useFakeTimers();
    mockT.mockClear();
    mockReduced.mockReturnValue(false);
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('renders nothing when actions is undefined', () => {
    const { container } = render(<CpuActionToast actions={undefined} />);
    expect(container.firstChild).toBeNull();
  });

  it('renders nothing when actions is empty', () => {
    const { container } = render(<CpuActionToast actions={[]} />);
    expect(container.firstChild).toBeNull();
  });

  it('renders toast when actions appear', () => {
    const actions = [{ playerIdx: 1, action: 2, amount: 0 }];
    render(<CpuActionToast actions={actions} />);
    expect(screen.getByRole('status')).toBeInTheDocument();
    expect(mockT).toHaveBeenCalledWith('player.player', { idx: 1 });
  });

  it('auto-dismisses after 5 seconds', () => {
    const actions = [{ playerIdx: 1, action: 2, amount: 0 }];
    render(<CpuActionToast actions={actions} />);
    expect(screen.getByRole('status')).toBeInTheDocument();

    act(() => {
      vi.advanceTimersByTime(4999);
    });
    expect(screen.getByRole('status')).toBeInTheDocument();

    act(() => {
      vi.advanceTimersByTime(1);
    });
    expect(screen.queryByRole('status')).toBeNull();
  });

  it('resets timer when new actions arrive', () => {
    const actions1 = [{ playerIdx: 1, action: 2, amount: 0 }];
    const { rerender } = render(<CpuActionToast actions={actions1} />);
    expect(screen.getByRole('status')).toBeInTheDocument();

    act(() => {
      vi.advanceTimersByTime(3000);
    });

    const actions2 = [
      { playerIdx: 1, action: 2, amount: 0 },
      { playerIdx: 2, action: 3, amount: 40 },
    ];
    rerender(<CpuActionToast actions={actions2} />);

    // 4 more seconds (7s total) — still visible because timer reset
    act(() => {
      vi.advanceTimersByTime(4000);
    });
    expect(screen.getByRole('status')).toBeInTheDocument();

    // Past the 5s window since last update
    act(() => {
      vi.advanceTimersByTime(1001);
    });
    expect(screen.queryByRole('status')).toBeNull();
  });

  it('shows amount when present', () => {
    const actions = [{ playerIdx: 2, action: 3, amount: 100 }];
    render(<CpuActionToast actions={actions} />);
    expect(screen.getByText(/100/)).toBeInTheDocument();
  });

  it('has aria-live polite attribute', () => {
    const actions = [{ playerIdx: 1, action: 2, amount: 0 }];
    render(<CpuActionToast actions={actions} />);
    expect(screen.getByRole('status')).toHaveAttribute('aria-live', 'polite');
  });

  it('dismisses when the close button is clicked', () => {
    const actions = [{ playerIdx: 1, action: 2, amount: 0 }];
    render(<CpuActionToast actions={actions} />);
    fireEvent.click(screen.getByRole('button', { name: 'button.dismiss' }));
    expect(screen.queryByRole('status')).toBeNull();
  });

  it('dismisses when Escape is pressed', () => {
    const actions = [{ playerIdx: 1, action: 2, amount: 0 }];
    render(<CpuActionToast actions={actions} />);
    act(() => {
      fireEvent.keyDown(window, { key: 'Escape' });
    });
    expect(screen.queryByRole('status')).toBeNull();
  });

  it('omits the slide-down animation when prefers-reduced-motion is set', () => {
    mockReduced.mockReturnValue(true);
    const actions = [{ playerIdx: 1, action: 2, amount: 0 }];
    render(<CpuActionToast actions={actions} />);
    expect(screen.getByRole('status').className).not.toContain('slideDown');
  });

  it('does not dismiss on Escape while an aria-modal dialog is open', () => {
    const dialog = document.createElement('div');
    dialog.setAttribute('role', 'dialog');
    dialog.setAttribute('aria-modal', 'true');
    document.body.appendChild(dialog);

    const actions = [{ playerIdx: 1, action: 2, amount: 0 }];
    render(<CpuActionToast actions={actions} />);
    act(() => {
      fireEvent.keyDown(window, { key: 'Escape' });
    });
    expect(screen.getByRole('status')).toBeInTheDocument();

    document.body.removeChild(dialog);
  });

  it('restores focus to the trigger element when dismissed', () => {
    const trigger = document.createElement('button');
    trigger.textContent = 'trigger';
    document.body.appendChild(trigger);
    trigger.focus();
    expect(document.activeElement).toBe(trigger);

    const actions = [{ playerIdx: 1, action: 2, amount: 0 }];
    const { rerender } = render(<CpuActionToast actions={undefined} />);
    rerender(<CpuActionToast actions={actions} />);
    expect(screen.getByRole('status')).toBeInTheDocument();

    act(() => {
      fireEvent.keyDown(window, { key: 'Escape' });
    });
    expect(document.activeElement).toBe(trigger);
    document.body.removeChild(trigger);
  });

  it('applies the slide-down animation when reduced motion is off', () => {
    mockReduced.mockReturnValue(false);
    const actions = [{ playerIdx: 1, action: 2, amount: 0 }];
    render(<CpuActionToast actions={actions} />);
    expect(screen.getByRole('status').className).toContain('slideDown');
  });
});
