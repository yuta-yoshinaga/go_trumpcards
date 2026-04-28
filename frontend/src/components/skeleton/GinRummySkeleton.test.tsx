import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { GinRummySkeleton } from './GinRummySkeleton';

describe('GinRummySkeleton', () => {
  it('renders skeleton structure', () => {
    render(<GinRummySkeleton />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
    expect(screen.getByTestId('skeleton').getAttribute('role')).toBe('status');
  });
});
