import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { FreeCellSkeleton } from './FreeCellSkeleton';

describe('FreeCellSkeleton', () => {
  it('renders skeleton structure', () => {
    render(<FreeCellSkeleton />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
    expect(screen.getByTestId('skeleton').getAttribute('aria-busy')).toBe('true');
  });
});
