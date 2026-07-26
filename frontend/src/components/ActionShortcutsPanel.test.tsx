import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { ActionBinding } from '../hooks/useActionKeyboardNav';
import { ActionShortcutsPanel } from './ActionShortcutsPanel';

const b = (key: string, labelKey?: string, enabled?: boolean): ActionBinding => ({
  key,
  action: vi.fn(),
  labelKey,
  ...(enabled === undefined ? {} : { enabled }),
});

/**
 * The panel is generated from the same bindings array that useActionKeyboardNav
 * binds, so an advertised key is by construction a bound key. See issue #4369.
 */
describe('ActionShortcutsPanel', () => {
  it('lists each labelled binding with its key and translated label', () => {
    render(<ActionShortcutsPanel bindings={[b('f', 'kbd.action.fold'), b('c', 'kbd.action.call')]} />);
    expect(screen.getByText('f')).toBeInTheDocument();
    expect(screen.getByText('フォールド')).toBeInTheDocument();
    expect(screen.getByText('c')).toBeInTheDocument();
    expect(screen.getByText('コール')).toBeInTheDocument();
  });

  it('omits bindings with no labelKey rather than showing a raw key name', () => {
    // A binding may be bound deliberately without being advertised; it must not
    // surface as an untranslated i18n key (the failure mode of #4374).
    render(<ActionShortcutsPanel bindings={[b('f', 'kbd.action.fold'), b('x')]} />);
    expect(screen.getByText('f')).toBeInTheDocument();
    expect(screen.queryByText('x')).not.toBeInTheDocument();
  });

  it('omits bindings that are currently disabled', () => {
    // enabled:false means the key does nothing right now, so advertising it
    // would be telling the player about an action they cannot take.
    render(<ActionShortcutsPanel bindings={[b('f', 'kbd.action.fold'), b('d', 'kbd.action.doubledown', false)]} />);
    expect(screen.getByText('f')).toBeInTheDocument();
    expect(screen.queryByText('ダブルダウン')).not.toBeInTheDocument();
  });

  it('renders nothing at all when no binding is advertisable', () => {
    // Rather than an empty <details> that opens onto blank space.
    const { container } = render(<ActionShortcutsPanel bindings={[b('x'), b('y')]} />);
    expect(container).toBeEmptyDOMElement();
  });

  it('stays collapsed by default so it costs no vertical space', () => {
    const { container } = render(<ActionShortcutsPanel bindings={[b('f', 'kbd.action.fold')]} />);
    expect(container.querySelector('details')).not.toHaveAttribute('open');
  });

  it('appends card-nav rows for pages that use both hooks', () => {
    render(<ActionShortcutsPanel bindings={[b('f', 'kbd.action.fold')]} includeCardNav />);
    expect(screen.getByText('Enter')).toBeInTheDocument();
    expect(screen.getByText('選択したカードを出す')).toBeInTheDocument();
    expect(screen.getByText('フォールド')).toBeInTheDocument();
  });
});
