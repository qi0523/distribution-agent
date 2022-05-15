package reference

import "regexp"

var (
	alphaNumeric = `[a-z0-9]+`

	separator = `(?:[._]|__|[-]*)`

	nameComponent = expression(
		alphaNumeric,
		optional(repeated(separator, alphaNumeric)))

	domainComponent = `(?:[a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9-]*[a-zA-Z0-9])`

	domain = expression(
		domainComponent,
		optional(repeated(literal(`.`), domainComponent)),
		optional(literal(`:`), `[0-9]+`))

	tag = `[\w][\w.-]{0,127}`

	TagRegexp = regexp.MustCompile(tag)

	namePattern = expression(
		optional(domain, literal(`/`)),
		nameComponent,
		optional(repeated(literal(`/`), nameComponent)))

	NameRegexp = regexp.MustCompile(namePattern)
)

func literal(s string) string {
	re := regexp.MustCompile(regexp.QuoteMeta(s))

	if _, complete := re.LiteralPrefix(); !complete {
		panic("must be a literal")
	}

	return re.String()
}

func expression(res ...string) string {
	var s string
	for _, re := range res {
		s += re
	}
	return s
}

func optional(res ...string) string {
	return group(expression(res...)) + `?`
}

func repeated(res ...string) string {
	return group(expression(res...)) + `+`
}

func group(res ...string) string {
	return `(?:` + expression(res...) + `)`
}
