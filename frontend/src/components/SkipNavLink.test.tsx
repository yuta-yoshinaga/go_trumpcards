import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { SkipNavLink } from './SkipNavLink';

describe('SkipNavLink', () => {
  it('renders a link with the given label and href', () => {
    render(<SkipNavLink targetId="main-content" label="Skip to main" />);
    const link = screen.getByRole('link', { name: 'Skip to main' });
    expect(link).toHaveAttribute('href', '#main-content');
  });

  it('has sr-only class for visual hiding', () => {
    render(<SkipNavLink targetId="content" label="Skip" />);
    const link = screen.getByRole('link', { name: 'Skip' });
    expect(link.className).toContain('sr-only');
  });

  it('receives focus when focused programmatically', () => {
    render(<SkipNavLink targetId="content" label="Skip" />);
    const link = screen.getByRole('link', { name: 'Skip' });

    link.focus();
    expect(link).toHaveFocus();
  });
});
