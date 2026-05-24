import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { BidProgressBar } from './BidProgressBar';

describe('BidProgressBar', () => {
  it('renders nothing for unset (negative) bid', () => {
    const { container } = render(<BidProgressBar bid={-1} tricksWon={0} />);
    expect(container.firstChild).toBeNull();
  });

  it('renders one segment per bid', () => {
    render(<BidProgressBar bid={4} tricksWon={1} testId="b" />);
    const bar = screen.getByTestId('b');
    expect(bar.querySelectorAll('[class*="rounded-sm"]').length).toBe(4);
    expect(bar).toHaveAttribute('aria-valuemax', '4');
    expect(bar).toHaveAttribute('aria-valuenow', '1');
  });

  it('fills all segments and tints success when bid is exactly met', () => {
    render(<BidProgressBar bid={3} tricksWon={3} testId="b" />);
    const bar = screen.getByTestId('b');
    const segments = Array.from(bar.querySelectorAll('[class*="rounded-sm"]'));
    expect(segments.every((s) => s.className.includes('bg-ds-success'))).toBe(true);
  });

  it('shows overtricks pill when tricksWon exceeds bid', () => {
    render(<BidProgressBar bid={3} tricksWon={5} />);
    expect(screen.getByText('+2')).toBeInTheDocument();
  });

  it('renders an error-tinted single segment for nil bid that has been broken', () => {
    render(<BidProgressBar bid={0} tricksWon={1} testId="b" />);
    const bar = screen.getByTestId('b');
    const segments = Array.from(bar.querySelectorAll('[class*="rounded-sm"]'));
    expect(segments.length).toBe(1);
    expect(segments[0].className).toContain('bg-ds-error');
    expect(screen.queryByText(/^\+/)).not.toBeInTheDocument();
  });

  it('renders a blank single segment for an unbroken nil bid', () => {
    render(<BidProgressBar bid={0} tricksWon={0} testId="b" />);
    const bar = screen.getByTestId('b');
    const segments = Array.from(bar.querySelectorAll('[class*="rounded-sm"]'));
    expect(segments.length).toBe(1);
    expect(segments[0].className).not.toContain('bg-ds-error');
  });
});
