import ts from 'typescript';

const BACKGROUND = /\bbg-ds-(?:warning|accent|info|success)(?:-[\w-]+|\/\d+)?\b/;
const WHITE_TEXT = /(?:^|[\s:])text-white\b/;
const CLASS_BUILDERS = new Set(['cn', 'clsx', 'classnames', 'classNames']);

/** Find colored action backgrounds whose class expression can also enable white text. */
export function findColoredActionTextViolations(source) {
  const file = ts.createSourceFile('source.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX);
  const violations = [];
  const reported = new Set();

  const report = (node, match) => {
    const index = node.getStart(file);
    if (reported.has(index)) return;
    reported.add(index);
    violations.push({ index, match });
  };

  const visit = (node, inheritedColoredFill = false) => {
    if (ts.isJsxElement(node) || ts.isJsxSelfClosingElement(node)) {
      let expression;
      for (const attribute of (ts.isJsxElement(node) ? node.openingElement.attributes : node.attributes).properties) {
        if (ts.isJsxAttribute(attribute) && attribute.name.getText(file) === 'className') {
          expression =
            attribute.initializer && ts.isJsxExpression(attribute.initializer)
              ? attribute.initializer.expression
              : attribute.initializer;
          if (expression) {
            const variants = evaluate(expression);
            const violation = variants.find(({ value }) => BACKGROUND.test(value) && WHITE_TEXT.test(value));
            if (violation) report(expression, violation.value);
            if (inheritedColoredFill) {
              const inheritedViolation = variants.find(({ value }) => WHITE_TEXT.test(value));
              if (inheritedViolation) report(expression, inheritedViolation.value);
            }
            if (!inheritedColoredFill && variants.length > 0 && variants.every(({ value }) => BACKGROUND.test(value))) {
              inheritedColoredFill = true;
            }
          }
        }
      }

      if (ts.isJsxElement(node)) {
        for (const child of node.children) {
          if (ts.isJsxElement(child) || ts.isJsxSelfClosingElement(child)) {
            visit(child, inheritedColoredFill);
          } else visit(child, inheritedColoredFill);
        }
      }
      return;
    }

    if (inheritedColoredFill && ts.isJsxAttribute(node) && node.name.getText(file) === 'className') {
      const initializer = node.initializer;
      const expression = initializer && ts.isJsxExpression(initializer) ? initializer.expression : initializer;
      if (expression) {
        const match = evaluate(expression).find(({ value }) => WHITE_TEXT.test(value));
        if (match) report(expression, match.value);
      }
      return;
    }
    ts.forEachChild(node, (child) => visit(child, inheritedColoredFill));
  };

  visit(file);
  return violations.sort((a, b) => a.index - b.index);

  function evaluate(node) {
    if (!node) return [{ value: '', choices: {} }];
    if (ts.isParenthesizedExpression(node) || ts.isAsExpression(node) || ts.isTypeAssertionExpression(node)) {
      return evaluate(node.expression);
    }
    if (ts.isStringLiteralLike(node) || ts.isNoSubstitutionTemplateLiteral(node))
      return [{ value: node.text, choices: {} }];
    if (ts.isTemplateExpression(node)) {
      let values = [{ value: node.head.text, choices: {} }];
      for (const span of node.templateSpans) {
        values = combine(values, evaluate(span.expression));
        values = combine(values, [{ value: span.literal.text, choices: {} }]);
      }
      return values;
    }
    if (ts.isConditionalExpression(node)) {
      const key = node.condition.getText(file);
      return unique([
        ...evaluate(node.whenTrue).map((item) => ({ ...item, choices: { ...item.choices, [key]: true } })),
        ...evaluate(node.whenFalse).map((item) => ({ ...item, choices: { ...item.choices, [key]: false } })),
      ]);
    }
    if (ts.isBinaryExpression(node)) {
      const op = node.operatorToken.kind;
      if (op === ts.SyntaxKind.AmpersandAmpersandToken) {
        return unique([{ value: '', choices: {} }, ...evaluate(node.right)]);
      }
      if (op === ts.SyntaxKind.BarBarToken || op === ts.SyntaxKind.QuestionQuestionToken) {
        return unique([...evaluate(node.left), ...evaluate(node.right)]);
      }
      if (op === ts.SyntaxKind.PlusToken) return combine(evaluate(node.left), evaluate(node.right));
      return [{ value: '', choices: {} }];
    }
    if (ts.isCallExpression(node)) {
      const name = node.expression.getText(file).split('.').at(-1);
      if (CLASS_BUILDERS.has(name))
        return node.arguments.reduce(
          (values, arg) =>
            combine(
              values,
              evaluate(arg).map((item) => ({ ...item, value: `${item.value} ` })),
            ),
          [{ value: '', choices: {} }],
        );
      return [{ value: '', choices: {} }];
    }
    if (ts.isArrayLiteralExpression(node))
      return node.elements.reduce((values, item) => combine(values, evaluate(item)), [{ value: '', choices: {} }]);
    if (ts.isObjectLiteralExpression(node)) {
      return node.properties.flatMap((property) => {
        if (ts.isPropertyAssignment(property)) {
          const key = property.name.getText(file).replace(/^['"]|['"]$/g, '');
          return [{ value: key, choices: {} }, ...evaluate(property.initializer)];
        }
        return [];
      });
    }
    return [{ value: '', choices: {} }];
  }
}

function combine(left, right) {
  return unique(
    left.flatMap((a) =>
      right.flatMap((b) => {
        if (Object.entries(a.choices).some(([key, value]) => key in b.choices && b.choices[key] !== value)) return [];
        return [{ value: a.value + b.value, choices: { ...a.choices, ...b.choices } }];
      }),
    ),
  );
}

function unique(values) {
  const seen = new Set();
  return values.filter((item) => {
    const key = `${item.value}\0${JSON.stringify(item.choices)}`;
    if (seen.has(key)) return false;
    seen.add(key);
    return true;
  });
}
