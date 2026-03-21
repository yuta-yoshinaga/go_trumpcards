import { describe, expect, it } from 'bun:test';
import { render, screen } from '@testing-library/react';
import { CrazyEightsSkeleton } from './CrazyEightsSkeleton';

describe('CrazyEightsSkeleton', () => {
  it('renders skeleton structure', () => {
    render(<CrazyEightsSkeleton />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
    expect(screen.getByTestId('skeleton').getAttribute('aria-busy')).toBe('true');
  });
});
