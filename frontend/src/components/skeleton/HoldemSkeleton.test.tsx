import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { HoldemSkeleton } from './HoldemSkeleton';

describe('HoldemSkeleton', () => {
  it('renders skeleton structure', () => {
    render(<HoldemSkeleton />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
    expect(screen.getByTestId('skeleton').getAttribute('aria-busy')).toBe('true');
  });
});
