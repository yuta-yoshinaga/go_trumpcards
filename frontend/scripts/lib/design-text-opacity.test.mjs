import { describe, expect, it } from 'vitest';
import { findOpacitySuffixedTextTokens } from './design-text-opacity.mjs';

describe('findOpacitySuffixedTextTokens', () => {
  it('finds forbidden tokens in individual class string literals, including conditional branches', () => {
    expect(findOpacitySuffixedTextTokens('"text-ds-success/80" : "text-ds-text-primary"')).toEqual([
      { token: 'text-ds-success/80', index: 1 },
    ]);
  });

  it('allows opaque text tokens and opacity utilities unrelated to text colors', () => {
    expect(findOpacitySuffixedTextTokens('"text-ds-success opacity-80 bg-ds-success/20"')).toEqual([]);
  });
});
