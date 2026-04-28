import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { CrazyEightsSkeleton } from './CrazyEightsSkeleton';

describe('CrazyEightsSkeleton', () => {
  it('renders skeleton structure', () => {
    render(<CrazyEightsSkeleton />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
    expect(screen.getByTestId('skeleton').getAttribute('role')).toBe('status');
  });
});
