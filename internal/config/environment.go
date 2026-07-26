package config

type Environment string

const (
	EnvironmentDevelopment Environment = "development"
	EnvironmentTest        Environment = "test"
	EnvironmentProduction  Environment = "production"
)

func (e Environment) IsProduction() bool {
	return e == EnvironmentProduction
}

func (e Environment) IsValid() bool {
	switch e {
	case EnvironmentDevelopment, EnvironmentTest, EnvironmentProduction:
		return true
	default:
		return false
	}
}
