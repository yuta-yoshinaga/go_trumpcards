import { describe, expect, it } from 'bun:test';
import { render, screen } from '@testing-library/react';
import { DoubtSkeleton } from './DoubtSkeleton';

describe('DoubtSkeleton', () => {
  it('renders skeleton structure', () => {
    render(<DoubtSkeleton />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
    expect(screen.getByTestId('skeleton').getAttribute('aria-busy')).toBe('true');
  });
});
