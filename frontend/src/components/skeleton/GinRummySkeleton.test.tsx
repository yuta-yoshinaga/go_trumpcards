import { describe, expect, it } from 'bun:test';
import { render, screen } from '@testing-library/react';
import { GinRummySkeleton } from './GinRummySkeleton';

describe('GinRummySkeleton', () => {
  it('renders skeleton structure', () => {
    render(<GinRummySkeleton />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
    expect(screen.getByTestId('skeleton').getAttribute('aria-busy')).toBe('true');
  });
});
