import { useTranslation } from 'react-i18next';
import { OldMaidMode } from '../../hooks/useOldMaidGame';
import { btnPrimary, btnSecondary } from '../../styles/buttonStyles';
import { Modal } from '../common/Modal';

/** Props for the OldMaidSettingsDialog component. */
interface OldMaidSettingsDialogProps {
  open: boolean;
  mode: number;
  cpuPlacementStrategy: boolean;
  cpuMemoryAI: boolean;
  cpuHesitationEnabled: boolean;
  cpuMetaAI: boolean;
  onModeChange: (m: number) => void;
  onStrategyChange: (v: boolean) => void;
  onMemoryAIChange: (v: boolean) => void;
  onHesitationChange: (v: boolean) => void;
  onMetaAIChange: (v: boolean) => void;
  onApply: () => void;
  onClose: () => void;
}

/** Renders a modal dialog for Old Maid game settings (mode and CPU options). */
export function OldMaidSettingsDialog({
  open,
  mode,
  cpuPlacementStrategy,
  cpuMemoryAI,
  cpuHesitationEnabled,
  cpuMetaAI,
  onModeChange,
  onStrategyChange,
  onMemoryAIChange,
  onHesitationChange,
  onMetaAIChange,
  onApply,
  onClose,
}: OldMaidSettingsDialogProps) {
  const { t } = useTranslation('oldmaid');

  return (
    <Modal
      open={open}
      onClose={onClose}
      role="dialog"
      ariaLabelledBy="oldmaid-settings-title"
      panelClassName="glass-panel rounded-lg shadow-xl p-6 max-w-sm mx-4"
    >
      <h2 id="oldmaid-settings-title" className="text-lg font-bold text-ds-text-primary mb-4">
        {t('setup.title')}
      </h2>
      <div className="flex flex-col gap-3">
        <fieldset className="flex flex-col gap-3 border-0 p-0 m-0">
          <legend className="text-ds-text-primary font-bold mb-1">{t('setup.modeSelect')}</legend>
          <label className="flex items-center gap-2 text-ds-text-primary cursor-pointer min-h-[44px]">
            <input
              type="radio"
              name="oldmaid-mode"
              value={OldMaidMode.Normal}
              checked={mode === OldMaidMode.Normal}
              onChange={() => onModeChange(OldMaidMode.Normal)}
            />
            {t('setup.normal')}
          </label>
          <label className="flex items-center gap-2 text-ds-text-primary cursor-pointer min-h-[44px]">
            <input
              type="radio"
              name="oldmaid-mode"
              value={OldMaidMode.JijiNuki}
              checked={mode === OldMaidMode.JijiNuki}
              onChange={() => onModeChange(OldMaidMode.JijiNuki)}
            />
            {t('setup.jijiNuki')}
          </label>
        </fieldset>
        <div className="border-t border-white/20 my-1" />
        <fieldset className="flex flex-col gap-3 border-0 p-0 m-0">
          <legend className="text-ds-text-primary font-bold mb-1">{t('setup.cpuSettings')}</legend>
          <label className="flex items-center gap-2 text-ds-text-primary cursor-pointer min-h-[44px]">
            <input
              type="checkbox"
              checked={cpuPlacementStrategy}
              onChange={(e) => onStrategyChange(e.target.checked)}
            />
            {t('setup.cpuStrategy')}
          </label>
          <label className="flex items-center gap-2 text-ds-text-primary cursor-pointer min-h-[44px]">
            <input type="checkbox" checked={cpuMemoryAI} onChange={(e) => onMemoryAIChange(e.target.checked)} />
            {t('setup.cpuMemoryAI')}
          </label>
          <label className="flex items-center gap-2 text-ds-text-primary cursor-pointer min-h-[44px]">
            <input
              type="checkbox"
              checked={cpuHesitationEnabled}
              onChange={(e) => onHesitationChange(e.target.checked)}
            />
            {t('setup.cpuHesitation')}
          </label>
          <label className="flex items-center gap-2 text-ds-text-primary cursor-pointer min-h-[44px]">
            <input type="checkbox" checked={cpuMetaAI} onChange={(e) => onMetaAIChange(e.target.checked)} />
            {t('setup.cpuMetaAI')}
          </label>
        </fieldset>
      </div>
      <div className="flex justify-end gap-2 mt-4">
        <button type="button" className={btnSecondary} onClick={onClose}>
          {t('setup.cancel')}
        </button>
        <button type="button" className={btnPrimary} onClick={onApply}>
          {t('setup.apply')}
        </button>
      </div>
    </Modal>
  );
}
