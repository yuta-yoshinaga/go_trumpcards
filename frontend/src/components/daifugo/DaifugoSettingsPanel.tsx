import { useTranslation } from 'react-i18next';
import type { DaifugoConfigInput } from '../../types/card';

interface SettingsPanelProps {
  config: DaifugoConfigInput;
  onChange: (key: keyof DaifugoConfigInput, value: boolean | number) => void;
}

export function DaifugoSettingsPanel({ config, onChange }: SettingsPanelProps) {
  const { t } = useTranslation('daifugo');
  const boolRules: { key: keyof DaifugoConfigInput; label: string }[] = [
    { key: 'eightCutEnabled', label: t('settings.eightCut') },
    { key: 'elevenBackEnabled', label: t('settings.elevenBack') },
    { key: 'sequenceEnabled', label: t('settings.sequence') },
    { key: 'cardExchangeEnabled', label: t('settings.cardExchange') },
    { key: 'fiveSkipEnabled', label: t('settings.fiveSkip') },
    { key: 'sevenPassEnabled', label: t('settings.sevenPass') },
    { key: 'tenDiscardEnabled', label: t('settings.tenDiscard') },
    { key: 'spadeThreeEnabled', label: t('settings.spadeThree') },
    { key: 'capitalFallEnabled', label: t('settings.capitalFall') },
    { key: 'nineReverseEnabled', label: t('settings.nineReverse') },
    { key: 'coupDetatEnabled', label: t('settings.coupDetat') },
    { key: 'numberLockEnabled', label: t('settings.numberLock') },
    { key: 'sandstormEnabled', label: t('settings.sandstorm') },
    { key: 'emperorEnabled', label: t('settings.emperor') },
    { key: 'sequenceRevolutionEnabled', label: t('settings.sequenceRevolution') },
    { key: 'illegalFinishEnabled', label: t('settings.illegalFinish') },
    { key: 'queenBomberEnabled', label: t('settings.queenBomber') },
  ];
  return (
    <details className="mb-2">
      <summary className="cursor-pointer text-[#ccc] text-[0.85em] select-none">{t('settings.title')}</summary>
      <div className="bg-black/40 rounded-lg p-2 mt-1 text-[0.82em] text-white">
        <div className="mb-1 flex flex-wrap gap-x-4 gap-y-1">
          <span>
            <label htmlFor="joker-count" className="mr-2">
              {t('settings.jokerCount')}
            </label>
            <select
              id="joker-count"
              value={config.jokerCount}
              onChange={(e) => onChange('jokerCount', Number(e.target.value))}
              className="bg-black/50 text-white rounded px-1"
            >
              <option value={0}>0</option>
              <option value={1}>1</option>
              <option value={2}>2</option>
            </select>
          </span>
          <span>
            <label htmlFor="cpu-difficulty" className="mr-2">
              {t('settings.cpuDifficulty')}
            </label>
            <select
              id="cpu-difficulty"
              value={config.cpuDifficulty}
              onChange={(e) => onChange('cpuDifficulty', Number(e.target.value))}
              className="bg-black/50 text-white rounded px-1"
            >
              <option value={1}>{t('settings.difficultyEasy')}</option>
              <option value={0}>{t('settings.difficultyNormal')}</option>
              <option value={2}>{t('settings.difficultyHard')}</option>
            </select>
          </span>
          <span>
            <label htmlFor="suit-lock-mode" className="mr-2">
              {t('settings.suitLockMode')}
            </label>
            <select
              id="suit-lock-mode"
              value={config.suitLockMode}
              onChange={(e) => onChange('suitLockMode', Number(e.target.value))}
              className="bg-black/50 text-white rounded px-1"
            >
              <option value={0}>{t('settings.suitLockNone')}</option>
              <option value={1}>{t('settings.suitLockPartial')}</option>
              <option value={2}>{t('settings.suitLockFull')}</option>
            </select>
          </span>
          <span>
            <label htmlFor="five-skip-count" className="mr-2">
              {t('settings.fiveSkipCount')}
            </label>
            <select
              id="five-skip-count"
              value={config.fiveSkipCount}
              onChange={(e) => onChange('fiveSkipCount', Number(e.target.value))}
              className="bg-black/50 text-white rounded px-1"
              disabled={!config.fiveSkipEnabled}
            >
              {[1, 2, 3, 4, 5].map((n) => (
                <option key={n} value={n}>
                  {n}
                </option>
              ))}
            </select>
          </span>
        </div>
        <div className="flex flex-wrap gap-x-4 gap-y-1">
          {boolRules.map(({ key, label }) => (
            <label key={key} className="flex items-center gap-1 cursor-pointer">
              <input
                type="checkbox"
                checked={config[key] as boolean}
                onChange={(e) => onChange(key, e.target.checked)}
              />
              {label}
            </label>
          ))}
        </div>
      </div>
    </details>
  );
}
