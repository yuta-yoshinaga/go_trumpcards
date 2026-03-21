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
