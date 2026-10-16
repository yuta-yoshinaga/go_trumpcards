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

  // The same condition enables both classes, so the selected branch is a real violation.
  it('flags colored background and white text enabled by the same ternary condition', () => {
    expect(
      findColoredActionTextViolations(
        '<button className={`' +
          '$' +
          "{selected ? 'bg-ds-accent' : 'bg-ds-surface-elevated'} " +
          '$' +
          "{selected ? 'text-white' : 'text-ds-text-primary'}" +
          '`} />',
      ),
    ).toHaveLength(1);
  });

  it('flags colored backgrounds with white text in template static text', () => {
    expect(
      findColoredActionTextViolations('<button className={`bg-ds-warning text-white ' + '$' + '{extra}`} />'),
    ).toHaveLength(1);
  });

  it('flags accent backgrounds with hover white text in template static text', () => {
    expect(
      findColoredActionTextViolations('<button className={`bg-ds-accent hover:text-white ' + '$' + '{extra}`} />'),
    ).toHaveLength(1);
  });

  it('reports a violation only once when it appears in interpolation and static text', () => {
    expect(
      findColoredActionTextViolations(
        '<button className={`bg-ds-warning text-white ' + '$' + '{"bg-ds-warning text-white"}`} />',
      ),
    ).toHaveLength(1);
  });

  it('finds mismatched quote boundaries in TSX source', () => {
    const source = `<button title={t('bid.label')}\n className={'bg-ds-accent text-white ring-2 ring-ds-accent'} />`;
    expect(findColoredActionTextViolations(source)).toHaveLength(1);
  });

  it('combines static white text with colored backgrounds from template branches', () => {
    const source = `<button className={\`text-white \${cond ? 'bg-ds-warning' : 'bg-ds-info'}\`} />`;
    expect(findColoredActionTextViolations(source)).toHaveLength(1);
  });

  it('finds white text on a child span when an ancestor has an unconditional colored fill', () => {
    const source = `<button className="bg-ds-accent text-ds-text-on-accent"><span className="text-[10px] text-white">120</span></button>`;
    expect(findColoredActionTextViolations(source)).toHaveLength(1);
  });

  it('keeps background and white text in mutually exclusive ternary branches separate', () => {
    const source = `<button className={cond ? 'bg-ds-accent text-ds-text-on-accent' : 'bg-ds-surface text-white'} />`;
    expect(findColoredActionTextViolations(source)).toEqual([]);
  });

  it('keeps separate ternaries with the same condition correlated', () => {
    const source = `<button className={\`\${sel ? 'bg-ds-accent' : 'bg-ds-surface'} \${sel ? 'text-ds-text-on-accent' : 'text-white'}\`} />`;
    expect(findColoredActionTextViolations(source)).toEqual([]);
  });

  it('ignores unrelated error backgrounds and supports class builder calls', () => {
    expect(findColoredActionTextViolations(`<button className="bg-ds-error text-white" />`)).toEqual([]);
    expect(
      findColoredActionTextViolations(`<button className={clsx('bg-ds-success', cond && 'text-white')} />`),
    ).toHaveLength(1);
  });

  it('checks standalone string literals outside className', () => {
    expect(findColoredActionTextViolations(`const STYLES = { primary: 'bg-ds-accent text-white' };`)).toHaveLength(1);
    expect(findColoredActionTextViolations('const styles = `bg-ds-accent text-white ' + '$' + '{tone}`;')).toHaveLength(
      1,
    );
  });

  it('combines optional clsx object keys and recognizes hover white text', () => {
    expect(
      findColoredActionTextViolations(`<div className={clsx({ 'bg-ds-accent': a, 'text-white': b })} />`),
    ).toHaveLength(1);
    expect(findColoredActionTextViolations(`<div className="bg-ds-accent hover:text-white" />`)).toHaveLength(1);
  });

  it('combines || and ?? alternatives conservatively', () => {
    expect(
      findColoredActionTextViolations(`<div className={(x || 'bg-ds-accent') + (y ?? 'text-white')} />`),
    ).toHaveLength(1);
  });

  it('does not inherit conditional fills to children or through a child background', () => {
    expect(
      findColoredActionTextViolations(
        `<div className={cond ? 'bg-ds-accent' : 'bg-ds-surface'}><span className="text-white" /></div>`,
      ),
    ).toEqual([]);
    expect(
      findColoredActionTextViolations(
        `<div className="bg-ds-accent"><span className="bg-ds-surface text-white" /></div>`,
      ),
    ).toEqual([]);
  });

  it('checks JSX nested in another attribute without inheriting the parent fill', () => {
    expect(
      findColoredActionTextViolations(`<Button icon={<span className="bg-ds-accent text-white" />} />`),
    ).toHaveLength(1);
  });

  // sel and !sel are modeled as independent conditions, so this conservative analysis can report a false positive.
});
