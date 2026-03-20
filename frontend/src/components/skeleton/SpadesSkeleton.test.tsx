import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { SpadesSkeleton } from './SpadesSkeleton';

describe('SpadesSkeleton', () => {
  it('renders skeleton structure', () => {
    render(<SpadesSkeleton />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
    expect(screen.getByTestId('skeleton').getAttribute('aria-busy')).toBe('true');
  });
});
