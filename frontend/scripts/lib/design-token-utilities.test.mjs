import { describe, expect, test } from 'vitest';
import { collectDesignTokens, findUndefinedDesignUtilities } from './design-token-utilities.mjs';

describe('design-token utility guard', () => {
  test('accepts color-backed and named shadow utilities, including variants and opacity', () => {
    const tokens = collectDesignTokens(`
      @theme {
        --color-ds-error: #b83a3a;
        --color-ds-warning: #e8923a;
        --shadow-ds-accent-glow: 0 0 10px #d4a853;
      }
    `);

    expect(
      findUndefinedDesignUtilities(
        'shadow-lg shadow-ds-error/50 hover:shadow-ds-warning/50 shadow-ds-accent-glow',
        tokens,
      ),
    ).toEqual([]);
  });

  test('rejects an undefined token without treating a shared prefix as a match', () => {
    const tokens = collectDesignTokens('@theme { --color-ds-text-primary: #e8e0d4; }');

    expect(findUndefinedDesignUtilities('text-ds-text text-ds-text-primary', tokens)).toEqual([
      { utility: 'text-ds-text', index: 0 },
    ]);
  });
});
