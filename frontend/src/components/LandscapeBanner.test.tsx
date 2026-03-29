import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { LandscapeBanner } from './LandscapeBanner';

describe('LandscapeBanner', () => {
  it('renders the message text', () => {
    render(<LandscapeBanner message="Rotate for easier play" />);
    expect(screen.getByText('Rotate for easier play')).toBeInTheDocument();
  });

  it('renders rotation icon as aria-hidden', () => {
    const { container } = render(<LandscapeBanner message="Test" />);
    const icon = container.querySelector('[aria-hidden="true"]');
    expect(icon).toBeInTheDocument();
  });

  it('has hidden portrait:flex sm:hidden classes for responsive visibility', () => {
    const { container } = render(<LandscapeBanner message="Test" />);
    const banner = container.firstElementChild;
    expect(banner?.className).toContain('hidden');
    expect(banner?.className).toContain('portrait:flex');
    expect(banner?.className).toContain('sm:hidden');
  });

  it('has animate-pulse-once class on icon for limited attention', () => {
    const { container } = render(<LandscapeBanner message="Test" />);
    const icon = container.querySelector('svg');
    expect(icon?.getAttribute('class')).toContain('animate-pulse-once');
  });
});
