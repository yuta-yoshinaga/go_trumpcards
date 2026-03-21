import { describe, expect, it } from 'bun:test';
import { render, screen } from '@testing-library/react';
import { MemorySkeleton } from './MemorySkeleton';

describe('MemorySkeleton', () => {
  it('renders skeleton structure', () => {
    render(<MemorySkeleton />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
    expect(screen.getByTestId('skeleton').getAttribute('aria-busy')).toBe('true');
  });
});
