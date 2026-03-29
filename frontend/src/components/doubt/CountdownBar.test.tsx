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

  it('applies green color when remaining > 6', () => {
    const { container } = render(<CountdownBar remaining={8} total={10} />);
    const inner = container.querySelector('[data-testid="countdown-bar-fill"]') as HTMLElement;
    expect(inner.className).toContain('bg-green-500');
  });

  it('applies yellow color when remaining is 4-6', () => {
    const { container } = render(<CountdownBar remaining={5} total={10} />);
    const inner = container.querySelector('[data-testid="countdown-bar-fill"]') as HTMLElement;
    expect(inner.className).toContain('bg-yellow-500');
  });

  it('applies red color when remaining <= 3', () => {
    const { container } = render(<CountdownBar remaining={2} total={10} />);
    const inner = container.querySelector('[data-testid="countdown-bar-fill"]') as HTMLElement;
    expect(inner.className).toContain('bg-red-500');
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

  it('renders label text in aria-live region when provided', () => {
    render(<CountdownBar remaining={3} total={10} label="残り 3 秒" />);
    const liveRegion = screen.getByText('残り 3 秒');
    expect(liveRegion).toHaveAttribute('aria-live', 'assertive');
    expect(liveRegion).toHaveAttribute('aria-atomic', 'true');
  });

  it('does not render label region when label is omitted', () => {
    const { container } = render(<CountdownBar remaining={7} total={10} />);
    expect(container.querySelector('[aria-live="assertive"]')).toBeNull();
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
