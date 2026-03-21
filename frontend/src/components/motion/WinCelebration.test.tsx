import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { WinCelebration } from './WinCelebration';

vi.mock('../../hooks/useReducedMotion', () => ({
  useReducedMotion: vi.fn(() => false),
}));

import { useReducedMotion } from '../../hooks/useReducedMotion';

describe('WinCelebration', () => {
  it('renders particles when show is true and motion is enabled', () => {
    vi.mocked(useReducedMotion).mockReturnValue(false);
    render(<WinCelebration show={true} />);
    expect(screen.getByTestId('win-celebration')).toBeInTheDocument();
    expect(screen.getByTestId('win-celebration').getAttribute('aria-hidden')).toBe('true');
  });

  it('renders nothing when show is false', () => {
    vi.mocked(useReducedMotion).mockReturnValue(false);
    const { container } = render(<WinCelebration show={false} />);
    expect(container.innerHTML).toBe('');
  });

  it('renders nothing when reduced motion is preferred', () => {
    vi.mocked(useReducedMotion).mockReturnValue(true);
    const { container } = render(<WinCelebration show={true} />);
    expect(container.innerHTML).toBe('');
  });

  it('renders nothing when both show is false and reduced motion', () => {
    vi.mocked(useReducedMotion).mockReturnValue(true);
    const { container } = render(<WinCelebration show={false} />);
    expect(container.innerHTML).toBe('');
  });
});
