import { RegexValidator } from "@/platform/validator"

export type Confa = {
  id: string
  ownerId: string
  handle: string
  title: string
  description: string
}

export const titleValidator = new RegexValidator(/^[\p{L}0-9\-!? ]{4,64}$/u, [
  "Must be from 4 to 64 characters long",
  "Can only contain letters, numbers, spaces, '-', '!', and '?'",
])

export const handleValidator = new RegexValidator(/^[a-z0-9-]{4,64}$/, [
  "Must be from 4 to 64 characters long",
  "Can only contain lower case letters, numbers, and '-'",
])
