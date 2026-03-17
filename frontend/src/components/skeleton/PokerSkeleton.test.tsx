import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { PokerSkeleton } from './PokerSkeleton';

describe('PokerSkeleton', () => {
  it('renders skeleton structure', () => {
    render(<PokerSkeleton />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
    expect(screen.getByTestId('skeleton').getAttribute('aria-busy')).toBe('true');
  });
});
