package generator_test

import (
	"strings"
	"testing"

	"github.com/NhaLeTruc/datagen-cli/internal/generator"
	"github.com/stretchr/testify/assert"
)

func TestSemanticDetection(t *testing.T) {
	t.Run("detect email patterns", func(t *testing.T) {
		detector := generator.NewSemanticDetector()

		assert.True(t, detector.IsEmail("email"))
		assert.True(t, detector.IsEmail("user_email"))
		assert.True(t, detector.IsEmail("contact_email"))
		assert.True(t, detector.IsEmail("email_address"))
		assert.False(t, detector.IsEmail("name"))
		assert.False(t, detector.IsEmail("id"))
	})

	t.Run("detect phone patterns", func(t *testing.T) {
		detector := generator.NewSemanticDetector()

		assert.True(t, detector.IsPhone("phone"))
		assert.True(t, detector.IsPhone("phone_number"))
		assert.True(t, detector.IsPhone("mobile"))
		assert.True(t, detector.IsPhone("mobile_phone"))
		assert.True(t, detector.IsPhone("cell_phone"))
		assert.False(t, detector.IsPhone("email"))
	})

	t.Run("detect name patterns", func(t *testing.T) {
		detector := generator.NewSemanticDetector()

		assert.True(t, detector.IsFirstName("first_name"))
		assert.True(t, detector.IsFirstName("firstname"))
		assert.True(t, detector.IsFirstName("given_name"))
		assert.False(t, detector.IsFirstName("last_name"))

		assert.True(t, detector.IsLastName("last_name"))
		assert.True(t, detector.IsLastName("lastname"))
		assert.True(t, detector.IsLastName("surname"))
		assert.False(t, detector.IsLastName("first_name"))

		assert.True(t, detector.IsFullName("name"))
		assert.True(t, detector.IsFullName("full_name"))
		assert.True(t, detector.IsFullName("fullname"))
		assert.False(t, detector.IsFullName("first_name"))
	})

	t.Run("detect address patterns", func(t *testing.T) {
		detector := generator.NewSemanticDetector()

		assert.True(t, detector.IsAddress("address"))
		assert.True(t, detector.IsAddress("street_address"))
		assert.True(t, detector.IsAddress("street"))
		assert.False(t, detector.IsAddress("city"))

		assert.True(t, detector.IsCity("city"))
		assert.True(t, detector.IsCity("town"))
		assert.False(t, detector.IsCity("country"))

		assert.True(t, detector.IsCountry("country"))
		assert.True(t, detector.IsCountry("country_name"))
		assert.False(t, detector.IsCountry("city"))

		assert.True(t, detector.IsPostalCode("postal_code"))
		assert.True(t, detector.IsPostalCode("zip_code"))
		assert.True(t, detector.IsPostalCode("zipcode"))
		assert.True(t, detector.IsPostalCode("postcode"))
		assert.False(t, detector.IsPostalCode("address"))
	})

	t.Run("detect timestamp patterns", func(t *testing.T) {
		detector := generator.NewSemanticDetector()

		assert.True(t, detector.IsCreatedAt("created_at"))
		assert.True(t, detector.IsCreatedAt("created"))
		assert.False(t, detector.IsCreatedAt("updated_at"))

		assert.True(t, detector.IsUpdatedAt("updated_at"))
		assert.True(t, detector.IsUpdatedAt("updated"))
		assert.True(t, detector.IsUpdatedAt("modified_at"))
		assert.False(t, detector.IsUpdatedAt("created_at"))
	})

	t.Run("get semantic type", func(t *testing.T) {
		detector := generator.NewSemanticDetector()

		assert.Equal(t, "email", detector.GetSemanticType("email"))
		assert.Equal(t, "phone", detector.GetSemanticType("phone_number"))
		assert.Equal(t, "first_name", detector.GetSemanticType("first_name"))
		assert.Equal(t, "last_name", detector.GetSemanticType("last_name"))
		assert.Equal(t, "full_name", detector.GetSemanticType("full_name"))
		assert.Equal(t, "address", detector.GetSemanticType("street_address"))
		assert.Equal(t, "city", detector.GetSemanticType("city"))
		assert.Equal(t, "country", detector.GetSemanticType("country"))
		assert.Equal(t, "postal_code", detector.GetSemanticType("zip_code"))
		assert.Equal(t, "", detector.GetSemanticType("random_field"))
	})
}

func TestSemanticGenerators(t *testing.T) {
	t.Run("email generator produces valid emails", func(t *testing.T) {
		gen := generator.NewEmailGenerator()
		ctx := generator.NewContextWithSeed(42)

		for i := 0; i < 20; i++ {
			val, err := gen.Generate(ctx)
			assert.NoError(t, err)

			email, ok := val.(string)
			assert.True(t, ok)
			assert.Contains(t, email, "@")
			assert.NotEmpty(t, strings.Split(email, "@")[0])
			assert.NotEmpty(t, strings.Split(email, "@")[1])
		}
	})

	t.Run("phone generator produces valid phones", func(t *testing.T) {
		gen := generator.NewPhoneGenerator()
		ctx := generator.NewContextWithSeed(42)

		for i := 0; i < 20; i++ {
			val, err := gen.Generate(ctx)
			assert.NoError(t, err)

			phone, ok := val.(string)
			assert.True(t, ok)
			assert.NotEmpty(t, phone)
			// Should contain digits
			assert.Regexp(t, `\d`, phone)
		}
	})

	t.Run("first name generator produces names", func(t *testing.T) {
		gen := generator.NewFirstNameGenerator()
		ctx := generator.NewContextWithSeed(42)

		for i := 0; i < 20; i++ {
			val, err := gen.Generate(ctx)
			assert.NoError(t, err)

			name, ok := val.(string)
			assert.True(t, ok)
			assert.NotEmpty(t, name)
			assert.Greater(t, len(name), 1)
		}
	})

	t.Run("last name generator produces names", func(t *testing.T) {
		gen := generator.NewLastNameGenerator()
		ctx := generator.NewContextWithSeed(42)

		for i := 0; i < 20; i++ {
			val, err := gen.Generate(ctx)
			assert.NoError(t, err)

			name, ok := val.(string)
			assert.True(t, ok)
			assert.NotEmpty(t, name)
			assert.Greater(t, len(name), 1)
		}
	})

	t.Run("city generator produces cities", func(t *testing.T) {
		gen := generator.NewCityGenerator()
		ctx := generator.NewContextWithSeed(42)

		for i := 0; i < 20; i++ {
			val, err := gen.Generate(ctx)
			assert.NoError(t, err)

			city, ok := val.(string)
			assert.True(t, ok)
			assert.NotEmpty(t, city)
		}
	})

	t.Run("country generator produces countries", func(t *testing.T) {
		gen := generator.NewCountryGenerator()
		ctx := generator.NewContextWithSeed(42)

		for i := 0; i < 20; i++ {
			val, err := gen.Generate(ctx)
			assert.NoError(t, err)

			country, ok := val.(string)
			assert.True(t, ok)
			assert.NotEmpty(t, country)
		}
	})
}