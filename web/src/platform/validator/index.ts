export class RegexValidator {
  readonly regex: RegExp
  readonly error: string

  constructor(regex: string, errors: string[]) {
    this.regex = new RegExp(regex)
    this.error = errors.map((err: string) => "• " + err).join("\n")
  }

  validate(value: string): string {
    if (this.regex.test(value)) {
      return ""
    }
    return this.error
  }
}
