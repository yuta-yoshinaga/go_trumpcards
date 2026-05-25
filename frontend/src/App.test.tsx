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
    render(<RouteSuspenseFallback />);
    // SkeletonBar marks itself aria-hidden so screen readers stay on the
    // sr-only loading label. Querying by that contract — not by class
    // name — keeps the test stable if styling changes.
    const status = screen.getByRole('status');
    expect(status.querySelector('[aria-hidden="true"]')).not.toBeNull();
  });

  it('exposes the localized loading label to screen readers', () => {
    render(<RouteSuspenseFallback />);
    // Accept either locale's translation of skeleton.loading; the test
    // setup defaults to ja but a contributor's env may differ.
    expect(screen.getByText(/^(Loading…|読み込み中…)$/)).toBeInTheDocument();
  });
});
