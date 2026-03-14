export interface SettingsItem {
  type: 'checkbox' | 'select';
  id: string;
  label: string;
  tooltip?: string;
  // checkbox
  checked?: boolean;
  onToggle?: (checked: boolean) => void;
  // select
  value?: string | number;
  options?: { value: string | number; label: string }[];
  onSelect?: (value: string) => void;
  disabled?: boolean;
}

export interface SettingsGroup {
  id?: string;
  title?: string;
  items: SettingsItem[];
}

interface SettingsPanelProps {
  title: string;
  groups: SettingsGroup[];
}

export function SettingsPanel({ title, groups }: SettingsPanelProps) {
  return (
    <details className="px-4 pt-2">
      <summary className="text-white/70 text-xs cursor-pointer select-none">{title}</summary>
      <div className="bg-black/30 rounded-lg p-3 mt-1 text-xs text-white">
        {groups.map((group, gi) => {
          const content = (
            <div className="flex flex-wrap gap-x-4 gap-y-1">
              {group.items.map((item) =>
                item.type === 'checkbox' ? (
                  <label
                    key={item.id}
                    htmlFor={item.id}
                    className="flex items-center gap-1 cursor-pointer relative group/tip"
                  >
                    <input
                      id={item.id}
                      type="checkbox"
                      checked={item.checked ?? false}
                      onChange={(e) => item.onToggle?.(e.target.checked)}
                    />
                    {item.label}
                    {item.tooltip && (
                      <span className="hidden group-hover/tip:block group-focus-within/tip:block absolute bottom-full left-0 mb-1 bg-gray-900 text-white text-xs rounded px-2 py-1 whitespace-nowrap z-10">
                        {item.tooltip}
                      </span>
                    )}
                  </label>
                ) : (
                  <span key={item.id} className="flex items-center gap-1 relative group/tip">
                    <label htmlFor={item.id}>{item.label}</label>
                    <select
                      id={item.id}
                      value={item.value}
                      onChange={(e) => item.onSelect?.(e.target.value)}
                      className="bg-black/50 text-white rounded px-1 py-0.5"
                      disabled={item.disabled}
                    >
                      {item.options?.map((opt) => (
                        <option key={opt.value} value={opt.value}>
                          {opt.label}
                        </option>
                      ))}
                    </select>
                    {item.tooltip && (
                      <span className="hidden group-hover/tip:block group-focus-within/tip:block absolute bottom-full left-0 mb-1 bg-gray-900 text-white text-xs rounded px-2 py-1 whitespace-nowrap z-10">
                        {item.tooltip}
                      </span>
                    )}
                  </span>
                ),
              )}
            </div>
          );

          if (group.title) {
            return (
              <fieldset key={group.id ?? group.title} className={gi > 0 ? 'mt-2' : ''}>
                <legend className="text-yellow-300 font-bold text-xs mb-1">{group.title}</legend>
                {content}
              </fieldset>
            );
          }
          return (
            <div key={group.id ?? `group-${gi.toString()}`} className={gi > 0 ? 'mt-2' : ''}>
              {content}
            </div>
          );
        })}
      </div>
    </details>
  );
}
