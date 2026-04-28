import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { GameSkeleton } from './GameSkeleton';

describe('GameSkeleton', () => {
  it('renders skeleton shell with data-testid and aria-busy', () => {
    render(
      <GameSkeleton bgClass="bg-game-bg-green" footerClassName="bg-green-dark px-4 py-3" footer={<div>footer</div>}>
        <div>body</div>
      </GameSkeleton>,
    );
    const el = screen.getByTestId('skeleton');
    expect(el).toBeInTheDocument();
    expect(el.getAttribute('aria-busy')).toBe('true');
  });

  it('exposes a status live region with a localized loading label for screen readers', () => {
    render(
      <GameSkeleton bgClass="bg-game-bg-green" footerClassName="bg-green-dark" footer={<div>f</div>}>
        <div>b</div>
      </GameSkeleton>,
    );
    const el = screen.getByTestId('skeleton');
    expect(el.getAttribute('role')).toBe('status');
    expect(el.getAttribute('aria-live')).toBe('polite');
    expect(el.getAttribute('aria-label')).toBe('読み込み中…');
    expect(screen.getByText('読み込み中…')).toBeInTheDocument();
  });

  it('applies bgClass to outer wrapper', () => {
    render(
      <GameSkeleton bgClass="bg-game-bg-blue" footerClassName="bg-blue-dark" footer={<div>f</div>}>
        <div>b</div>
      </GameSkeleton>,
    );
    expect(screen.getByTestId('skeleton').className).toContain('bg-game-bg-blue');
  });

  it('renders body and footer content', () => {
    render(
      <GameSkeleton bgClass="bg-x" footerClassName="bg-y" footer={<span>my-footer</span>}>
        <span>my-body</span>
      </GameSkeleton>,
    );
    expect(screen.getByText('my-body')).toBeInTheDocument();
    expect(screen.getByText('my-footer')).toBeInTheDocument();
  });

  it('uses default bodyClassName when not specified', () => {
    render(
      <GameSkeleton bgClass="bg-x" footerClassName="bg-y" footer={<div>f</div>}>
        <div data-testid="body-content">b</div>
      </GameSkeleton>,
    );
    const bodyDiv = screen.getByTestId('body-content').parentElement;
    expect(bodyDiv).not.toBeNull();
    expect(bodyDiv?.className).toContain('pt-3 px-4');
  });

  it('uses custom bodyClassName when specified', () => {
    render(
      <GameSkeleton bgClass="bg-x" bodyClassName="p-4" footerClassName="bg-y" footer={<div>f</div>}>
        <div data-testid="body-content">b</div>
      </GameSkeleton>,
    );
    const bodyDiv = screen.getByTestId('body-content').parentElement;
    expect(bodyDiv).not.toBeNull();
    expect(bodyDiv?.className).toContain('p-4');
    expect(bodyDiv?.className).not.toContain('pt-3');
  });

  it('applies footerClassName to GameFooter', () => {
    render(
      <GameSkeleton
        bgClass="bg-x"
        footerClassName="bg-game-bg-green-dark border-white/20 px-4 py-3"
        footer={<div data-testid="footer-content">f</div>}
      >
        <div>b</div>
      </GameSkeleton>,
    );
    const footer = screen.getByTestId('footer-content').closest('footer');
    expect(footer).not.toBeNull();
    expect(footer?.className).toContain('bg-game-bg-green-dark');
    expect(footer?.className).toContain('border-white/20');
  });
});
