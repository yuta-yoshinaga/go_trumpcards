import { describe, expect, it, vi } from 'vitest';
import { findLiteralAccessibleNames } from './check-literal-accessible-names.mjs';

describe('check-literal-accessible-names', () => {
  it('does not scan or print when imported', async () => {
    const log = vi.spyOn(console, 'log').mockImplementation(() => {});
    const error = vi.spyOn(console, 'error').mockImplementation(() => {});
    try {
      await import('./check-literal-accessible-names.mjs?import-test');
      expect(log).not.toHaveBeenCalled();
      expect(error).not.toHaveBeenCalled();
    } finally {
      log.mockRestore();
      error.mockRestore();
    }
  });

  it('finds English string literals in aria-label, title, and alt attributes', () => {
    expect(
      findLiteralAccessibleNames('<div aria-label="Loading" /><img alt="Card back" /><button title="Close" />'),
    ).toEqual(['aria-label="Loading"', 'alt="Card back"', 'title="Close"']);
  });

  it('ignores non-English literals and translated expressions', () => {
    expect(findLiteralAccessibleNames('<div aria-label="ゲームナビゲーション" /><img alt={t("card.back")} />')).toEqual(
      [],
    );
  });
});
