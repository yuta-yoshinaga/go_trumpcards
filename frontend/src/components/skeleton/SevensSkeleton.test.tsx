import { describe, expect, it } from 'bun:test';
import { render, screen } from '@testing-library/react';
import { SevensSkeleton } from './SevensSkeleton';

describe('SevensSkeleton', () => {
  it('renders skeleton structure', () => {
    render(<SevensSkeleton />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
    expect(screen.getByTestId('skeleton').getAttribute('aria-busy')).toBe('true');
  });
});
