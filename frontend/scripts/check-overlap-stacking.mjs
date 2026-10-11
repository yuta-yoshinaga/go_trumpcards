#!/usr/bin/env bun
import { readdirSync, readFileSync } from 'node:fs';
import { join, relative } from 'node:path';
import ts from 'typescript';
import { assertFloor } from './lib/floor.mjs';

const ROOT = new URL('../src/', import.meta.url).pathname;
const POSITIONED = /(?:^|\s)(?:relative|absolute|fixed|sticky)(?:\s|$)/;
const NEGATIVE_OVERLAP = /(?:^|\s)-m[lrse]-\d+(?:\.\d+)?(?:\/\S+)?(?:\s|$)/;
const NEGATIVE_SPACE = /(?:^|\s)-space-x-\d+(?:\.\d+)?(?:\/\S+)?(?:\s|$)/;

function tsxFiles(dir) {
  return readdirSync(dir, { withFileTypes: true }).flatMap((entry) => {
    const path = join(dir, entry.name);
    if (entry.isDirectory()) return tsxFiles(path);
    return entry.isFile() && path.endsWith('.tsx') && !/\.test\.tsx$/.test(path) ? [path] : [];
  });
}

function attr(node, name) {
  return node.attributes.properties.find((property) => ts.isJsxAttribute(property) && property.name.getText() === name);
}

function classValue(node) {
  const value = attr(node, 'className')?.initializer;
  return value ? value.getText().replace(/[{}'"`]/g, ' ') : '';
}

function hasNegativeMargin(node) {
  const style = attr(node, 'style')?.initializer?.getText() ?? '';
  return (
    /marginLeft\s*:\s*(?:-\s*(?:\d|Math\.)|[^,}]*\?\s*-\s*(?:\d|Math\.))/.test(style) ||
    NEGATIVE_OVERLAP.test(classValue(node))
  );
}

export function findOverlapStackingViolations(source, file = 'source.tsx') {
  const sf = ts.createSourceFile(file, source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX);
  const violations = [];
  function walk(node, inheritedSpace = false) {
    if (ts.isJsxElement(node) || ts.isJsxSelfClosingElement(node)) {
      const opening = ts.isJsxElement(node) ? node.openingElement : node;
      const classes = classValue(opening);
      const ownOverlap = hasNegativeMargin(opening);
      const spaceOverlap = inheritedSpace || NEGATIVE_SPACE.test(classes);
      if ((ownOverlap || spaceOverlap) && /opacity-/.test(classes) && !POSITIONED.test(classes)) {
        const { line } = sf.getLineAndCharacterOfPosition(opening.getStart(sf));
        violations.push({
          file,
          line: line + 1,
          reason: ownOverlap ? 'negative horizontal overlap' : 'inside -space-x-* parent',
        });
      }
      const nextSpace = spaceOverlap;
      if (ts.isJsxElement(node)) for (const child of node.children) walk(child, nextSpace);
      return;
    }
    ts.forEachChild(node, (child) => walk(child, inheritedSpace));
  }
  walk(sf);
  return violations;
}

export function scanOverlapStacking(root = ROOT) {
  const files = tsxFiles(root);
  const violations = files.flatMap((file) =>
    findOverlapStackingViolations(readFileSync(file, 'utf8'), relative(root, file)),
  );
  return { inspected: files.length, violations };
}

if (import.meta.main) {
  const { inspected, violations } = scanOverlapStacking();
  assertFloor('overlap-stacking', inspected, 100, 'TSX source files');
  for (const item of violations)
    console.error(`${item.file}:${item.line}: opacity element in ${item.reason} needs positioning`);
  if (violations.length) process.exit(1);
  console.log(`checked overlap stacking across ${inspected} TSX files`);
}
