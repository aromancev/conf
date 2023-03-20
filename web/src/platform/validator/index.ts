export class RegexValidator {
  readonly regex: RegExp
  readonly errors: string[]

  constructor(regex: string, errors: string[]) {
    this.regex = new RegExp(regex)
    this.errors = errors
  }

  validate(value: string): string[] {
    if (this.regex.test(value)) {
      return []
    }
    return this.errors
  }
}
