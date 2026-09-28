/** Tags that should not trigger keyboard navigation shortcuts. */
export const IGNORED_TAGS = new Set(['INPUT', 'TEXTAREA', 'SELECT']);

let openModalCount = 0;

/** Registers an open modal and returns an idempotent function to unregister it. */
export function registerOpenModal(): () => void {
  openModalCount += 1;
  let registered = true;
  return () => {
    if (!registered) return;
    registered = false;
    openModalCount = Math.max(0, openModalCount - 1);
  };
}

/** Returns whether at least one modal is currently registered as open. */
export function isModalOpen(): boolean {
  return openModalCount > 0;
}
