import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { OmahaSkeleton } from './OmahaSkeleton';

describe('OmahaSkeleton', () => {
  it('renders skeleton structure', () => {
    render(<OmahaSkeleton />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
    expect(screen.getByTestId('skeleton').getAttribute('role')).toBe('status');
  });
});
