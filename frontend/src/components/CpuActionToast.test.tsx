import { act, fireEvent, render, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { CpuActionToast } from './CpuActionToast';

const mockT = vi.fn((key: string) => key);
vi.mock('react-i18next', async () => ({
  ...(await vi.importActual('react-i18next')),
  useTranslation: () => ({ t: mockT }),
}));

describe('CpuActionToast', () => {
  beforeEach(() => {
    vi.useFakeTimers();
    mockT.mockClear();
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

  it('auto-dismisses after the long duration (6s)', () => {
    const actions = [{ playerIdx: 1, action: 2, amount: 0 }];
    render(<CpuActionToast actions={actions} />);
    expect(screen.getByRole('status')).toBeInTheDocument();

    act(() => {
      vi.advanceTimersByTime(5999);
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

    // 5s more (8s total) — still visible because the 6s timer reset on the update
    act(() => {
      vi.advanceTimersByTime(5000);
    });
    expect(screen.getByRole('status')).toBeInTheDocument();

    // Past the 6s window since the last update
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

  it('has a 44x44px close button (WCAG 2.5.5)', () => {
    const actions = [{ playerIdx: 1, action: 2, amount: 0 }];
    render(<CpuActionToast actions={actions} />);
    const btn = screen.getByRole('button', { name: 'button.dismiss' });
    expect(btn.className).toContain('min-h-[44px]');
    expect(btn.className).toContain('min-w-[44px]');
  });

  it('dismisses when Escape is pressed', () => {
    const actions = [{ playerIdx: 1, action: 2, amount: 0 }];
    render(<CpuActionToast actions={actions} />);
    act(() => {
      fireEvent.keyDown(window, { key: 'Escape' });
    });
    expect(screen.queryByRole('status')).toBeNull();
  });

  it('uses an opaque surface background (DESIGN.md Opacity rule)', () => {
    const actions = [{ playerIdx: 1, action: 2, amount: 0 }];
    render(<CpuActionToast actions={actions} />);
    const cls = screen.getByRole('status').className;
    expect(cls).toContain('bg-ds-surface-elevated');
    expect(cls).not.toContain('bg-black/');
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

  it('does not clobber triggerRef when new actions arrive while toast is already visible', () => {
    const initialTrigger = document.createElement('button');
    initialTrigger.textContent = 'initial-trigger';
    document.body.appendChild(initialTrigger);
    initialTrigger.focus();

    const actions1 = [{ playerIdx: 1, action: 2, amount: 0 }];
    const { rerender } = render(<CpuActionToast actions={undefined} />);
    rerender(<CpuActionToast actions={actions1} />);

    // Focus the toast's close button to simulate user tabbing into it
    const closeBtn = screen.getByRole('button', { name: 'button.dismiss' });
    closeBtn.focus();
    expect(document.activeElement).toBe(closeBtn);

    // New action while toast already visible → triggerRef must NOT be overwritten
    const actions2 = [
      { playerIdx: 1, action: 2, amount: 0 },
      { playerIdx: 2, action: 3, amount: 40 },
    ];
    rerender(<CpuActionToast actions={actions2} />);

    // Blur focus so activeElement is body (simulating toast unmount on dismiss)
    closeBtn.blur();
    act(() => {
      fireEvent.keyDown(window, { key: 'Escape' });
    });
    expect(document.activeElement).toBe(initialTrigger);
    document.body.removeChild(initialTrigger);
  });

  it('does not hijack focus when user has moved focus elsewhere before dismissal', () => {
    const trigger = document.createElement('button');
    trigger.textContent = 'trigger';
    const other = document.createElement('button');
    other.textContent = 'other';
    document.body.appendChild(trigger);
    document.body.appendChild(other);
    trigger.focus();

    const actions = [{ playerIdx: 1, action: 2, amount: 0 }];
    const { rerender } = render(<CpuActionToast actions={undefined} />);
    rerender(<CpuActionToast actions={actions} />);

    // User moves focus to a different element while toast is visible
    other.focus();
    expect(document.activeElement).toBe(other);

    act(() => {
      fireEvent.keyDown(window, { key: 'Escape' });
    });
    // Focus must stay on `other`, not jump back to trigger
    expect(document.activeElement).toBe(other);
    document.body.removeChild(trigger);
    document.body.removeChild(other);
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

  it('always applies the slide-down animation (reduced motion is handled by CSS)', () => {
    const actions = [{ playerIdx: 1, action: 2, amount: 0 }];
    render(<CpuActionToast actions={actions} />);
    expect(screen.getByRole('status').className).toContain('slideDown');
  });
});
