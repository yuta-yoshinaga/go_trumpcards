import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { KlondikeSkeleton } from './KlondikeSkeleton';

describe('KlondikeSkeleton', () => {
  it('renders skeleton structure', () => {
    render(<KlondikeSkeleton />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
    expect(screen.getByTestId('skeleton').getAttribute('role')).toBe('status');
  });
});
