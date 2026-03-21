import { describe, expect, it } from 'bun:test';
import { render, screen } from '@testing-library/react';
import { OldMaidSkeleton } from './OldMaidSkeleton';

describe('OldMaidSkeleton', () => {
  it('renders skeleton structure', () => {
    render(<OldMaidSkeleton />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
    expect(screen.getByTestId('skeleton').getAttribute('aria-busy')).toBe('true');
  });
});
