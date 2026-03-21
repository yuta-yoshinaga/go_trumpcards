import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { AnimatedCardBack } from './AnimatedCardBack';

vi.mock('../../hooks/useReducedMotion', () => ({
  useReducedMotion: vi.fn(() => false),
}));

import { useReducedMotion } from '../../hooks/useReducedMotion';

describe('AnimatedCardBack', () => {
  it('renders animated wrapper when motion is enabled', () => {
    vi.mocked(useReducedMotion).mockReturnValue(false);
    render(<AnimatedCardBack />);
    expect(screen.getByTestId('animated-card-back')).toBeInTheDocument();
  });

  it('renders plain CardBack when reduced motion is preferred', () => {
    vi.mocked(useReducedMotion).mockReturnValue(true);
    render(<AnimatedCardBack />);
    expect(screen.queryByTestId('animated-card-back')).not.toBeInTheDocument();
    expect(screen.getByRole('img')).toBeInTheDocument();
  });

  it('passes width and style to CardBack', () => {
    vi.mocked(useReducedMotion).mockReturnValue(true);
    render(<AnimatedCardBack width={60} style={{ opacity: 0.5 }} />);
    const img = screen.getByRole('img');
    expect(img).toHaveStyle({ width: '60px', opacity: '0.5' });
  });

  it('renders as button when onClick is provided in reduced mode', () => {
    vi.mocked(useReducedMotion).mockReturnValue(true);
    const onClick = vi.fn();
    render(<AnimatedCardBack onClick={onClick} ariaLabel="flip card" />);
    expect(screen.getByRole('button', { name: 'flip card' })).toBeInTheDocument();
  });

  it('renders as button when onClick is provided in animated mode', () => {
    vi.mocked(useReducedMotion).mockReturnValue(false);
    const onClick = vi.fn();
    render(<AnimatedCardBack onClick={onClick} ariaLabel="flip card" />);
    expect(screen.getByRole('button', { name: 'flip card' })).toBeInTheDocument();
  });

  it('applies custom dealDelay', () => {
    vi.mocked(useReducedMotion).mockReturnValue(false);
    render(<AnimatedCardBack dealDelay={0.3} />);
    expect(screen.getByTestId('animated-card-back')).toBeInTheDocument();
  });
});
