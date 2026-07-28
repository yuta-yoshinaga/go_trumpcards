import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { KeyboardShortcutsPanel } from './KeyboardShortcutsPanel';

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

describe('KeyboardShortcutsPanel', () => {
  const shortcuts = [
    { keys: ['1', '9'], description: 'カードを選択' },
    { keys: ['Enter'], description: '選択したカードを出す' },
  ];

  it('renders the title as a collapsible summary, closed by default', () => {
    // Deliberately does NOT open the panel — this asserts the closed state.
    render(<KeyboardShortcutsPanel title="キーボードショートカット" shortcuts={shortcuts} data-testid="wh-kbd" />);
    const panel = screen.getByTestId('wh-kbd');
    expect(panel.tagName).toBe('DETAILS');
    expect(panel).not.toHaveAttribute('open');
    // The title lives in the <summary>, so it is present even while collapsed.
    expect(screen.getByText('キーボードショートカット')).toBeInTheDocument();
  });

  it('is hidden below the sm breakpoint, where there is no keyboard to use', () => {
    // A 375px phone cannot press these shortcuts, and the panel costs 52px of
    // the mobile vertical budget on all 111 game pages (44px tap-target summary
    // + mt-2). Measured: hearts went 683 -> 735 when this shipped, crossing the
    // 667px viewport that #4373 tracks. Hiding it on mobile returns those pages
    // to their previous height and loses nothing.
    render(<KeyboardShortcutsPanel title="キーボードショートカット" shortcuts={shortcuts} data-testid="kbd" />);
    const panel = screen.getByTestId('kbd');
    expect(panel).toHaveClass('hidden');
    expect(panel).toHaveClass('sm:block');
  });

  it('contributes no shortcut text to the page while collapsed', () => {
    // The property that keeps 111 game pages' text unchanged: these rows name
    // the same actions as the buttons they describe, so leaving them mounted
    // made getByText ambiguous in unit tests and a strict-mode error in
    // Playwright. See issue #4369.
    render(<KeyboardShortcutsPanel title="キーボードショートカット" shortcuts={shortcuts} />);
    for (const s of shortcuts) {
      expect(screen.queryByText(s.description)).not.toBeInTheDocument();
    }
    openPanel();
    for (const s of shortcuts) {
      expect(screen.getByText(s.description)).toBeInTheDocument();
    }
  });

  it('lists every shortcut key and description', () => {
    render(<KeyboardShortcutsPanel title="ショートカット" shortcuts={shortcuts} />);
    openPanel();
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
