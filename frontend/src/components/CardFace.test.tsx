import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { CardFace } from './CardFace';

const wizardCard: Card = {
  design: 'JOKER',
  value: 1,
  glyph: '✦',
  label: 'Wizard',
  color: 'purple',
  deck: 'wizard',
};

describe('CardFace', () => {
  it('renders as a div with role img and no image src', () => {
    render(<CardFace card={wizardCard} />);
    const el = screen.getByRole('img');
    expect(el.tagName).toBe('DIV');
    expect(el).not.toHaveAttribute('src');
  });

  it('renders the label in both corners and the glyph in the center', () => {
    render(<CardFace card={wizardCard} />);
    expect(screen.getAllByText('Wizard')).toHaveLength(2);
    expect(screen.getByText('✦')).toBeInTheDocument();
  });

  it('builds an accessible aria-label from label and glyph', () => {
    render(<CardFace card={wizardCard} />);
    expect(screen.getByRole('img')).toHaveAttribute('aria-label', 'Wizard ✦');
  });

  it('applies the color token as ink color', () => {
    render(<CardFace card={wizardCard} />);
    // purple → #7C3AED
    expect(screen.getByRole('img')).toHaveStyle({ color: '#7C3AED' });
  });

  it('defaults ink to near-black for an unknown color token', () => {
    render(<CardFace card={{ ...wizardCard, color: 'chartreuse' }} />);
    expect(screen.getByRole('img')).toHaveStyle({ color: '#1A1A1A' });
  });

  it('falls back to value when label is absent and first char when glyph is absent', () => {
    render(<CardFace card={{ design: 'JOKER', value: 7, deck: 'x' }} />);
    // label defaults to "7"; both corners show it, glyph defaults to "7"
    expect(screen.getAllByText('7').length).toBeGreaterThanOrEqual(2);
  });

  it('applies width prop and derived 2:3 height', () => {
    render(<CardFace card={wizardCard} width={60} />);
    expect(screen.getByRole('img')).toHaveStyle({ width: '60px', height: '90px' });
  });

  it('applies custom className and style', () => {
    render(<CardFace card={wizardCard} className="my-face" style={{ opacity: 0.4 }} />);
    const el = screen.getByRole('img');
    expect(el).toHaveClass('my-face');
    expect(el).toHaveStyle({ opacity: '0.4' });
  });

  it('disables text selection for iOS Safari drag', () => {
    render(<CardFace card={wizardCard} />);
    expect(screen.getByRole('img').getAttribute('style') ?? '').toContain('user-select: none');
  });
});
