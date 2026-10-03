import { describe, expect, it } from 'vitest';
import { findErrorTextTokenViolations } from './error-text-token.mjs';

describe('findErrorTextTokenViolations', () => {
  it('flags legacy error text on a dark surface, including variant utilities', () => {
    expect(findErrorTextTokenViolations('<span className="text-ds-error hover:text-ds-error" />')).toHaveLength(1);
  });

  it('allows error text on a white background and leaves non-text utilities alone', () => {
    expect(
      findErrorTextTokenViolations(
        '<button className="bg-white text-ds-error" /><div className="bg-ds-error border-ds-error ring-ds-error" />',
      ),
    ).toEqual([]);
  });
});
