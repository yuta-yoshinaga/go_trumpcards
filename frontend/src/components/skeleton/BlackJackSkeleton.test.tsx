import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { BlackJackSkeleton } from './BlackJackSkeleton';

describe('BlackJackSkeleton', () => {
  it('renders skeleton structure', () => {
    render(<BlackJackSkeleton />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
    expect(screen.getByTestId('skeleton').getAttribute('role')).toBe('status');
  });
});
