import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { GameFooter } from './GameFooter';

describe('GameFooter', () => {
  it('renders children', () => {
    render(
      <GameFooter className="bg-ds-success">
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

  it('includes shrink-0 and border-t base classes', () => {
    render(
      <GameFooter className="bg-ds-success">
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
      <GameFooter className="bg-game-bg-green-bright-dark border-white/20 px-4 py-3">
        <span>content</span>
      </GameFooter>,
    );
    const footer = screen.getByRole('contentinfo');
    expect(footer.className).toContain('bg-game-bg-green-bright-dark');
    expect(footer.className).toContain('border-white/20');
    expect(footer.className).toContain('px-4');
    expect(footer.className).toContain('py-3');
  });

  // The footer is `shrink-0`, so whatever height its content wants, it takes —
  // and the sibling scroll region is what gives way. Measured at 375x667 across
  // all 219 game pages, 26 pages were left with under 80px for their content and
  // the tallest footer was 558px (`watten`), 84% of the viewport. Capping it at
  // 45vh guarantees at least 102px of content on every page that has both a
  // footer and a scroll region. See issue #4373.
  it('caps its height and scrolls internally on mobile', () => {
    render(
      <GameFooter>
        <span>content</span>
      </GameFooter>,
    );
    const footer = screen.getByRole('contentinfo');
    expect(footer.className).toContain('max-h-[45vh]');
    expect(footer.className).toContain('overflow-y-auto');
  });

  // The cap exists to protect a 667px-tall phone; on a laptop it would only add a
  // pointless inner scrollbar, so it must be lifted from `sm` up.
  it('lifts the cap from the sm breakpoint up', () => {
    render(
      <GameFooter>
        <span>content</span>
      </GameFooter>,
    );
    const footer = screen.getByRole('contentinfo');
    expect(footer.className).toContain('sm:max-h-none');
    expect(footer.className).toContain('sm:overflow-y-visible');
  });

  it('keeps the legacy footer classes when actions are omitted', () => {
    render(
      <GameFooter>
        <span>content</span>
      </GameFooter>,
    );
    const footer = screen.getByRole('contentinfo');
    expect(footer.className).toContain('max-h-[45vh] overflow-y-auto sm:max-h-none sm:overflow-y-visible');
    expect(screen.queryByTestId('game-footer-content')).not.toBeInTheDocument();
  });

  it('separates scrollable content from persistent actions', () => {
    render(
      <GameFooter actions={<button type="button">進む</button>}>
        <span>手札</span>
      </GameFooter>,
    );
    const footer = screen.getByRole('contentinfo');
    const content = screen.getByTestId('game-footer-content');
    const actions = screen.getByTestId('game-footer-actions');
    expect(content).toContainElement(screen.getByText('手札'));
    expect(actions).toContainElement(screen.getByRole('button', { name: '進む' }));
    expect(content).not.toContainElement(actions);
    expect(footer.className).toContain('sm:max-h-[50vh]');
    expect(footer.className).not.toContain('overflow-y-auto');
    expect(content.className).toContain('min-h-0');
    expect(content.className).toContain('overflow-y-auto');
  });

  it('keeps the actions layout but omits the action row when actions is null', () => {
    render(
      <GameFooter actions={null}>
        <span>手札</span>
      </GameFooter>,
    );
    const footer = screen.getByRole('contentinfo');
    expect(footer.className).toContain('sm:max-h-[50vh]');
    expect(screen.getByTestId('game-footer-content')).toContainElement(screen.getByText('手札'));
    expect(screen.queryByTestId('game-footer-actions')).not.toBeInTheDocument();
  });
});
