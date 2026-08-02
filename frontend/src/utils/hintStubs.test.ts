import { describe, expect, it } from 'vitest';
import { splitRegistrations, stubbedFactoryNames } from './hintStubs';

const stubSrc = `
export function getFooHint(_state: FooResponse): HintResult | null {
  return null;
}
`;

const realSrc = `
export function getBarHint(state: BarResponse): HintResult | null {
  if (!state.hint) return null;
  return { targetAction: 'play', reason: 'x', confidence: 'moderate' };
}
`;

// **接頭辞 get を仮定しない。**chinesepokerHint には付いていない。
const unprefixedStubSrc = `
export function chinesepokerHint(_state: ChinesePokerResponse): HintResult | null {
  return null;
}
`;

describe('stubbedFactoryNames', () => {
  it('finds a factory whose whole body is a null return', () => {
    expect(stubbedFactoryNames({ 'fooHint.ts': stubSrc })).toEqual(new Set(['getFooHint']));
  });

  it('does not flag a factory that can return a hint', () => {
    expect(stubbedFactoryNames({ 'barHint.ts': realSrc }).size).toBe(0);
  });

  it('finds a stub whose name lacks the get prefix', () => {
    expect(stubbedFactoryNames({ 'chinesepokerHint.ts': unprefixedStubSrc })).toEqual(new Set(['chinesepokerHint']));
  });

  it('ignores test files', () => {
    expect(stubbedFactoryNames({ 'fooHint.test.ts': stubSrc }).size).toBe(0);
  });

  // レビュー指摘 (#4602): 本体に何か足された stub は「実装済み」として
  // 黙って通る。それは検出漏れであって誤検出ではないので、**通す側**が正しい。
  it('treats a stub with any other statement as implemented', () => {
    const withLog = stubSrc.replace('  return null;', '  console.log("x");\n  return null;');
    expect(stubbedFactoryNames({ 'fooHint.ts': withLog }).size).toBe(0);
  });
});

describe('splitRegistrations', () => {
  const stubs = new Set(['getFooHint']);

  it('separates inline nulls, stub delegates and real factories', () => {
    const body = `export const hintFactories = {
  alpha: (s) => getBarHint(s as BarResponse),
  beta: () => null,
  gamma: (s) => getFooHint(s as FooResponse),
  delta: (_s) => null,
`;
    const { hinted, stubbed } = splitRegistrations(body, stubs);
    expect(hinted).toEqual(new Set(['alpha']));
    expect(stubbed).toEqual(new Set(['beta', 'gamma', 'delta']));
  });

  // **キャスト変数名を s に決め打ちしない。**レビューで指摘された脆さ。
  it('recognises a delegate whose cast variable is not named s', () => {
    const body = 'export const hintFactories = {\n  gamma: (state) => getFooHint(state as FooResponse),\n';
    expect(splitRegistrations(body, stubs).stubbed).toEqual(new Set(['gamma']));
  });
});
