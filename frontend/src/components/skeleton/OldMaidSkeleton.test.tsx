import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { OldMaidSkeleton } from './OldMaidSkeleton';

describe('OldMaidSkeleton', () => {
  it('renders skeleton structure', () => {
    render(<OldMaidSkeleton />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
    expect(screen.getByTestId('skeleton').getAttribute('role')).toBe('status');
  });
});
