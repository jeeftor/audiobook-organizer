import { createContext, useContext, useState, useEffect, useCallback, ReactNode } from 'react'
import { main } from '../../wailsjs/go/models'
import {
  GetCurrentLayout,
  GetCurrentAuthorFormat,
  GetRenameConfig,
  GetFieldMappingOptions,
  GetAvailableLayouts,
} from '../../wailsjs/go/main/App'

interface AppSettings {
  layout: string
  authorFormat: string
  renameConfig: main.RenameConfig
  fieldOptions: main.FieldMappingOption[]
  layoutOptions: main.LayoutOption[]
}

const defaultSettings: AppSettings = {
  layout: 'author-series-title',
  authorFormat: 'preserve',
  renameConfig: { enabled: false, template: '', preset: '', separator: '', author_format: '', replace_spaces: false, space_char: '' },
  fieldOptions: [],
  layoutOptions: [],
}

interface SettingsContextValue {
  settings: AppSettings
  refreshSettings: () => void
}

const SettingsContext = createContext<SettingsContextValue>({
  settings: defaultSettings,
  refreshSettings: () => {},
})

export function SettingsProvider({ children }: { children: ReactNode }) {
  const [settings, setSettings] = useState<AppSettings>(defaultSettings)

  const fetchAll = useCallback(async () => {
    try {
      const [layout, authorFormat, renameConfig, fieldOptions, layoutOptions] = await Promise.all([
        GetCurrentLayout(),
        GetCurrentAuthorFormat(),
        GetRenameConfig(),
        GetFieldMappingOptions(),
        GetAvailableLayouts(),
      ])
      setSettings({ layout, authorFormat, renameConfig, fieldOptions, layoutOptions })
    } catch {
      // backend not ready yet (during startup), silently ignore
    }
  }, [])

  // Single poll for the whole app
  useEffect(() => {
    fetchAll()
    const interval = setInterval(fetchAll, 500)
    return () => clearInterval(interval)
  }, [fetchAll])

  return (
    <SettingsContext.Provider value={{ settings, refreshSettings: fetchAll }}>
      {children}
    </SettingsContext.Provider>
  )
}

export function useSettings() {
  return useContext(SettingsContext)
}
