import { useCallback, useState } from 'react';
import { btnSecondary } from '../styles/buttonStyles';
import { ManualModal } from './ManualModal';

/** Props for the ManualButton component. */
export interface ManualButtonProps {
  gamePath: string;
}

/** Renders a button that opens the game manual modal for the given game path. */
export function ManualButton({ gamePath }: ManualButtonProps) {
  const [open, setOpen] = useState(false);
  const handleOpen = useCallback(() => setOpen(true), []);
  const handleClose = useCallback(() => setOpen(false), []);

  return (
    <>
      <button
        type="button"
        className={`${btnSecondary} text-xs`}
        onClick={handleOpen}
        aria-label="Manual"
        title="Manual"
      >
        {'\u{1F4D6}'}
      </button>
      <ManualModal open={open} onClose={handleClose} gamePath={gamePath} />
    </>
  );
}
