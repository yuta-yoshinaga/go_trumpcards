import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { CountdownBar } from './CountdownBar';

// Mock useReducedMotion
vi.mock('../../hooks/useReducedMotion', () => ({
  useReducedMotion: vi.fn(() => false),
}));

import { useReducedMotion } from '../../hooks/useReducedMotion';

describe('CountdownBar', () => {
  it('renders a progressbar with correct aria attributes', () => {
    render(<CountdownBar remaining={7} total={10} />);
    const bar = screen.getByRole('progressbar');
    expect(bar).toHaveAttribute('aria-valuenow', '7');
    expect(bar).toHaveAttribute('aria-valuemax', '10');
    expect(bar).toHaveAttribute('aria-valuemin', '0');
  });

  it('renders bar width proportional to remaining/total', () => {
    const { container } = render(<CountdownBar remaining={5} total={10} />);
    const inner = container.querySelector('[data-testid="countdown-bar-fill"]') as HTMLElement;
    expect(inner).not.toBeNull();
    expect(inner.style.width).toBe('50%');
  });

  it('applies success color when remaining > 6', () => {
    const { container } = render(<CountdownBar remaining={8} total={10} />);
    const inner = container.querySelector('[data-testid="countdown-bar-fill"]') as HTMLElement;
    expect(inner.className).toContain('bg-ds-success');
  });

  it('applies warning color when remaining is 4-6', () => {
    const { container } = render(<CountdownBar remaining={5} total={10} />);
    const inner = container.querySelector('[data-testid="countdown-bar-fill"]') as HTMLElement;
    expect(inner.className).toContain('bg-ds-warning');
  });

  it('applies error color when remaining <= 3', () => {
    const { container } = render(<CountdownBar remaining={2} total={10} />);
    const inner = container.querySelector('[data-testid="countdown-bar-fill"]') as HTMLElement;
    expect(inner.className).toContain('bg-ds-error');
  });

  it('includes transition style when reduced motion is off', () => {
    const { container } = render(<CountdownBar remaining={5} total={10} />);
    const inner = container.querySelector('[data-testid="countdown-bar-fill"]') as HTMLElement;
    expect(inner.style.transition).toContain('width');
  });

  it('omits transition style when reduced motion is on', () => {
    vi.mocked(useReducedMotion).mockReturnValue(true);
    const { container } = render(<CountdownBar remaining={5} total={10} />);
    const inner = container.querySelector('[data-testid="countdown-bar-fill"]') as HTMLElement;
    expect(inner.style.transition).toBe('');
    vi.mocked(useReducedMotion).mockReturnValue(false);
  });

  it('exposes a role=timer sr-only region that announces only at 5s marks and the final 3s', () => {
    const { rerender } = render(<CountdownBar remaining={5} total={10} label="残り 5 秒" />);
    const timer = screen.getByTestId('countdown-sr-timer');
    expect(timer).toHaveAttribute('role', 'timer');
    expect(timer).toHaveAttribute('aria-live', 'polite');
    expect(timer).toHaveAttribute('aria-atomic', 'true');
    expect(timer).toHaveTextContent('残り 5 秒'); // 5 % 5 === 0

    rerender(<CountdownBar remaining={2} total={10} label="残り 2 秒" />);
    expect(screen.getByTestId('countdown-sr-timer')).toHaveTextContent('残り 2 秒'); // ≤3s
  });

  it('keeps the timer region empty on non-throttled ticks (no per-second readout)', () => {
    render(<CountdownBar remaining={7} total={10} label="残り 7 秒" />);
    // 7 is neither ≤3 nor a multiple of 5 → nothing spoken this tick.
    expect(screen.getByTestId('countdown-sr-timer')).toHaveTextContent('');
  });

  it('renders the visible countdown text as aria-hidden with ds-warning color', () => {
    render(<CountdownBar remaining={7} total={10} label="残り 7 秒" />);
    // At 7 the sr timer is empty, so this matches only the visible text node.
    const visible = screen.getByText('残り 7 秒');
    expect(visible).toHaveAttribute('aria-hidden', 'true');
    expect(visible.className).toContain('text-ds-warning');
  });

  it('does not render the timer region when label is omitted', () => {
    render(<CountdownBar remaining={7} total={10} />);
    expect(screen.queryByTestId('countdown-sr-timer')).toBeNull();
  });

  it('sets aria-label on progressbar from label prop', () => {
    render(<CountdownBar remaining={5} total={10} label="残り 5 秒" />);
    const bar = screen.getByRole('progressbar');
    expect(bar).toHaveAttribute('aria-label', '残り 5 秒');
  });

  it('uses default aria-label on progressbar when label is omitted', () => {
    render(<CountdownBar remaining={5} total={10} />);
    const bar = screen.getByRole('progressbar');
    expect(bar).toHaveAttribute('aria-label', 'Countdown');
  });
});
