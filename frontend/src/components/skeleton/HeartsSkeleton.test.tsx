import { describe, expect, it } from 'bun:test';
import { render, screen } from '@testing-library/react';
import { HeartsSkeleton } from './HeartsSkeleton';

describe('HeartsSkeleton', () => {
  it('renders skeleton structure', () => {
    render(<HeartsSkeleton />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
    expect(screen.getByTestId('skeleton').getAttribute('aria-busy')).toBe('true');
  });
});
