import { render } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { ScrollFadeHint } from './ScrollFadeHint';

describe('ScrollFadeHint', () => {
  it('renders aria-hidden gradient div', () => {
    const { container } = render(<ScrollFadeHint />);
    const el = container.querySelector('[aria-hidden="true"]');
    expect(el).toBeInTheDocument();
  });

  it('has sm:hidden class to hide on desktop', () => {
    const { container } = render(<ScrollFadeHint />);
    expect(container.firstElementChild?.className).toContain('sm:hidden');
  });

  it('has pointer-events-none to not block clicks', () => {
    const { container } = render(<ScrollFadeHint />);
    expect(container.firstElementChild?.className).toContain('pointer-events-none');
  });

  it('uses from-black/50 gradient for dark background visibility', () => {
    const { container } = render(<ScrollFadeHint />);
    expect(container.firstElementChild?.className).toContain('from-black/50');
  });

  it('has w-8 width for sufficient gradient visibility', () => {
    const { container } = render(<ScrollFadeHint />);
    expect(container.firstElementChild?.className).toContain('w-8');
  });
});
