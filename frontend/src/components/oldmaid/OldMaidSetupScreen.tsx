import { useTranslation } from 'react-i18next';
import { OldMaidMode } from '../../hooks/useOldMaidGame';
import { btnPrimary } from '../../styles/buttonStyles';

interface SetupScreenProps {
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
  onStart: () => void;
}

export function OldMaidSetupScreen({
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
  onStart,
}: SetupScreenProps) {
  const { t } = useTranslation('oldmaid');
  return (
    <div className="flex-1 flex flex-col items-center justify-center bg-game-bg-green p-6 gap-4" aria-busy="false">
      <div className="text-white text-2xl font-bold mb-2">{t('setup.title')}</div>
      <div className="bg-black/40 rounded-xl p-4 w-full max-w-sm flex flex-col gap-3">
        <fieldset className="flex flex-col gap-3 border-0 p-0 m-0">
          <legend className="text-white font-bold mb-1">{t('setup.modeSelect')}</legend>
          <label className="flex items-center gap-2 text-white cursor-pointer">
            <input
              type="radio"
              name="oldmaid-mode"
              value={OldMaidMode.Normal}
              checked={mode === OldMaidMode.Normal}
              onChange={() => onModeChange(OldMaidMode.Normal)}
            />
            {t('setup.normal')}
          </label>
          <label className="flex items-center gap-2 text-white cursor-pointer">
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
          <legend className="text-white font-bold mb-1">{t('setup.cpuSettings')}</legend>
          <label className="flex items-center gap-2 text-white cursor-pointer">
            <input
              type="checkbox"
              checked={cpuPlacementStrategy}
              onChange={(e) => onStrategyChange(e.target.checked)}
            />
            {t('setup.cpuStrategy')}
          </label>
          <label className="flex items-center gap-2 text-white cursor-pointer">
            <input type="checkbox" checked={cpuMemoryAI} onChange={(e) => onMemoryAIChange(e.target.checked)} />
            {t('setup.cpuMemoryAI')}
          </label>
          <label className="flex items-center gap-2 text-white cursor-pointer">
            <input
              type="checkbox"
              checked={cpuHesitationEnabled}
              onChange={(e) => onHesitationChange(e.target.checked)}
            />
            {t('setup.cpuHesitation')}
          </label>
          <label className="flex items-center gap-2 text-white cursor-pointer">
            <input type="checkbox" checked={cpuMetaAI} onChange={(e) => onMetaAIChange(e.target.checked)} />
            {t('setup.cpuMetaAI')}
          </label>
        </fieldset>
      </div>
      <button type="button" className={`${btnPrimary} min-w-[120px] mt-2`} onClick={onStart}>
        {t('setup.start')}
      </button>
    </div>
  );
}
