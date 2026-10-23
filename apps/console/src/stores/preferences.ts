import { defineStore } from 'pinia'
import { ref } from 'vue'
export const usePreferences = defineStore('preferences', () => {
  const collapsed = ref(false)
  return { collapsed }
})
