import { describe, expect, it } from 'bun:test';
import { render, screen } from '@testing-library/react';
import { SpadesSkeleton } from './SpadesSkeleton';

describe('SpadesSkeleton', () => {
  it('renders skeleton structure', () => {
    render(<SpadesSkeleton />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
    expect(screen.getByTestId('skeleton').getAttribute('aria-busy')).toBe('true');
  });
});
