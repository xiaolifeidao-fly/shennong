/// <reference types="@dcloudio/types" />

declare module '*.vue' {
  import type { DefineComponent } from 'vue'

  const component: DefineComponent<object, object, unknown>
  export default component
}

interface ImportMetaEnv {
  readonly VITE_APP_API_TRANSPORT?: 'cloud' | 'https'
  readonly VITE_APP_API_BASE_URL: string
  readonly VITE_APP_CLOUD_ENV?: string
  readonly VITE_APP_CLOUD_SERVICE?: string
  readonly VITE_APP_CLOUD_FILE_BASE_URL?: string
  readonly VITE_APP_CLOUD_STORAGE_PREFIX?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
