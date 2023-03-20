import { RegexValidator } from "@/platform/validator"

export enum Platform {
  EMAIL = "EMAIL",
  GITHUB = "GITHUB",
  TWITTER = "TWITTER",
}

export type Identifier = {
  platform: Platform
  value: string
}

export type User = {
  id: string
  identifiers: Identifier[]
  hasPassword: boolean
}

export const emailValidator = new RegexValidator(
  "^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9-]+(?:.[a-zA-Z0-9-]+)*$",
  ["Must be a valid email"],
)

export const passwordValidator = new RegexValidator("^[^ \t].{6,64}[^ \t]$", [
  "Must be from 8 to 64 charachters long",
  "Cannot start or end with a space",
])
