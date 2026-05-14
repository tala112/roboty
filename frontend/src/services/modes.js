import {
  ListModes,
  CreateMode,
  UpdateMode,
  DeleteMode,
  ToggleMode,
  GetInstalledApps,
  ActivateMode,
  DeactivateMode,
  GetActiveSession,
} from "../../wailsjs/go/main/App"

export const modesService = {
  async list() {
    const raw = await ListModes()
    return raw ? JSON.parse(raw) : []
  },

  async get(id) {
    const modes = await this.list()
    return modes.find(m => m.id === id) || null
  },

  async create({ name, description, durationMinutes, muteNotifications, enabled, icon, color, apps }) {
    const raw = await CreateMode(
      name,
      description,
      durationMinutes,
      muteNotifications,
      icon,
      color,
      JSON.stringify(apps)
    )
    return raw ? JSON.parse(raw) : null
  },

  async update({ id, name, description, durationMinutes, muteNotifications, enabled, icon, color, apps }) {
    const raw = await UpdateMode(
      id,
      name,
      description,
      durationMinutes,
      muteNotifications,
      enabled,
      icon,
      color,
      JSON.stringify(apps)
    )
    return raw ? JSON.parse(raw) : null
  },

  async delete(id) {
    return DeleteMode(id)
  },

  async toggle(id, enabled) {
    return ToggleMode(id, enabled)
  },

  async getInstalledApps() {
    const raw = await GetInstalledApps()
    return raw ? JSON.parse(raw) : []
  },

  async activate(modeId) {
    const raw = await ActivateMode(modeId)
    return raw ? JSON.parse(raw) : null
  },

  async deactivate(sessionId) {
    return DeactivateMode(sessionId)
  },

  async getActiveSession() {
    const raw = await GetActiveSession()
    if (!raw) return null
    return JSON.parse(raw)
  },
}
