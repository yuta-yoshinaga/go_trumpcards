import { describe, expect, it } from 'vitest';
import { countVacuousAssertions } from './check-vacuous-text-content.mjs';

describe('check-vacuous-text-content', () => {
  it('counts the empty positive assertion', () => {
    expect(countVacuousAssertions("expect(node).toHaveTextContent('')")).toBe(1);
    expect(countVacuousAssertions('expect(node).toHaveTextContent("")')).toBe(1);
  });
  it('ignores meaningful and inverse assertions', () => {
    const source =
      "toBeEmptyDOMElement()\ntoHaveTextContent('x')\ntoHaveTextContent(/^$/)\nexpect(node).not.toHaveTextContent('')";
    expect(countVacuousAssertions(source)).toBe(0);
  });
});
