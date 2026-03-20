import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { GameFooter } from './GameFooter';

describe('GameFooter', () => {
  it('renders children', () => {
    render(
      <GameFooter className="bg-green-800">
        <button type="button">リセット</button>
      </GameFooter>,
    );
    expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument();
  });

  it('renders as a semantic footer element', () => {
    render(
      <GameFooter>
        <span>content</span>
      </GameFooter>,
    );
    expect(screen.getByRole('contentinfo')).toBeInTheDocument();
  });

  it('applies safe-area paddingBottom style', () => {
    render(
      <GameFooter className="bg-green-800">
        <span>content</span>
      </GameFooter>,
    );
    const footer = screen.getByRole('contentinfo');
    expect(footer.style.paddingBottom).toContain('env(safe-area-inset-bottom)');
  });

  it('includes shrink-0 and border-t base classes', () => {
    render(
      <GameFooter className="bg-green-800">
        <span>content</span>
      </GameFooter>,
    );
    const footer = screen.getByRole('contentinfo');
    expect(footer.className).toContain('shrink-0');
    expect(footer.className).toContain('border-t');
  });

  it('renders without className prop', () => {
    render(
      <GameFooter>
        <span>content</span>
      </GameFooter>,
    );
    const footer = screen.getByRole('contentinfo');
    expect(footer.className).toContain('shrink-0');
    expect(footer.className).toContain('border-t');
  });

  it('merges provided className', () => {
    render(
      <GameFooter className="bg-game-bg-green-bright-dark border-white/15 px-4 py-3">
        <span>content</span>
      </GameFooter>,
    );
    const footer = screen.getByRole('contentinfo');
    expect(footer.className).toContain('bg-game-bg-green-bright-dark');
    expect(footer.className).toContain('border-white/15');
    expect(footer.className).toContain('px-4');
    expect(footer.className).toContain('py-3');
  });
});
