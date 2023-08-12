export const config = {
  storage: {
    baseURL: import.meta.env.VITE_STORAGE_URL as string,
  },
  api: {
    graphURL: import.meta.env.VITE_API_GRAPH_URL as string,
    iamURL: import.meta.env.VITE_API_IAM_URL as string,
    roomRTCURL: import.meta.env.VITE_API_ROOM_RTC_URL as string,
  },
  sfu: {
    url: import.meta.env.VITE_SFU_URL as string,
  },
  gsi: {
    clientId: import.meta.env.VITE_GSI_CLIENT_ID as string,
  },
}
