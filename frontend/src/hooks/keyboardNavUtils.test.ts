import { afterEach, describe, expect, it } from 'vitest';
import { isModalOpen, registerOpenModal } from './keyboardNavUtils';

let unregister: Array<() => void> = [];

afterEach(() => {
  unregister.forEach((fn) => {
    fn();
  });
  unregister = [];
});

describe('modal keyboard registration', () => {
  it('tracks registrations and makes each unregister function idempotent', () => {
    const first = registerOpenModal();
    unregister.push(first);
    expect(isModalOpen()).toBe(true);

    const second = registerOpenModal();
    unregister.push(second);
    first();
    first();
    expect(isModalOpen()).toBe(true);

    second();
    expect(isModalOpen()).toBe(false);
  });
});
