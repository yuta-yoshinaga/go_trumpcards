import { render } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { KbdBadge } from './KbdBadge';

describe('KbdBadge', () => {
  it('renders the label inside a <kbd>', () => {
    const { container } = render(<KbdBadge label="Space" />);
    const kbd = container.querySelector('kbd');
    expect(kbd).not.toBeNull();
    expect(kbd?.textContent).toBe('Space');
  });
});
