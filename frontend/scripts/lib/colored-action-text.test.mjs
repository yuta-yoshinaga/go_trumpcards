import { describe, expect, it } from 'vitest';
import { findColoredActionTextViolations } from './colored-action-text.mjs';

describe('findColoredActionTextViolations', () => {
  it('flags warning/accent/info/success backgrounds with white text', () => {
    expect(
      findColoredActionTextViolations(`<button className="bg-ds-warning hover:bg-ds-warning-hover text-white" />`),
    ).toHaveLength(1);
  });

  it('allows dark foregrounds and white text on unrelated backgrounds', () => {
    expect(
      findColoredActionTextViolations(
        `<button className="bg-ds-info text-ds-text-on-accent" /><span className="bg-ds-error text-white" />`,
      ),
    ).toEqual([]);
  });

  it('does not combine colored backgrounds and white text across ternary branches', () => {
    expect(
      findColoredActionTextViolations(
        '<button className={`' +
          '$' +
          "{selected ? 'bg-ds-accent' : 'bg-ds-surface-elevated'} " +
          '$' +
          "{selected ? 'text-white' : 'text-ds-text-primary'}" +
          '`} />',
      ),
    ).toEqual([]);
  });
});
