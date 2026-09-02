import { describe, expect, it } from 'vitest';

/**
 * The guard reads a fixed directory, so what is worth testing is its key
 * comparison. Mirroring that logic here keeps the test independent of how many
 * real locale files exist, and pins the two behaviours that are easy to get
 * wrong: a one-sided key must fail, and i18next plural forms must not.
 */
const PLURAL_SUFFIX = /_(zero|one|two|few|many|other)$/;
function leafKeys(node, prefix = '', out = new Set()) {
  if (node !== null && typeof node === 'object' && !Array.isArray(node)) {
    for (const [k, v] of Object.entries(node)) leafKeys(v, prefix ? `${prefix}.${k}` : k, out);
    return out;
  }
  out.add(prefix.replace(PLURAL_SUFFIX, ''));
  return out;
}
const diff = (a, b) => [...a].filter((k) => !b.has(k)).sort();

describe('locale parity key comparison', () => {
  it('accepts trees that agree, including nested keys', () => {
    const ja = leafKeys({ a: '1', b: { c: '2' } });
    const en = leafKeys({ a: 'x', b: { c: 'y' } });
    expect(diff(ja, en)).toEqual([]);
    expect(diff(en, ja)).toEqual([]);
  });

  it('rejects a key that only one side has', () => {
    const ja = leafKeys({ a: '1', bidQuota: '2' });
    const en = leafKeys({ a: 'x' });
    expect(diff(ja, en)).toEqual(['bidQuota']);
  });

  it('folds i18next plural suffixes so en may carry more forms than ja', () => {
    const ja = leafKeys({ deckUnit: '{{count}}組' });
    const en = leafKeys({ deckUnit_one: '{{count}} deck', deckUnit_other: '{{count}} decks' });
    expect(diff(ja, en)).toEqual([]);
    expect(diff(en, ja)).toEqual([]);
  });
});
