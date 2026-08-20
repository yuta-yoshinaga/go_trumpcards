import { render, screen, waitFor } from '@testing-library/react';
import { renderToStaticMarkup } from 'react-dom/server';
import { describe, expect, it } from 'vitest';
import { LiveAnnouncement } from './LiveAnnouncement';

describe('LiveAnnouncement', () => {
  // **これが本題で、DOM を見ただけでは証明できない。**RTL の render は effect を
  // 流してしまうので、最初のコミットの中身を見るには effect の走らない
  // 静的レンダリングを使う。領域だけが出て、本文は入っていないこと ——
  // これが崩れると、領域と本文が同時に現れて読み上げられなくなる。
  it('paints the region before it paints the message', () => {
    const html = renderToStaticMarkup(<LiveAnnouncement message="choose a direction" />);
    expect(html).toContain('role="status"');
    expect(html).not.toContain('choose a direction');
  });

  // 言うことが無いうちから領域が居ること。
  it('renders the region even with nothing to say', () => {
    render(<LiveAnnouncement message="" />);
    expect(screen.getByRole('status')).toHaveTextContent('');
  });

  // 本文が付くとき、領域は**同じ DOM ノードのまま**であること。作り直すと
  // 「既存の領域が変化した」ではなく「領域が現れた」になり、読み上げられない。
  it('reuses the same node when the message arrives', () => {
    const { rerender } = render(<LiveAnnouncement message="" />);
    const before = screen.getByRole('status');
    rerender(<LiveAnnouncement message="choose a direction" />);
    const after = screen.getByRole('status');
    expect(after).toBe(before);
    expect(after).toHaveTextContent('choose a direction');
  });

  // role と aria-live は対。role="alert" は暗黙に assertive なので、
  // 片方だけ切り替えると宣言が食い違う (GameMessageBox と同じ形にした)。
  it('is polite by default and assertive on request', () => {
    render(<LiveAnnouncement message="hi" />);
    const polite = screen.getByRole('status');
    expect(polite).toHaveAttribute('aria-live', 'polite');
  });

  it('switches role and aria-live together when assertive', () => {
    render(<LiveAnnouncement message="hi" assertive />);
    const alert = screen.getByRole('alert');
    expect(alert).toHaveAttribute('aria-live', 'assertive');
    expect(screen.queryByRole('status')).not.toBeInTheDocument();
  });

  it('clears when the message goes away', async () => {
    const { rerender } = render(<LiveAnnouncement message="choose a direction" />);
    rerender(<LiveAnnouncement message="" />);
    await waitFor(() => expect(screen.getByRole('status')).toHaveTextContent(''));
  });

  // 読み上げ専用。見た目には出ない。
  it('stays visually hidden', () => {
    render(<LiveAnnouncement message="hi" />);
    expect(screen.getByRole('status')).toHaveClass('sr-only');
  });
});
