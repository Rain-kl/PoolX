import type { Backend } from '@/types'
import { useStorage } from '@vueuse/core'
import { isEqual, omit } from 'lodash'
import { v4 as uuid } from 'uuid'
import { computed } from 'vue'
import { sourceIPLabelList } from './settings'

const FIXED_BACKEND_UUID = 'poolx-clash'

const buildFixedBackend = (): Backend => ({
  uuid: FIXED_BACKEND_UUID,
  protocol: window.location.protocol.replace(':', ''),
  host: window.location.hostname,
  port:
    window.location.port ||
    (window.location.protocol === 'https:' ? '443' : '80'),
  secondaryPath: '/api/zashboard/clash',
  password: '',
  label: 'PoolX Clash',
})

export const backendList = useStorage<Backend[]>('setup/api-list', [])
export const activeUuid = useStorage<string>('setup/active-uuid', FIXED_BACKEND_UUID)

const ensureFixedBackend = () => {
  backendList.value = [buildFixedBackend()]
  activeUuid.value = FIXED_BACKEND_UUID
}

ensureFixedBackend()

export const activeBackend = computed(() => backendList.value[0] || buildFixedBackend())

export const addBackend = (backend: Omit<Backend, 'uuid'>) => {
  if (backend.secondaryPath === '/api/zashboard/clash') {
    ensureFixedBackend()
    return
  }
  const currentEnd = backendList.value.find((end) => {
    return isEqual(omit(end, 'uuid'), backend)
  })

  if (currentEnd) {
    activeUuid.value = currentEnd.uuid
    return
  }

  const id = uuid()

  backendList.value.push({
    ...backend,
    uuid: id,
  })
  activeUuid.value = id
}

export const updateBackend = (uuid: string, backend: Omit<Backend, 'uuid'>) => {
  if (uuid === FIXED_BACKEND_UUID) {
    ensureFixedBackend()
    return
  }
  const index = backendList.value.findIndex((end) => end.uuid === uuid)
  if (index !== -1) {
    backendList.value[index] = {
      ...backend,
      uuid,
    }
  }
}

export const removeBackend = (uuid: string) => {
  if (uuid === FIXED_BACKEND_UUID) {
    ensureFixedBackend()
    return
  }
  backendList.value = backendList.value.filter((end) => end.uuid !== uuid)
  sourceIPLabelList.value.forEach((label) => {
    if (label.scope && label.scope.includes(uuid)) {
      label.scope = label.scope.filter((scope) => scope !== uuid)
      if (!label.scope.length) {
        delete label.scope
      }
    }
  })
}
