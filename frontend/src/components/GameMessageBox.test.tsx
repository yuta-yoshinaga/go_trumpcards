import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { GameMessageBox } from './GameMessageBox';

describe('GameMessageBox', () => {
  it('renders null when message is undefined and alwaysVisible is false', () => {
    const { container } = render(<GameMessageBox message={undefined} />);
    expect(container.innerHTML).toBe('');
  });

  it('renders null when message is empty string and alwaysVisible is false', () => {
    const { container } = render(<GameMessageBox message="" />);
    expect(container.innerHTML).toBe('');
  });

  it('renders message text when message is provided', () => {
    render(<GameMessageBox message="テスト結果" />);
    expect(screen.getByText('テスト結果')).toBeInTheDocument();
  });

  it('has role="status" and aria-live="polite" by default', () => {
    render(<GameMessageBox message="勝ちました" />);
    const el = screen.getByRole('status');
    expect(el).toBeInTheDocument();
    expect(el).toHaveAttribute('aria-live', 'polite');
  });

  it('has role="alert" and aria-live="assertive" when severity is alert', () => {
    render(<GameMessageBox message="バスト！" severity="alert" />);
    const el = screen.getByRole('alert');
    expect(el).toBeInTheDocument();
    expect(el).toHaveAttribute('aria-live', 'assertive');
  });

  it('renders empty div when message is undefined and alwaysVisible is true', () => {
    const { container } = render(<GameMessageBox message={undefined} alwaysVisible />);
    const div = container.firstChild as HTMLElement;
    expect(div).toBeInTheDocument();
    expect(div.textContent).toBe('');
    expect(div.className).toContain('glass-panel');
  });

  it('renders message text when alwaysVisible is true and message is provided', () => {
    render(<GameMessageBox message="勝ちました" alwaysVisible />);
    expect(screen.getByText('勝ちました')).toBeInTheDocument();
  });

  // **平坦なドット付きキーが本当に解決するか。**`messageCode` の下は
  // `"scopone.errHandIndexOutOfRange"` という**1 本の平坦なキー**で、i18next の
  // 既定の keySeparator は `.` ── 入れ子として引きに行って外す形なら、生の
  // 識別子がそのまま両言語で画面に出る。ここは推論ではなく実際に描画して見る
  // (#6457、レビュー指摘で Web に同じバグが残っていた)。
  it('resolves a flat dotted messageCode into real text, params included', () => {
    render(
      <GameMessageBox
        message="scopone.errHandIndexOutOfRange"
        messageCode="scopone.errHandIndexOutOfRange"
        messageParams={{ idx: '7' }}
      />,
    );
    const box = screen.getByRole('status');
    expect(box).toHaveTextContent('手札の番号 7 は範囲外です。');
    // 生の識別子も未置換のプレースホルダも残らない。
    expect(box.textContent).not.toContain('scopone.');
    expect(box.textContent).not.toContain('{{');
  });

  it('translates messageCode when translation exists', () => {
    render(<GameMessageBox message="fallback" messageCode="blackjack.result.draw" />);
    expect(screen.getByText('引き分けです。')).toBeInTheDocument();
  });

  it('falls back to message when messageCode has no translation', () => {
    render(<GameMessageBox message="fallback text" messageCode="nonexistent.code" />);
    expect(screen.getByText('fallback text')).toBeInTheDocument();
  });

  it('translates messageCode with messageParams', () => {
    render(<GameMessageBox message="fallback" messageCode="doubt.result.cpuWin" messageParams={{ cpuId: '2' }} />);
    expect(screen.getByText('ゲーム終了！ CPU 2の勝ち！')).toBeInTheDocument();
  });

  it('translates messageCode without messageParams using default empty object', () => {
    render(<GameMessageBox message="fallback" messageCode="blackjack.result.win" />);
    expect(screen.getByText('あなたの勝ちです。')).toBeInTheDocument();
  });
});
