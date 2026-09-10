import { describe, expect, it } from 'vitest';
import { getBassetHint } from './bassetHint';

describe('bassetHint', () => {
  it('provides a hint', () => expect(getBassetHint({ phase: 3, gameEndFlag: false } as never)).toBeTruthy());
});
