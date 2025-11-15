package generator

import (
	"strings"

	"github.com/brianvoe/gofakeit/v6"
)

// SemanticDetector detects semantic meaning from column names
type SemanticDetector struct{}

func NewSemanticDetector() *SemanticDetector {
	return &SemanticDetector{}
}

// GetSemanticType returns the semantic type for a column name
func (d *SemanticDetector) GetSemanticType(columnName string) string {
	lower := strings.ToLower(columnName)

	if d.IsEmail(lower) {
		return "email"
	}
	if d.IsPhone(lower) {
		return "phone"
	}
	if d.IsFirstName(lower) {
		return "first_name"
	}
	if d.IsLastName(lower) {
		return "last_name"
	}
	if d.IsFullName(lower) {
		return "full_name"
	}
	if d.IsAddress(lower) {
		return "address"
	}
	if d.IsCity(lower) {
		return "city"
	}
	if d.IsCountry(lower) {
		return "country"
	}
	if d.IsPostalCode(lower) {
		return "postal_code"
	}
	if d.IsCreatedAt(lower) {
		return "created_at"
	}
	if d.IsUpdatedAt(lower) {
		return "updated_at"
	}

	return ""
}

func (d *SemanticDetector) IsEmail(name string) bool {
	return strings.Contains(name, "email")
}

func (d *SemanticDetector) IsPhone(name string) bool {
	return strings.Contains(name, "phone") || strings.Contains(name, "mobile") || strings.Contains(name, "cell")
}

func (d *SemanticDetector) IsFirstName(name string) bool {
	return strings.Contains(name, "first") && strings.Contains(name, "name") ||
		strings.Contains(name, "firstname") ||
		strings.Contains(name, "given") && strings.Contains(name, "name")
}

func (d *SemanticDetector) IsLastName(name string) bool {
	return strings.Contains(name, "last") && strings.Contains(name, "name") ||
		strings.Contains(name, "lastname") ||
		strings.Contains(name, "surname") ||
		strings.Contains(name, "family") && strings.Contains(name, "name")
}

func (d *SemanticDetector) IsFullName(name string) bool {
	// Must be exactly "name" or "full_name", not "first_name" or "last_name"
	return (name == "name" || name == "full_name" || name == "fullname") &&
		!d.IsFirstName(name) && !d.IsLastName(name)
}

func (d *SemanticDetector) IsAddress(name string) bool {
	return (strings.Contains(name, "address") || name == "street") &&
		!d.IsCity(name) && !d.IsCountry(name) && !d.IsPostalCode(name)
}

func (d *SemanticDetector) IsCity(name string) bool {
	return name == "city" || name == "town"
}

func (d *SemanticDetector) IsCountry(name string) bool {
	return strings.Contains(name, "country")
}

func (d *SemanticDetector) IsPostalCode(name string) bool {
	return strings.Contains(name, "postal") || strings.Contains(name, "zip") || name == "postcode"
}

func (d *SemanticDetector) IsCreatedAt(name string) bool {
	return strings.Contains(name, "created") && !strings.Contains(name, "updated")
}

func (d *SemanticDetector) IsUpdatedAt(name string) bool {
	return strings.Contains(name, "updated") || strings.Contains(name, "modified")
}

// Semantic Generators using gofakeit

type EmailGenerator struct{}

func NewEmailGenerator() *EmailGenerator {
	return &EmailGenerator{}
}

func (g *EmailGenerator) Generate(ctx *Context) (interface{}, error) {
	faker := gofakeit.New(ctx.Rand.Int63())
	return faker.Email(), nil
}

func (g *EmailGenerator) Name() string {
	return "email"
}

type PhoneGenerator struct{}

func NewPhoneGenerator() *PhoneGenerator {
	return &PhoneGenerator{}
}

func (g *PhoneGenerator) Generate(ctx *Context) (interface{}, error) {
	faker := gofakeit.New(ctx.Rand.Int63())
	return faker.Phone(), nil
}

