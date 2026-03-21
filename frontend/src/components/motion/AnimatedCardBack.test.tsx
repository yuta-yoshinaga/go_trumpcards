import { afterEach, beforeEach, describe, expect, it, vi } from 'bun:test';
import { render, screen } from '@testing-library/react';
import * as useReducedMotionModule from '../../hooks/useReducedMotion';
import { AnimatedCardBack } from './AnimatedCardBack';

describe('AnimatedCardBack', () => {
  let spy: ReturnType<typeof vi.spyOn>;

  beforeEach(() => {
    spy = vi.spyOn(useReducedMotionModule, 'useReducedMotion').mockReturnValue(false);
  });

  afterEach(() => {
    spy.mockRestore();
  });

  it('renders animated wrapper when motion is enabled', () => {
    spy.mockReturnValue(false);
    render(<AnimatedCardBack />);
    expect(screen.getByTestId('animated-card-back')).toBeInTheDocument();
  });

  it('renders plain CardBack when reduced motion is preferred', () => {
    spy.mockReturnValue(true);
    render(<AnimatedCardBack />);
    expect(screen.queryByTestId('animated-card-back')).not.toBeInTheDocument();
    expect(screen.getByRole('img')).toBeInTheDocument();
  });

  it('passes width and style to CardBack', () => {
    spy.mockReturnValue(true);
    render(<AnimatedCardBack width={60} style={{ opacity: 0.5 }} />);
    const img = screen.getByRole('img');
    expect(img).toHaveStyle({ width: '60px', opacity: '0.5' });
  });

  it('renders as button when onClick is provided in reduced mode', () => {
    spy.mockReturnValue(true);
    const onClick = vi.fn();
    render(<AnimatedCardBack onClick={onClick} ariaLabel="flip card" />);
    expect(screen.getByRole('button', { name: 'flip card' })).toBeInTheDocument();
  });

  it('renders as button when onClick is provided in animated mode', () => {
    spy.mockReturnValue(false);
    const onClick = vi.fn();
    render(<AnimatedCardBack onClick={onClick} ariaLabel="flip card" />);
    expect(screen.getByRole('button', { name: 'flip card' })).toBeInTheDocument();
  });

  it('applies custom dealDelay', () => {
    spy.mockReturnValue(false);
    render(<AnimatedCardBack dealDelay={0.3} />);
    expect(screen.getByTestId('animated-card-back')).toBeInTheDocument();
  });
});
