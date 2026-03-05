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

  it('applies safe-area paddingBottom style', () => {
    const { container } = render(
      <GameFooter className="bg-green-800">
        <span>content</span>
      </GameFooter>,
    );
    const div = container.firstChild as HTMLElement;
    expect(div.style.paddingBottom).toContain('env(safe-area-inset-bottom)');
  });

  it('includes shrink-0 and border-t base classes', () => {
    const { container } = render(
      <GameFooter className="bg-green-800">
        <span>content</span>
      </GameFooter>,
    );
    const div = container.firstChild as HTMLElement;
    expect(div.className).toContain('shrink-0');
    expect(div.className).toContain('border-t');
  });

  it('merges provided className', () => {
    const { container } = render(
      <GameFooter className="bg-[#005a00] border-white/15 px-4 py-3">
        <span>content</span>
      </GameFooter>,
    );
    const div = container.firstChild as HTMLElement;
    expect(div.className).toContain('bg-[#005a00]');
    expect(div.className).toContain('border-white/15');
    expect(div.className).toContain('px-4');
    expect(div.className).toContain('py-3');
  });
});
