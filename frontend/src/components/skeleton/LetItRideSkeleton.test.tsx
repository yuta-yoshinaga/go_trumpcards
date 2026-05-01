import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { LetItRideSkeleton } from './LetItRideSkeleton';

describe('LetItRideSkeleton', () => {
  it('renders skeleton structure with role=status', () => {
    render(<LetItRideSkeleton />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
    expect(screen.getByTestId('skeleton').getAttribute('role')).toBe('status');
  });
});
