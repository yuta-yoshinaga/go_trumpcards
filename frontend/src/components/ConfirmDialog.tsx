import { useId } from 'react';
import { btnDanger, btnSecondary } from '../styles/buttonStyles';
import { Modal } from './common/Modal';

/** Props for the ConfirmDialog component. */
export interface ConfirmDialogProps {
  open: boolean;
  title: string;
  message: string;
  confirmLabel: string;
  cancelLabel: string;
  onConfirm: () => void;
  onCancel: () => void;
}

/** Renders a modal confirmation dialog with confirm and cancel buttons. */
export function ConfirmDialog(props: ConfirmDialogProps) {
  const id = useId();
  const titleId = `${id}-title`;
  const descId = `${id}-desc`;
  const hasDescription = props.message.length > 0;

  return (
    <Modal
      open={props.open}
      onClose={props.onCancel}
      role="alertdialog"
      ariaLabelledBy={titleId}
      ariaDescribedBy={hasDescription ? descId : undefined}
      panelClassName="glass-panel rounded-lg shadow-xl p-6 max-w-sm mx-4"
    >
      <h2 id={titleId} className="text-lg font-bold text-ds-text-primary mb-2">
        {props.title}
      </h2>
      {hasDescription && (
        <p id={descId} className="text-ds-text-primary mb-4">
          {props.message}
        </p>
      )}
      <div className="flex justify-end gap-2">
        <button type="button" className={btnSecondary} onClick={props.onCancel}>
          {props.cancelLabel}
        </button>
        <button type="button" className={btnDanger} onClick={props.onConfirm}>
          {props.confirmLabel}
        </button>
      </div>
    </Modal>
  );
}
