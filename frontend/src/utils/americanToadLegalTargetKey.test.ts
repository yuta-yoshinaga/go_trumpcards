import { describe, expect, it } from 'vitest';
import { americanToadLegalTargetKey } from './americanToadLegalTargetKey';

describe('americanToadLegalTargetKey', () => {
  it.each([
    [{ zone: 'tableau', col: 2, cardIndex: 1 }, 'tableau-2-1'],
    [{ zone: 'reserve' }, 'reserve'],
    [{ zone: 'waste' }, 'waste'],
  ] as const)('builds key for %s', (source, key) => {
    expect(americanToadLegalTargetKey(source)).toBe(key);
  });

  it.each([null, undefined, { zone: 'tableau', col: 1 }, { zone: 'tableau', cardIndex: 2 }, { zone: 'stock' }])(
    'returns no key when source lacks a legal-target key: %s',
    (source) => expect(americanToadLegalTargetKey(source)).toBeUndefined(),
  );
});
