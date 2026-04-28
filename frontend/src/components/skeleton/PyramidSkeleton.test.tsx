import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { PyramidSkeleton } from './PyramidSkeleton';

describe('PyramidSkeleton', () => {
  it('renders skeleton structure', () => {
    render(<PyramidSkeleton />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
    expect(screen.getByTestId('skeleton').getAttribute('role')).toBe('status');
  });
});
