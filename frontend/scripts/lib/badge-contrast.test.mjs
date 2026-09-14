import { describe, expect, it } from 'vitest';
import { findBadgeContrastViolations } from './badge-contrast.mjs';

const violations = (source) => findBadgeContrastViolations(source).length;

describe('badge contrast element boundaries', () => {
  it.each([
    [
      'a matching foreground in a descendant',
      '<div className="bg-ds-warning/20"><span className="text-ds-warning" /></div>',
      1,
    ],
    ['a matching foreground in the same className', '<div className="bg-ds-warning/20 text-ds-warning" />', 1],
    [
      'a matching foreground in a sibling',
      '<div className="bg-ds-warning/20" /><span className="text-ds-warning" />',
      0,
    ],
    [
      'a different foreground in a descendant',
      '<div className="bg-ds-warning/20"><span className="text-ds-error" /></div>',
      0,
    ],
  ])('%s: returns %i violation(s)', (_name, source, expected) => {
    expect(violations(source)).toBe(expected);
  });
});
