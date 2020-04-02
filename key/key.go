package key

// Generator defines a key generator.
type Generator interface {
	// Generate generates key from given parts or templates as parameter.
	Generate(parts ...string) (string, error)

	// Check checks if generator functional.
	Check() error
}
