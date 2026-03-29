import { describe, expect, it } from 'vitest';
import { getFocusableElements } from './dom';

describe('getFocusableElements', () => {
  function createContainer(...children: HTMLElement[]): HTMLElement {
    const div = document.createElement('div');
    for (const child of children) {
      div.appendChild(child);
    }
    return div;
  }

  function el(tag: string, attrs: Record<string, string> = {}, text?: string): HTMLElement {
    const e = document.createElement(tag);
    for (const [k, v] of Object.entries(attrs)) {
      e.setAttribute(k, v);
    }
    if (text) e.textContent = text;
    return e;
  }

  it('returns focusable elements (a[href], button, input, select, textarea)', () => {
    const container = createContainer(
      el('a', { href: 'https://example.com' }, 'Link'),
      el('button', {}, 'Click'),
      el('input', { type: 'text' }),
      el('select'),
      el('textarea'),
    );
    const result = getFocusableElements(container);
    expect(result).toHaveLength(5);
  });

  it('returns elements with tabindex that is not -1', () => {
    const container = createContainer(
      el('div', { tabindex: '0' }, 'Focusable div'),
      el('span', { tabindex: '1' }, 'Focusable span'),
    );
    const result = getFocusableElements(container);
    expect(result).toHaveLength(2);
  });

  it('excludes elements with tabindex="-1"', () => {
    const container = createContainer(el('div', { tabindex: '-1' }, 'Not focusable'), el('button', {}, 'Focusable'));
    const result = getFocusableElements(container);
    expect(result).toHaveLength(1);
    expect(result[0].textContent).toBe('Focusable');
  });

  it('excludes disabled elements', () => {
    const container = createContainer(
      el('button', { disabled: '' }, 'Disabled'),
      el('button', {}, 'Enabled'),
      el('input', { disabled: '' }),
      el('input'),
    );
    const result = getFocusableElements(container);
    expect(result).toHaveLength(2);
  });

  it('excludes anchor tags without href', () => {
    const container = createContainer(el('a', {}, 'No href'), el('a', { href: '' }, 'Empty href'));
    const result = getFocusableElements(container);
    expect(result).toHaveLength(1);
  });

  it('returns empty array for container with no focusable elements', () => {
    const container = createContainer(el('div', {}, 'Not focusable'), el('p', {}, 'Also not focusable'));
    const result = getFocusableElements(container);
    expect(result).toHaveLength(0);
  });
});
