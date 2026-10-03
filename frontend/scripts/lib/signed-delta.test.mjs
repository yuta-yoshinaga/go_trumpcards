import { readFileSync } from 'node:fs';
import { describe, expect, it } from 'vitest';
import { findInlineSignedDelta } from './signed-delta.mjs';

const designTokenGuard = readFileSync('scripts/check-design-tokens.mjs', 'utf8');

describe('findInlineSignedDelta', () => {
  it('floors the number of source files scanned and reports that scan count', () => {
    expect(designTokenGuard).toMatch(/assertFloor\('signed-delta',\s*files\.length,\s*\d+,\s*'source files scanned'\)/);
    expect(designTokenGuard).toMatch(
      /signed-delta: OK \(\$\{files\.length\} source files scanned, \$\{signedDeltaViolations\.length\} inline expressions\)\./,
    );
  });

  it('detects positive and non-negative inline plus formatting', () => {
    const source = ['score > 0 ? `+', '$', "{score}` : score; delta >= 0 ? '+' : ''"].join('');
    expect(findInlineSignedDelta(source)).toHaveLength(2);
  });

  it('does not flag colors, arithmetic, units, or zero placeholders', () => {
    const source = "score > 0 ? 'text-green' : 'text-red'; total + 1; rate > 0 ? '+1.5x' : '-'";
    expect(findInlineSignedDelta(source)).toHaveLength(0);
  });

  it('allows signed values with a special zero label', () => {
    const source = ['net > 0 ? `+', '$', '{net}` : net < 0 ? `', '$', "{net}` : 'no change'"].join('');
    expect(findInlineSignedDelta(source)).toHaveLength(0);
  });

  it('allows decimal formatting that deliberately formats a different value', () => {
    const source = ['n > 0 ? `+', '$', '{formatted}` : formatted'].join('');
    expect(findInlineSignedDelta(source)).toHaveLength(0);
  });
});
