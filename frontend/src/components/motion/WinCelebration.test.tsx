import { afterEach, beforeEach, describe, expect, it, vi } from 'bun:test';
import { render, screen } from '@testing-library/react';
import * as useReducedMotionModule from '../../hooks/useReducedMotion';
import { WinCelebration } from './WinCelebration';

describe('WinCelebration', () => {
  let spy: ReturnType<typeof vi.spyOn>;

  beforeEach(() => {
    spy = vi.spyOn(useReducedMotionModule, 'useReducedMotion').mockReturnValue(false);
  });

  afterEach(() => {
    spy.mockRestore();
  });

  it('renders particles when show is true and motion is enabled', () => {
    spy.mockReturnValue(false);
    render(<WinCelebration show={true} />);
    expect(screen.getByTestId('win-celebration')).toBeInTheDocument();
    expect(screen.getByTestId('win-celebration').getAttribute('aria-hidden')).toBe('true');
  });

  it('renders nothing when show is false', () => {
    spy.mockReturnValue(false);
    const { container } = render(<WinCelebration show={false} />);
    expect(container.innerHTML).toBe('');
  });

  it('renders nothing when reduced motion is preferred', () => {
    spy.mockReturnValue(true);
    const { container } = render(<WinCelebration show={true} />);
    expect(container.innerHTML).toBe('');
  });

  it('renders nothing when both show is false and reduced motion', () => {
    spy.mockReturnValue(true);
    const { container } = render(<WinCelebration show={false} />);
    expect(container.innerHTML).toBe('');
  });
});
