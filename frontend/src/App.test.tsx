import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { RouteSuspenseFallback } from './App';

describe('RouteSuspenseFallback', () => {
  it('preserves the role="status" + aria-busy contract for assistive tech', () => {
    render(<RouteSuspenseFallback />);
    const status = screen.getByRole('status');
    expect(status).toHaveAttribute('aria-busy', 'true');
  });

  it('renders a SkeletonBar so the visual channel is not blank during chunk download', () => {
    const { container } = render(<RouteSuspenseFallback />);
    // SkeletonBar carries animate-pulse so the user sees motion-based
    // structure forming; this assertion guards against a regression that
    // would silently drop the visual placeholder back to a blank div.
    const pulsing = container.querySelector('.animate-pulse');
    expect(pulsing).not.toBeNull();
  });

  it('exposes the localized loading label to screen readers', () => {
    render(<RouteSuspenseFallback />);
    // ja is the test setup locale; the key is "skeleton.loading".
    expect(screen.getByText('読み込み中…')).toBeInTheDocument();
  });
});
