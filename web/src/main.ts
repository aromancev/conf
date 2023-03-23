import { createApp, watch } from "vue"
import App from "./App.vue"
import router from "./router"
import { api, profileClient } from "@/api"
import { accessStore, Account } from "@/api/models/access"
import { gsiPromptOneTap } from "@/components/gsi"

const app = createApp(App)

app.directive("click-outside", {
  mounted(el, binding) {
    el.__vueClickOutsideEventHandler__ = function (event: Event) {
      if (el !== event.target && !el.contains(event.target)) {
        binding.value(event, el)
      }
    }
    document.body.addEventListener("click", el.__vueClickOutsideEventHandler__)
  },
  unmounted(el) {
    document.body.removeEventListener("click", el.__vueClickOutsideEventHandler__)
  },
})

app.use(router)
app.mount("#app")

watch(accessStore.state, () => {
  if (accessStore.state.account === Account.Guest) {
    gsiPromptOneTap()
  } else {
    profileClient.refreshProfile()
  }
})
api.refreshToken()
