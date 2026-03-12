import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { LoadingSpinner } from './LoadingSpinner';

describe('LoadingSpinner', () => {
  it('renders nothing when loading is false', () => {
    const { container } = render(<LoadingSpinner loading={false} />);
    expect(container.firstChild).toBeNull();
  });

  it('renders spinner with status role and accessible name when loading is true', () => {
    render(<LoadingSpinner loading={true} />);
    expect(screen.getByRole('status', { name: '処理中...' })).toBeInTheDocument();
  });
});
