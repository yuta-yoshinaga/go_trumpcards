import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { KeyboardShortcutsPanel } from './KeyboardShortcutsPanel';

describe('KeyboardShortcutsPanel', () => {
  const shortcuts = [
    { keys: ['1', '9'], description: 'カードを選択' },
    { keys: ['Enter'], description: '選択したカードを出す' },
  ];

  it('renders the title as a collapsible summary, closed by default', () => {
    render(<KeyboardShortcutsPanel title="キーボードショートカット" shortcuts={shortcuts} data-testid="wh-kbd" />);
    const panel = screen.getByTestId('wh-kbd');
    expect(panel.tagName).toBe('DETAILS');
    expect(panel).not.toHaveAttribute('open');
    expect(screen.getByText('キーボードショートカット')).toBeInTheDocument();
  });

  it('lists every shortcut key and description', () => {
    render(<KeyboardShortcutsPanel title="ショートカット" shortcuts={shortcuts} />);
    expect(screen.getByText('カードを選択')).toBeInTheDocument();
    expect(screen.getByText('選択したカードを出す')).toBeInTheDocument();
    // Each key label is rendered inside its own <kbd> chip.
    const keys = screen.getAllByText(/^(1|9|Enter)$/);
    expect(keys).toHaveLength(3);
    for (const k of keys) {
      expect(k.tagName).toBe('KBD');
    }
  });
});