func (g *PhoneGenerator) Name() string {
	return "phone"
}

type FirstNameGenerator struct{}

func NewFirstNameGenerator() *FirstNameGenerator {
	return &FirstNameGenerator{}
}

func (g *FirstNameGenerator) Generate(ctx *Context) (interface{}, error) {
	faker := gofakeit.New(ctx.Rand.Int63())
	return faker.FirstName(), nil
}

func (g *FirstNameGenerator) Name() string {
	return "first_name"
}

type LastNameGenerator struct{}

func NewLastNameGenerator() *LastNameGenerator {
	return &LastNameGenerator{}
}

func (g *LastNameGenerator) Generate(ctx *Context) (interface{}, error) {
	faker := gofakeit.New(ctx.Rand.Int63())
	return faker.LastName(), nil
}

func (g *LastNameGenerator) Name() string {
	return "last_name"
}

type FullNameGenerator struct{}

func NewFullNameGenerator() *FullNameGenerator {
	return &FullNameGenerator{}
}

func (g *FullNameGenerator) Generate(ctx *Context) (interface{}, error) {
	faker := gofakeit.New(ctx.Rand.Int63())
	return faker.Name(), nil
}

func (g *FullNameGenerator) Name() string {
	return "full_name"
}

type AddressGenerator struct{}

func NewAddressGenerator() *AddressGenerator {
	return &AddressGenerator{}
}

func (g *AddressGenerator) Generate(ctx *Context) (interface{}, error) {
	faker := gofakeit.New(ctx.Rand.Int63())
	return faker.Street(), nil
}

func (g *AddressGenerator) Name() string {
	return "address"
}

type CityGenerator struct{}

func NewCityGenerator() *CityGenerator {
	return &CityGenerator{}
}

func (g *CityGenerator) Generate(ctx *Context) (interface{}, error) {
	faker := gofakeit.New(ctx.Rand.Int63())
	return faker.City(), nil
}

func (g *CityGenerator) Name() string {
	return "city"
}

type CountryGenerator struct{}

func NewCountryGenerator() *CountryGenerator {
	return &CountryGenerator{}
}

func (g *CountryGenerator) Generate(ctx *Context) (interface{}, error) {
	faker := gofakeit.New(ctx.Rand.Int63())
	return faker.Country(), nil
}

func (g *CountryGenerator) Name() string {
	return "country"
}

type PostalCodeGenerator struct{}

func NewPostalCodeGenerator() *PostalCodeGenerator {
	return &PostalCodeGenerator{}
}

func (g *PostalCodeGenerator) Generate(ctx *Context) (interface{}, error) {
	faker := gofakeit.New(ctx.Rand.Int63())
	return faker.Zip(), nil
}

func (g *PostalCodeGenerator) Name() string {
	return "postal_code"
}

type CreatedAtGenerator struct{}

func NewCreatedAtGenerator() *CreatedAtGenerator {
	return &CreatedAtGenerator{}
}

func (g *CreatedAtGenerator) Generate(ctx *Context) (interface{}, error) {
	faker := gofakeit.New(ctx.Rand.Int63())
	return faker.DateRange(
		faker.Date().AddDate(-1, 0, 0),
		faker.Date(),
	), nil
}

func (g *CreatedAtGenerator) Name() string {
	return "created_at"
}

type UpdatedAtGenerator struct{}

func NewUpdatedAtGenerator() *UpdatedAtGenerator {
	return &UpdatedAtGenerator{}
}

func (g *UpdatedAtGenerator) Generate(ctx *Context) (interface{}, error) {
	faker := gofakeit.New(ctx.Rand.Int63())
	return faker.DateRange(
		faker.Date().AddDate(0, -3, 0),
		faker.Date(),
	), nil
}

func (g *UpdatedAtGenerator) Name() string {
	return "updated_at"
}