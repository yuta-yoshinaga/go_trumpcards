import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { CardBackFan } from './CardBackFan';

describe('CardBackFan', () => {
  it('renders the requested backs in an overlapping fan', () => {
    render(<CardBackFan count={4} width={40} />);
    const fan = screen.getByTestId('card-back-fan');
    expect(fan.querySelectorAll('img')).toHaveLength(4);
    expect(fan.children[1]).toHaveStyle({ marginLeft: '-30px' });
  });
  it('renders an empty container for zero cards', () => {
    render(<CardBackFan count={0} width={40} />);
    expect(screen.getByTestId('card-back-fan')).toBeEmptyDOMElement();
  });
});
