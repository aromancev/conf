export const config = {
  storage: {
    baseURL: import.meta.env.VITE_STORAGE_URL as string,
  },
  rtc: {
    room: {
      baseURL: import.meta.env.VITE_RTC_ROOM_URL as string,
    },
  },
  sfu: {
    stunURLs: split(import.meta.env.VITE_SFU_STUN_URLS),
    turnURLs: split(import.meta.env.VITE_SFU_TURN_URLS),
  },
  gsi: {
    clientId: import.meta.env.VITE_GSI_CLIENT_ID as string,
  },
}

function split(val: string | boolean | undefined): string[] {
  return typeof val === "string" ? val.split(",") : []
}
