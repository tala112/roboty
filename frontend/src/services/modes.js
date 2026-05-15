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
  GetAllDetectableApps,
  CheckAppOnPC,
  AddAllowedApp,
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

  async create({ name, description, durationMinutes, muteNotifications, enabled, icon, color, apps, allowedUrls }) {
    const raw = await CreateMode(
      name,
      description,
      durationMinutes,
      muteNotifications,
      icon,
      color,
      JSON.stringify(apps),
      JSON.stringify(allowedUrls || [])
    )
    return raw ? JSON.parse(raw) : null
  },

  async update({ id, name, description, durationMinutes, muteNotifications, enabled, icon, color, apps, allowedUrls }) {
    const raw = await UpdateMode(
      id,
      name,
      description,
      durationMinutes,
      muteNotifications,
      enabled,
      icon,
      color,
      JSON.stringify(apps),
      JSON.stringify(allowedUrls || [])
    )
    return raw ? JSON.parse(raw) : null
  },

  async delete(id) {
    return DeleteMode(id)
  },

  async toggle(id, enabled) {
    return ToggleMode(id, enabled)
  },

  async getAllDetectableApps() {
    const raw = await GetAllDetectableApps()
    return raw ? JSON.parse(raw) : []
  },

  async checkAppOnPC(appExec) {
    return CheckAppOnPC(appExec)
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

  async addAllowedApp(modeId, appName, appExec, category, force) {
    const raw = await AddAllowedApp(modeId, appName, appExec, category || "productive", force || false)
    return raw ? JSON.parse(raw) : null
  },
}
