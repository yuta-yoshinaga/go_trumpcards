import { afterEach, describe, expect, it } from 'vitest';
import i18n from '../i18n';
import type { Card } from '../types/card';
import { hanafudaCornerLabel, hanafudaLabel } from './hanafudaLabel';

const card: Card = { design: 'SPADE', value: 0, deck: 'hanafuda', month: 1, kind: 'bright', label: 'old server label' };
const originalLanguage = i18n.language;
afterEach(async () => {
  await i18n.changeLanguage(originalLanguage);
});

describe('hanafudaLabel', () => {
  it.each([
    ['ja', card, '松·光'],
    ['en', card, 'Pine·Bright'],
    ['ja', { ...card, month: 4, kind: 'yeol' as const }, '藤·열끗'],
    ['en', { ...card, month: 4, kind: 'yeol' as const }, 'Wisteria·Yeol'],
    ['ja', { ...card, kind: 'chaff' as const }, '松·カス'],
    ['en', { ...card, kind: 'chaff' as const }, 'Pine·Chaff'],
  ])('formats %s labels', async (language, input, expected) => {
    await i18n.changeLanguage(language);
    expect(hanafudaLabel(input as Card)).toBe(expected);
  });
  it.each([
    { ...card, deck: 'wizard' },
    { ...card, month: undefined },
    { ...card, kind: undefined },
  ])('returns undefined for incomplete or non-hanafuda cards', (input) => {
    expect(hanafudaLabel(input)).toBeUndefined();
  });
});

describe('hanafudaCornerLabel', () => {
  it.each([
    ['ja', card, '松·光'],
    ['en', card, 'Jan'],
    ['en', { ...card, month: 9, kind: 'chaff' as const }, 'Sep'],
  ])('formats %s corner labels', async (language, input, expected) => {
    await i18n.changeLanguage(language);
    expect(hanafudaCornerLabel(input as Card)).toBe(expected);
  });

  it('returns undefined for non-hanafuda and incomplete cards', () => {
    expect(hanafudaCornerLabel({ ...card, deck: 'wizard' })).toBeUndefined();
    expect(hanafudaCornerLabel({ ...card, month: undefined })).toBeUndefined();
    expect(hanafudaCornerLabel({ ...card, kind: undefined })).toBeUndefined();
  });
});
