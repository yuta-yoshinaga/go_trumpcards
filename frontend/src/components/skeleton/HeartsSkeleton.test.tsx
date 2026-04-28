import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { HeartsSkeleton } from './HeartsSkeleton';

describe('HeartsSkeleton', () => {
  it('renders skeleton structure', () => {
    render(<HeartsSkeleton />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
    expect(screen.getByTestId('skeleton').getAttribute('role')).toBe('status');
  });
});
