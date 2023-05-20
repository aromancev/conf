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
    url: import.meta.env.VITE_SFU_URL as string,
  },
  gsi: {
    clientId: import.meta.env.VITE_GSI_CLIENT_ID as string,
  },
}
