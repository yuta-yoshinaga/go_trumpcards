import { btnDanger, btnSecondary } from '../styles/buttonStyles';

export interface ConfirmDialogProps {
  open: boolean;
  title: string;
  message: string;
  confirmLabel: string;
  cancelLabel: string;
  onConfirm: () => void;
  onCancel: () => void;
}

export function ConfirmDialog(props: ConfirmDialogProps) {
  if (!props.open) return null;

  return (
    // biome-ignore lint/a11y/noStaticElementInteractions: overlay backdrop dismisses dialog on click
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
      onClick={props.onCancel}
      onKeyDown={() => {}}
      role="presentation"
    >
      <div
        role="alertdialog"
        aria-modal="true"
        aria-label={props.title}
        className="bg-white rounded-lg shadow-xl p-6 max-w-sm mx-4"
        onClick={(e) => e.stopPropagation()}
        onKeyDown={() => {}}
      >
        <h2 className="text-lg font-bold text-gray-900 mb-2">{props.title}</h2>
        <p className="text-gray-700 mb-4">{props.message}</p>
        <div className="flex justify-end gap-2">
          <button type="button" className={btnSecondary} onClick={props.onCancel}>
            {props.cancelLabel}
          </button>
          <button type="button" className={btnDanger} onClick={props.onConfirm}>
            {props.confirmLabel}
          </button>
        </div>
      </div>
    </div>
  );
}
