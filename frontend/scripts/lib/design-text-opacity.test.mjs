import { describe, expect, it } from 'vitest';
import { findOpacitySuffixedTextTokens } from './design-text-opacity.mjs';

describe('findOpacitySuffixedTextTokens', () => {
  it('finds forbidden tokens in individual class string literals, including conditional branches', () => {
    const successToken = 'text-ds-success' + '/80';
    expect(findOpacitySuffixedTextTokens(`"${successToken}" : "text-ds-text-primary"`)).toEqual([
      { token: successToken, index: 1 },
    ]);
    const mutedToken = 'text-ds-text-muted';
    const accentHoverToken = 'text-ds-accent-hover';
    expect(
      findOpacitySuffixedTextTokens(`${mutedToken}/60 ${accentHoverToken}/50 ${mutedToken}/[.6]`).map(
        ({ token }) => token,
      ),
    ).toEqual([`${mutedToken}/60`, `${accentHoverToken}/50`, `${mutedToken}/[.6]`]);
  });

  it('allows opaque text tokens and opacity utilities unrelated to text colors', () => {
    expect(findOpacitySuffixedTextTokens('"text-ds-success opacity-80 bg-ds-success/20"')).toEqual([]);
    expect(findOpacitySuffixedTextTokens('bg-ds-surface/80 text-ds-text-muted')).toEqual([]);
  });
});
