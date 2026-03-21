import { describe, expect, it } from 'bun:test';
import { render, screen } from '@testing-library/react';
import { BlackJackSkeleton } from './BlackJackSkeleton';

describe('BlackJackSkeleton', () => {
  it('renders skeleton structure', () => {
    render(<BlackJackSkeleton />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
    expect(screen.getByTestId('skeleton').getAttribute('aria-busy')).toBe('true');
  });
});
