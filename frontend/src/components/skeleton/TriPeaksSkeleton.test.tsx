import { render } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { TriPeaksSkeleton } from './TriPeaksSkeleton';

describe('TriPeaksSkeleton', () => {
  it('renders without crashing', () => {
    render(<TriPeaksSkeleton />);
    // Should render skeleton cards (pulse animation elements)
    const pulseElements = document.querySelectorAll('.animate-pulse');
    expect(pulseElements.length).toBeGreaterThan(0);
  });

  it('renders the correct number of skeleton card rows', () => {
    const { container } = render(<TriPeaksSkeleton />);
    // 4 rows of cards + 1 stock/waste row
    const rows = container.querySelectorAll('.flex.justify-center');
    expect(rows.length).toBeGreaterThanOrEqual(4);
  });
});
