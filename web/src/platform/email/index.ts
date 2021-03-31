const pattern = new RegExp(
  `^([!#-'*+/-9=?A-Z^-~-]+(.[!#-'*+/-9=?A-Z^-~-]+)*|"([]!#-[^-~ \t]|([\t -~]))+")@([!#-'*+/-9=?A-Z^-~-]+(.[!#-'*+/-9=?A-Z^-~-]+)*|[[\t -Z^-~]*])$`,
)

export function isValid(email: string): boolean {
  return pattern.test(email)
}
