package generator

import (
	"fmt"
	"time"
)

// IntegerGenerator generates random integer values
type IntegerGenerator struct{}

func NewIntegerGenerator() *IntegerGenerator {
	return &IntegerGenerator{}
}

func (g *IntegerGenerator) Generate(ctx *Context) (interface{}, error) {
	// Generate integers in a reasonable range
	return int64(ctx.Rand.Int31()), nil
}

func (g *IntegerGenerator) Name() string {
	return "integer"
}

// VarcharGenerator generates random string values
type VarcharGenerator struct {
	maxLength int
}

func NewVarcharGenerator(maxLength int) *VarcharGenerator {
	return &VarcharGenerator{maxLength: maxLength}
}

func (g *VarcharGenerator) Generate(ctx *Context) (interface{}, error) {
	// Generate random string with varying length
	length := ctx.Rand.Intn(g.maxLength) + 1
	if length > g.maxLength {
		length = g.maxLength
	}

	// Generate random characters
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789 "
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = chars[ctx.Rand.Intn(len(chars))]
	}

	return string(result), nil
}

func (g *VarcharGenerator) Name() string {
	return "varchar"
}

// TextGenerator generates longer random text values
type TextGenerator struct{}

func NewTextGenerator() *TextGenerator {
	return &TextGenerator{}
}

func (g *TextGenerator) Generate(ctx *Context) (interface{}, error) {
	// Generate text between 10 and 500 characters
	length := ctx.Rand.Intn(490) + 10

	words := []string{
		"lorem", "ipsum", "dolor", "sit", "amet", "consectetur",
		"adipiscing", "elit", "sed", "do", "eiusmod", "tempor",
		"incididunt", "ut", "labore", "et", "dolore", "magna",
		"aliqua", "enim", "ad", "minim", "veniam", "quis",
	}

	result := ""
	for len(result) < length {
		if len(result) > 0 {
			result += " "
		}
		result += words[ctx.Rand.Intn(len(words))]
	}

	// Trim to exact length
	if len(result) > length {
		result = result[:length]
	}

	return result, nil
}

func (g *TextGenerator) Name() string {
	return "text"
}

// TimestampGenerator generates random timestamp values
type TimestampGenerator struct{}

func NewTimestampGenerator() *TimestampGenerator {
	return &TimestampGenerator{}
}

func (g *TimestampGenerator) Generate(ctx *Context) (interface{}, error) {
	// Generate timestamps within the past year
	now := time.Now()
	pastYear := now.AddDate(-1, 0, 0)

	// Random timestamp between past year and now
	diff := now.Unix() - pastYear.Unix()
	randomSeconds := ctx.Rand.Int63n(diff)

	timestamp := pastYear.Add(time.Duration(randomSeconds) * time.Second)
	return timestamp, nil
}

func (g *TimestampGenerator) Name() string {
	return "timestamp"
}

// BooleanGenerator generates random boolean values
type BooleanGenerator struct{}

func NewBooleanGenerator() *BooleanGenerator {
	return &BooleanGenerator{}
}

func (g *BooleanGenerator) Generate(ctx *Context) (interface{}, error) {
	return ctx.Rand.Intn(2) == 1, nil
}

func (g *BooleanGenerator) Name() string {
	return "boolean"
}

// SerialGenerator generates auto-incrementing sequence values
type SerialGenerator struct{}

func NewSerialGenerator() *SerialGenerator {
	return &SerialGenerator{}
}

func (g *SerialGenerator) Generate(ctx *Context) (interface{}, error) {
	// Get current sequence value from context
	key := fmt.Sprintf("serial_%s_%s", ctx.TableName, ctx.ColumnName)
	val, exists := ctx.Get(key)

	var currentVal int64
	if exists {
		currentVal = val.(int64)
	}

	// Increment and store
	currentVal++
	ctx.Set(key, currentVal)

	return currentVal, nil
}

func (g *SerialGenerator) Name() string {
	return "serial"
}