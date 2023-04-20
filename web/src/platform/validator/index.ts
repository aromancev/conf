export class RegexValidator {
  readonly regex: RegExp
  readonly errors: string[]

  constructor(regex: RegExp, errors: string[]) {
    this.regex = regex
    this.errors = errors
  }

  validate(value: string): string[] {
    if (this.regex.test(value)) {
      return []
    }
    return [...this.errors]
  }
}
