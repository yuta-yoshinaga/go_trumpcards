import { describe, expect, it } from 'bun:test';
import { render, screen } from '@testing-library/react';
import { FreeCellSkeleton } from './FreeCellSkeleton';

describe('FreeCellSkeleton', () => {
  it('renders skeleton structure', () => {
    render(<FreeCellSkeleton />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
    expect(screen.getByTestId('skeleton').getAttribute('aria-busy')).toBe('true');
  });
});
