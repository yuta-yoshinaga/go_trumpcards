import ts from 'typescript';
import { describe, expect, it } from 'vitest';

const sources = import.meta.glob<string>('./*Commands.ts', { query: '?raw', import: 'default', eager: true });
const UNRESOLVED_ALLOWLIST = ['sharedBettingCommands.ts', 'sharedTrickCommands.ts'];

function unwrap(expression: ts.Expression): ts.Expression {
  while (
    ts.isAsExpression(expression) ||
    ts.isSatisfiesExpression(expression) ||
    ts.isTypeAssertionExpression(expression)
  ) {
    expression = expression.expression;
  }
  return expression;
}

function resolveArray(source: ts.SourceFile, name: string): Set<string> | null {
  for (const statement of source.statements) {
    if (!ts.isVariableStatement(statement)) continue;
    for (const declaration of statement.declarationList.declarations) {
      if (!ts.isIdentifier(declaration.name) || declaration.name.text !== name || !declaration.initializer) continue;
      const value = unwrap(declaration.initializer);
      if (!ts.isArrayLiteralExpression(value)) return null;
      return new Set(value.elements.filter(ts.isStringLiteralLike).map((element) => element.text));
    }
  }
  return null;
}

describe('valid command arrays', () => {
  it('contain every command accepted by guarded switches', () => {
    const files = Object.keys(sources).filter((file) => !file.endsWith('.test.ts'));
    const uncovered: string[] = [];
    const unresolved: string[] = [];
    let inspected = 0;

    for (const file of files) {
      const source = ts.createSourceFile(file, sources[file]!, ts.ScriptTarget.Latest, true);
      const declarations = new Map<string, Set<string> | null>();
      const visit = (node: ts.Node): void => {
        if (ts.isSwitchStatement(node) && ts.isIdentifier(node.expression) && node.expression.text === 'cmd') {
          const fallback = node.caseBlock.clauses.find(ts.isDefaultClause);
          let commandArray: string | undefined;
          if (fallback) {
            const scanDefault = (item: ts.Node): void => {
              if (
                ts.isCallExpression(item) &&
                ts.isIdentifier(item.expression) &&
                (item.expression.text === 'unknownCommand' || item.expression.text === 'suggestCommand') &&
                item.arguments.length >= 2 &&
                ts.isIdentifier(item.arguments[0]) &&
                item.arguments[0].text === 'cmd' &&
                ts.isIdentifier(item.arguments[1])
              )
                commandArray = item.arguments[1].text;
              ts.forEachChild(item, scanDefault);
            };
            fallback.statements.forEach(scanDefault);
          }
          if (commandArray) {
            if (!declarations.has(commandArray)) declarations.set(commandArray, resolveArray(source, commandArray));
            const valid = declarations.get(commandArray);
            if (!valid) unresolved.push(file.slice(2));
            else {
              inspected++;
              const accepted = new Set<string>();
              for (const clause of node.caseBlock.clauses) {
                if (!ts.isCaseClause(clause) || !ts.isStringLiteralLike(clause.expression)) continue;
                accepted.add(clause.expression.text);
              }
              for (const command of accepted)
                if (!valid.has(command) && command !== 'help' && command !== '?')
                  uncovered.push(`${file.slice(2)}: ${command}`);
            }
          }
        }
        ts.forEachChild(node, visit);
      };
      visit(source);
    }

    const unresolvedUnique = [...new Set(unresolved)].sort();
    expect(unresolvedUnique).toEqual([...UNRESOLVED_ALLOWLIST].sort());
    expect(inspected).toBeGreaterThanOrEqual(270);
    expect(uncovered).toEqual([]);
  });
});
