package db

type Role string

const (
	RoleUser       Role = "user"
	RoleAssistant Role = "assistant"
	RoleSystem    Role = "system"
	RoleTool      Role = "tool"
)

func (r Role) String() string { return string(r) }

type Provider string

const (
	ProviderOpenAI     Provider = "openai"
	ProviderAnthropic Provider = "anthropic"
	ProviderGoogle    Provider = "google"
	ProviderOllama    Provider = "ollama"
	ProviderLocal    Provider = "local"
)

func (p Provider) String() string { return string(p) }

type Model struct {
	Name         string   `json:"name"`
	Provider    Provider `json:"provider"`
	InputCost   float64  `json:"input_cost"`
	OutputCost  float64  `json:"output_cost"`
}

var DefaultModels = []Model{
	{Name: "gpt-4o", Provider: ProviderOpenAI, InputCost: 0.0025, OutputCost: 0.01},
	{Name: "gpt-4o-mini", Provider: ProviderOpenAI, InputCost: 0.00015, OutputCost: 0.0006},
	{Name: "gpt-4-turbo", Provider: ProviderOpenAI, InputCost: 0.01, OutputCost: 0.03},
	{Name: "gpt-3.5-turbo", Provider: ProviderOpenAI, InputCost: 0.0005, OutputCost: 0.0015},
	{Name: "claude-3-5-sonnet", Provider: ProviderAnthropic, InputCost: 0.003, OutputCost: 0.015},
	{Name: "claude-3-opus", Provider: ProviderAnthropic, InputCost: 0.015, OutputCost: 0.075},
	{Name: "claude-3-haiku", Provider: ProviderAnthropic, InputCost: 0.00025, OutputCost: 0.00125},
	{Name: "gemini-1.5-pro", Provider: ProviderGoogle, InputCost: 0.00125, OutputCost: 0.005},
	{Name: "gemini-1.5-flash", Provider: ProviderGoogle, InputCost: 0.000075, OutputCost: 0.0003},
	{Name: "llama-3.1-70b", Provider: ProviderOllama, InputCost: 0, OutputCost: 0},
	{Name: "llama-3.1-8b", Provider: ProviderOllama, InputCost: 0, OutputCost: 0},
	{Name: "mistral-7b", Provider: ProviderOllama, InputCost: 0, OutputCost: 0},
}

func GetModelCost(modelName string) (float64, float64) {
	for _, m := range DefaultModels {
		if m.Name == modelName {
			return m.InputCost, m.OutputCost
		}
	}
	return 0, 0
}