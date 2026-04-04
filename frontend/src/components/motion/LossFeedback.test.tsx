import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { LossFeedback } from './LossFeedback';

vi.mock('../../hooks/useReducedMotion', () => ({
  useReducedMotion: vi.fn(() => false),
}));

import { useReducedMotion } from '../../hooks/useReducedMotion';

describe('LossFeedback', () => {
  it('renders vignette overlay when show is true', () => {
    vi.mocked(useReducedMotion).mockReturnValue(false);
    render(<LossFeedback show={true} />);
    expect(screen.getByTestId('loss-feedback')).toBeInTheDocument();
  });

  it('renders nothing when show is false', () => {
    vi.mocked(useReducedMotion).mockReturnValue(false);
    const { container } = render(<LossFeedback show={false} />);
    expect(container.innerHTML).toBe('');
  });

  it('includes ARIA live region with loss text', () => {
    vi.mocked(useReducedMotion).mockReturnValue(false);
    render(<LossFeedback show={true} />);
    expect(screen.getByRole('status')).toBeInTheDocument();
  });

  it('renders text banner in reduced motion mode', () => {
    vi.mocked(useReducedMotion).mockReturnValue(true);
    render(<LossFeedback show={true} />);
    const feedback = screen.getByTestId('loss-feedback');
    expect(feedback).toBeInTheDocument();
    expect(feedback).toHaveAttribute('role', 'status');
  });

  it('renders nothing when show is false in reduced motion mode', () => {
    vi.mocked(useReducedMotion).mockReturnValue(true);
    const { container } = render(<LossFeedback show={false} />);
    expect(container.innerHTML).toBe('');
  });
});
