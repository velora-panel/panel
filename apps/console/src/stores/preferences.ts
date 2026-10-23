import { defineStore } from 'pinia'
import { ref } from 'vue'
export const usePreferences = defineStore('preferences', () => {
  const collapsed = ref(false)
  function toggle() {
    collapsed.value = !collapsed.value
  }
  return { collapsed, toggle }
})
