import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { DaifugoSkeleton } from './DaifugoSkeleton';

describe('DaifugoSkeleton', () => {
  it('renders skeleton structure', () => {
    render(<DaifugoSkeleton />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
    expect(screen.getByTestId('skeleton').getAttribute('aria-busy')).toBe('true');
  });
});
