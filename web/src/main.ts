import { createApp } from "vue"
import App from "./App.vue"
import router from "./router"
import { client, profileClient } from "@/api"

createApp(App).use(router).mount("#app")

client.refreshToken().then(() => {
  profileClient.refreshProfile()
})
