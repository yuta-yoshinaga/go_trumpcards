import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { MemorySkeleton } from './MemorySkeleton';

describe('MemorySkeleton', () => {
  it('renders skeleton structure', () => {
    render(<MemorySkeleton />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
    expect(screen.getByTestId('skeleton').getAttribute('aria-busy')).toBe('true');
  });
});
