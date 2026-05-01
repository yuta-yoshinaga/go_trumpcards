import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { SkeletonShell } from './SkeletonShell';

describe('SkeletonShell', () => {
  it('renders skeleton shell with data-testid and role=status', () => {
    render(
      <SkeletonShell bgClass="bg-game-bg-green" footerClassName="bg-green-dark px-4 py-3" footer={<div>footer</div>}>
        <div>body</div>
      </SkeletonShell>,
    );
    const el = screen.getByTestId('skeleton');
    expect(el).toBeInTheDocument();
    expect(el.getAttribute('role')).toBe('status');
  });

  it('exposes a localized loading label via an sr-only span inside the status region', () => {
    // role="status" already implies aria-live="polite". The sr-only span is the
    // accessible name source — assistive tech announces its text content when the
    // status region first mounts. We deliberately do not duplicate it as aria-label
    // (which would override the text) or set aria-busy (which can suppress the
    // announcement on some ATs that wait for aria-busy="false" before rendering).
    render(
      <SkeletonShell bgClass="bg-game-bg-green" footerClassName="bg-green-dark" footer={<div>f</div>}>
        <div>b</div>
      </SkeletonShell>,
    );
    const el = screen.getByTestId('skeleton');
    expect(el).not.toHaveAttribute('aria-label');
    expect(el).not.toHaveAttribute('aria-busy');
    expect(screen.getByText('読み込み中…')).toBeInTheDocument();
  });

  it('applies bgClass to outer wrapper', () => {
    render(
      <SkeletonShell bgClass="bg-game-bg-blue" footerClassName="bg-blue-dark" footer={<div>f</div>}>
        <div>b</div>
      </SkeletonShell>,
    );
    expect(screen.getByTestId('skeleton').className).toContain('bg-game-bg-blue');
  });

  it('renders body and footer content', () => {
    render(
      <SkeletonShell bgClass="bg-x" footerClassName="bg-y" footer={<span>my-footer</span>}>
        <span>my-body</span>
      </SkeletonShell>,
    );
    expect(screen.getByText('my-body')).toBeInTheDocument();
    expect(screen.getByText('my-footer')).toBeInTheDocument();
  });

  it('uses default bodyClassName when not specified', () => {
    render(
      <SkeletonShell bgClass="bg-x" footerClassName="bg-y" footer={<div>f</div>}>
        <div data-testid="body-content">b</div>
      </SkeletonShell>,
    );
    const bodyDiv = screen.getByTestId('body-content').parentElement;
    expect(bodyDiv).not.toBeNull();
    expect(bodyDiv?.className).toContain('pt-3 px-4');
  });

  it('uses custom bodyClassName when specified', () => {
    render(
      <SkeletonShell bgClass="bg-x" bodyClassName="p-4" footerClassName="bg-y" footer={<div>f</div>}>
        <div data-testid="body-content">b</div>
      </SkeletonShell>,
    );
    const bodyDiv = screen.getByTestId('body-content').parentElement;
    expect(bodyDiv).not.toBeNull();
    expect(bodyDiv?.className).toContain('p-4');
    expect(bodyDiv?.className).not.toContain('pt-3');
  });

  it('applies footerClassName to GameFooter', () => {
    render(
      <SkeletonShell
        bgClass="bg-x"
        footerClassName="bg-game-bg-green-dark border-white/20 px-4 py-3"
        footer={<div data-testid="footer-content">f</div>}
      >
        <div>b</div>
      </SkeletonShell>,
    );
    const footer = screen.getByTestId('footer-content').closest('footer');
    expect(footer).not.toBeNull();
    expect(footer?.className).toContain('bg-game-bg-green-dark');
    expect(footer?.className).toContain('border-white/20');
  });
});
