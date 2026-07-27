import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { CardNavShortcutsPanel } from './CardNavShortcutsPanel';

/**
 * Open the panel before asserting on its rows: they are mounted on demand, so a
 * collapsed panel contributes no text to the page (which is the point — see the
 * component docs and issue #4369).
 */
function openPanel() {
  // By element, not by title text: KeyboardShortcutsPanel takes an arbitrary
  // title, so keying on the shared i18n string would not work for its own tests.
  const summary = document.querySelector('summary');
  if (!summary) throw new Error('no <summary> to open — the panel did not render');
  fireEvent.click(summary);
}

/**
 * useCardKeyboardNav's bindings are fixed for every page that uses it — number
 * keys pick a card, Enter confirms, Escape clears — so the advertised copy lives
 * in common.json once instead of being restated in each game's namespace, where
 * it could drift from the hook. See issue #4369.
 */
describe('CardNavShortcutsPanel', () => {
  it('lists the number-key range, Enter and Esc', () => {
    render(<CardNavShortcutsPanel />);
    openPanel();
    expect(screen.getByText('キーボードショートカット')).toBeInTheDocument();
    for (const key of ['1', '0', 'Enter', 'Esc']) {
      expect(screen.getByText(key)).toBeInTheDocument();
    }
    expect(screen.getByText('数字キーで手札のカードを選択')).toBeInTheDocument();
    expect(screen.getByText('選択したカードを出す')).toBeInTheDocument();
    expect(screen.getByText('カードの選択を解除')).toBeInTheDocument();
  });

  it('describes number keys as playing directly when the page passes onDirectPlay', () => {
    // useCardKeyboardNav plays the card immediately instead of toggling
    // selection when onDirectPlay is supplied, so Enter/Esc are meaningless
    // there and advertising them would be a lie.
    render(<CardNavShortcutsPanel directPlay />);
    openPanel();
    expect(screen.getByText('数字キーで手札のカードをそのまま出す')).toBeInTheDocument();
    expect(screen.queryByText('Enter')).not.toBeInTheDocument();
    expect(screen.queryByText('Esc')).not.toBeInTheDocument();
  });

  it('stays collapsed by default so it costs no vertical space', () => {
    // The mobile viewport budget is tight (issue #4373): this must not expand
    // the page until the player asks for it.
    const { container } = render(<CardNavShortcutsPanel />);
    expect(container.querySelector('details')).not.toHaveAttribute('open');
  });

  it('forwards extra props so pages can attach a test id', () => {
    render(<CardNavShortcutsPanel data-testid="kbd-panel" />);
    openPanel();
    expect(screen.getByTestId('kbd-panel')).toBeInTheDocument();
  });

  it('appends extra shortcuts a page binds on top of the card-nav set', () => {
    // Pages that also bind their own keys (e.g. `n` for "next trick") pass them
    // here rather than rendering a second panel.
    render(<CardNavShortcutsPanel extra={[{ keys: ['n'], description: '次へ' }]} />);
    openPanel();
    expect(screen.getByText('n')).toBeInTheDocument();
    expect(screen.getByText('次へ')).toBeInTheDocument();
  });
});
