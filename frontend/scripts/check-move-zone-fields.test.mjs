import { spawnSync } from 'node:child_process';
import { existsSync, mkdirSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { afterAll, describe, expect, it } from 'vitest';

const GUARD = join(process.cwd(), 'scripts', 'check-move-zone-fields.mjs');
if (!existsSync(GUARD)) throw new Error(`check-move-zone-fields.mjs not found at ${GUARD} (cwd: ${process.cwd()})`);

const dirs = [];
afterAll(() => {
  for (const d of dirs) rmSync(d, { recursive: true, force: true });
});

const goZone = (game, fields) =>
  `package controller\n\ntype ${game}WebZone struct {\n${fields
    .map((f) => `\t${f[0].toUpperCase()}${f.slice(1)} *int \`json:"${f},omitempty"\`\n`)
    .join('')}}\n`;

const tsZone = (game, fields) =>
  `export interface ${game}MoveZone {\n${fields.map((f) => `  ${f}?: number;\n`).join('')}}\n`;

function check(goSrc, tsSrc) {
  const dir = mkdtempSync(join(tmpdir(), 'move-zone-'));
  dirs.push(dir);
  mkdirSync(join(dir, 'go'));
  mkdirSync(join(dir, 'ts'));
  writeFileSync(join(dir, 'go', 'DemoWebController.go'), goSrc);
  writeFileSync(join(dir, 'ts', 'demo.ts'), tsSrc);
  const r = spawnSync(process.execPath, [GUARD, join(dir, 'go'), join(dir, 'ts')], { encoding: 'utf8' });
  return { code: r.status, out: `${r.stdout}${r.stderr}` };
}

describe('check-move-zone-fields', () => {
  it('accepts a zone whose fields match on both sides', () => {
    const r = check(goZone('Demo', ['zone', 'col']), tsZone('Demo', ['zone', 'col']));
    expect(r.code).toBe(0);
    expect(r.out).toContain('OK');
  });

  it('accepts the other spelling when both sides use it', () => {
    const r = check(goZone('Demo', ['zone', 'idx']), tsZone('Demo', ['zone', 'idx']));
    expect(r.code).toBe(0);
  });

  // **これが #5288 (Colorado) そのもの。** Go が col、TS が idx で、型検査も
  // 単体テストも通ったまま、行き先つきの移動が全部 400 になった。
  it('rejects the col/idx split that shipped in #5288', () => {
    const r = check(goZone('Demo', ['zone', 'col']), tsZone('Demo', ['zone', 'idx']));
    expect(r.code).toBe(1);
    expect(r.out).toContain('only in Go: col');
    expect(r.out).toContain('only in TS: idx');
  });

  it('rejects a field the client never sends', () => {
    const r = check(goZone('Demo', ['zone', 'col', 'cardIndex']), tsZone('Demo', ['zone', 'col']));
    expect(r.code).toBe(1);
    expect(r.out).toContain('only in Go: cardIndex');
  });

  // **タグの無いフィールドは Go のフィールド名で送られる。** 読み飛ばすと
  // 「TS に無い」ことに気付けないまま通ってしまう。
  it('counts a Go field that carries no json tag', () => {
    const goSrc = 'package controller\n\ntype DemoWebZone struct {\n\tZone string `json:"zone"`\n\tCol *int\n}\n';
    const r = check(goSrc, tsZone('Demo', ['zone', 'col']));
    expect(r.code).toBe(1);
    expect(r.out).toContain('only in Go: Col');
  });

  // extends を挟んだ宣言も拾う。取りこぼすと比較から静かに外れる。
  it('reads an interface declared with extends', () => {
    const ts =
      'interface Base {\n  zone: string;\n}\nexport interface DemoMoveZone extends Base {\n  idx?: number;\n}\n';
    const r = check(goZone('Demo', ['zone', 'col']), ts);
    expect(r.code).toBe(1);
    expect(r.out).toContain('only in TS: idx');
  });

  // サブディレクトリに置かれた構造体も走査する。走査漏れは「一致」に見える。
  it('walks sub-directories of the Go controller tree', () => {
    const dir = mkdtempSync(join(tmpdir(), 'move-zone-'));
    dirs.push(dir);
    mkdirSync(join(dir, 'go', 'nested'), { recursive: true });
    mkdirSync(join(dir, 'ts'));
    writeFileSync(join(dir, 'go', 'nested', 'DemoWebController.go'), goZone('Demo', ['zone', 'col']));
    writeFileSync(join(dir, 'ts', 'demo.ts'), tsZone('Demo', ['zone', 'idx']));
    const r = spawnSync(process.execPath, [GUARD, join(dir, 'go'), join(dir, 'ts')], { encoding: 'utf8' });
    expect(r.status).toBe(1);
    expect(`${r.stdout}${r.stderr}`).toContain('only in Go: col');
  });

  // 片側にしか無いゲームは不一致ではない (クライアントがゾーンを送らない設計)。
  it('ignores a Go zone with no TypeScript counterpart', () => {
    const r = check(goZone('Demo', ['zone', 'col']), 'export interface OtherMoveZone {\n  zone?: number;\n}\n');
    expect(r.code).toBe(0);
    expect(r.out).toContain('0 of 1');
  });
});
