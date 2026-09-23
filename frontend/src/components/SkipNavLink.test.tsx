import { fireEvent, render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { SkipNavLink } from './SkipNavLink';

const originalScrollIntoView = Element.prototype.scrollIntoView;

afterEach(() => {
  if (originalScrollIntoView) {
    Element.prototype.scrollIntoView = originalScrollIntoView;
  } else {
    Reflect.deleteProperty(Element.prototype, 'scrollIntoView');
  }
});

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

  it('focuses and scrolls to the target without changing the URL hash', () => {
    const scrollIntoView = vi.fn();
    Element.prototype.scrollIntoView = scrollIntoView;
    window.history.replaceState(null, '', '/');
    render(
      <>
        <SkipNavLink targetId="content" label="Skip" />
        <main id="content" tabIndex={-1} />
      </>,
    );
    const initialHash = window.location.hash;

    fireEvent.click(screen.getByRole('link', { name: 'Skip' }));

    expect(screen.getByRole('main')).toHaveFocus();
    expect(window.location.hash).toBe(initialHash);
    expect(scrollIntoView).toHaveBeenCalledWith({ block: 'start' });
  });

  it('does nothing when the target element is missing', () => {
    render(<SkipNavLink targetId="missing" label="Skip" />);

    expect(() => fireEvent.click(screen.getByRole('link', { name: 'Skip' }))).not.toThrow();
  });
});
