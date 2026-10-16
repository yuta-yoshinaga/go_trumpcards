import ts from 'typescript';

const BACKGROUND = /\bbg-ds-(?:warning|accent|info|success)(?:-[\w-]+|\/\d+)?\b/;
const WHITE_TEXT = /(?:^|[\s:])text-white\b/;
const CLASS_BUILDERS = new Set(['cn', 'clsx', 'classnames', 'classNames']);

/** Find colored action backgrounds whose class expression can also enable white text. */
export function findColoredActionTextViolations(source) {
  const file = ts.createSourceFile('source.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX);
  const violations = [];
  const reported = new Set();
  const classExpressions = new Set();
  const classExpressionRanges = [];

  const report = (node, match) => {
    const index = node.getStart(file);
    if (reported.has(index)) return;
    reported.add(index);
    violations.push({ index, match });
  };

  const visit = (node, inheritedColoredFill = false) => {
    if (ts.isJsxElement(node) || ts.isJsxSelfClosingElement(node)) {
      let expression;
      let hasOwnBackground = false;
      let ownColoredBackground = false;
      for (const attribute of (ts.isJsxElement(node) ? node.openingElement.attributes : node.attributes).properties) {
        if (ts.isJsxAttribute(attribute) && attribute.name.getText(file) === 'className') {
          expression =
            attribute.initializer && ts.isJsxExpression(attribute.initializer)
              ? attribute.initializer.expression
              : attribute.initializer;
          if (expression) {
            classExpressions.add(expression);
            classExpressionRanges.push([expression.getStart(file), expression.end]);
            const markClassLiteral = (child) => {
              if (ts.isStringLiteralLike(child) || ts.isNoSubstitutionTemplateLiteral(child))
                classExpressions.add(child);
              ts.forEachChild(child, markClassLiteral);
            };
            markClassLiteral(expression);
            const variants = evaluate(expression);
            const literalText = [];
            const gatherLiterals = (child) => {
              if (ts.isStringLiteralLike(child) || ts.isNoSubstitutionTemplateLiteral(child))
                literalText.push(child.text);
              ts.forEachChild(child, gatherLiterals);
            };
            gatherLiterals(expression);
            const hasAlternative = (child) =>
              (ts.isBinaryExpression(child) &&
                [ts.SyntaxKind.BarBarToken, ts.SyntaxKind.QuestionQuestionToken].includes(child.operatorToken.kind)) ||
              (ts.isObjectLiteralExpression(child) && child.parent && ts.isCallExpression(child.parent)) ||
              child.getChildren(file).some((nested) => hasAlternative(nested));
            if (hasAlternative(expression)) {
              const literals = literalText.join(' ');
              if (BACKGROUND.test(literals) && WHITE_TEXT.test(literals)) report(expression, literals);
            }
            hasOwnBackground = variants.some(({ value }) => /(?:^|\s)bg-[\w/-]+/.test(value));
            ownColoredBackground = variants.length > 0 && variants.every(({ value }) => BACKGROUND.test(value));
            const violation = variants.find(({ value }) => BACKGROUND.test(value) && WHITE_TEXT.test(value));
            if (violation) report(expression, violation.value);
            if (inheritedColoredFill && !variants.some(({ value }) => /(?:^|\s)bg-[\w/-]+/.test(value))) {
              const inheritedViolation = variants.find(({ value }) => WHITE_TEXT.test(value));
              if (inheritedViolation) report(expression, inheritedViolation.value);
            }
          }
        }
      }
      if (hasOwnBackground) inheritedColoredFill = ownColoredBackground;

      // JSX embedded in a non-class attribute has an unknown rendering context.
      for (const attribute of (ts.isJsxElement(node) ? node.openingElement.attributes : node.attributes).properties) {
        if (!ts.isJsxAttribute(attribute) || attribute.name.getText(file) === 'className') continue;
        ts.forEachChild(attribute, (child) => visit(child, false));
      }
      if (ts.isJsxElement(node)) for (const child of node.children) visit(child, inheritedColoredFill);
      return;
    }
    if ((ts.isStringLiteralLike(node) || ts.isNoSubstitutionTemplateLiteral(node)) && !classExpressions.has(node)) {
      if (BACKGROUND.test(node.text) && WHITE_TEXT.test(node.text)) report(node, node.text);
    }
    if (
      ts.isTemplateExpression(node) &&
      !classExpressionRanges.some(([start, end]) => start <= node.getStart(file) && node.end <= end)
    ) {
      const segments = [node.head, ...node.templateSpans.map((span) => span.literal)];
      for (const segment of segments) {
        if (BACKGROUND.test(segment.text) && WHITE_TEXT.test(segment.text)) report(segment, segment.text);
      }
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
      return node.properties.reduce(
        (values, property) => {
          if (ts.isPropertyAssignment(property)) {
            const key = property.name.getText(file).replace(/^['"]|['"]$/g, '');
            return combine(
              values,
              combine(
                [
                  { value: '', choices: {} },
                  { value: key, choices: {} },
                ],
                evaluate(property.initializer),
              ),
            );
          }
          return values;
        },
        [{ value: '', choices: {} }],
      );
    }
    return [{ value: '', choices: {} }];
  }
}

function combine(left, right) {
  if (left.length * right.length > 256) {
    // Beyond this cap, conservatively flag if colored fill and white text occur in any branch.
    const all = [...left, ...right].map(({ value }) => value).join(' ');
    return [
      {
        value: `${BACKGROUND.test(all) ? 'bg-ds-accent ' : ''}${WHITE_TEXT.test(all) ? 'text-white' : ''}`,
        choices: {},
      },
    ];
  }
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
