import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { ActionBinding } from '../hooks/useActionKeyboardNav';
import { ActionShortcutsPanel } from './ActionShortcutsPanel';

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

const b = (key: string, label?: ActionBinding['label'], enabled?: boolean): ActionBinding => ({
  key,
  action: vi.fn(),
  label,
  ...(enabled === undefined ? {} : { enabled }),
});

/**
 * The panel is generated from the same bindings array that useActionKeyboardNav
 * binds, so an advertised key is by construction a bound key. See issue #4369.
 */
describe('ActionShortcutsPanel', () => {
  it('lists each labelled binding with its key and translated label', () => {
    render(<ActionShortcutsPanel bindings={[b('f', 'fold'), b('c', 'call')]} />);
    openPanel();
    expect(screen.getByText('f')).toBeInTheDocument();
    expect(screen.getByText('フォールド')).toBeInTheDocument();
    expect(screen.getByText('c')).toBeInTheDocument();
    expect(screen.getByText('コール')).toBeInTheDocument();
  });

  it('omits bindings with no label rather than showing a raw key name', () => {
    // A binding may be bound deliberately without being advertised; it must not
    // surface as an untranslated i18n key (the failure mode of #4374).
    render(<ActionShortcutsPanel bindings={[b('f', 'fold'), b('x')]} />);
    openPanel();
    expect(screen.getByText('f')).toBeInTheDocument();
    expect(screen.queryByText('x')).not.toBeInTheDocument();
  });

  it('omits bindings that are currently disabled', () => {
    // enabled:false means the key does nothing right now, so advertising it
    // would be telling the player about an action they cannot take.
    render(<ActionShortcutsPanel bindings={[b('f', 'fold'), b('d', 'doubledown', false)]} />);
    openPanel();
    expect(screen.getByText('f')).toBeInTheDocument();
    expect(screen.queryByText('ダブルダウン')).not.toBeInTheDocument();
  });

  it('renders nothing at all when no binding is advertisable', () => {
    // Rather than an empty <details> that opens onto blank space.
    const { container } = render(<ActionShortcutsPanel bindings={[b('x'), b('y')]} />);
    expect(container).toBeEmptyDOMElement();
  });

  it('stays collapsed by default so it costs no vertical space', () => {
    const { container } = render(<ActionShortcutsPanel bindings={[b('f', 'fold')]} />);
    expect(container.querySelector('details')).not.toHaveAttribute('open');
  });

  it('appends card-nav rows for pages that use both hooks', () => {
    render(<ActionShortcutsPanel bindings={[b('f', 'fold')]} includeCardNav />);
    openPanel();
    expect(screen.getByText('Enter')).toBeInTheDocument();
    expect(screen.getByText('選択したカードを出す')).toBeInTheDocument();
    expect(screen.getByText('フォールド')).toBeInTheDocument();
  });
});
