package caskin_test

import (
	"github.com/awatercolorpen/caskin/example"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestExecutorDomain(t *testing.T) {
	stage, _ := getStage(t)
	provider := example.Provider{
		User:   stage.SuperadminUser,
		Domain: stage.Domain,
	}
	executor := stage.Caskin.GetExecutor(provider)

	domain := &example.Domain{Name: "domain_02"}
	assert.NoError(t, executor.CreateDomain(domain))
	assert.NoError(t, executor.DeleteDomain(domain))
	domains1, err := executor.GetAllDomain()
	assert.NoError(t, err)
	assert.Len(t, domains1, 1)

	assert.NoError(t, executor.RecoverDomain(domain))
	domains, err := executor.GetAllDomain()
	assert.NoError(t, err)
	assert.Len(t, domains, 2)

	domain.Name = "domain_02_new_name"
	assert.NoError(t, executor.UpdateDomain(domain))

	domain0 := &example.Domain{ID:3, Name: "domain_03"}
	assert.Error(t, executor.UpdateDomain(domain0))
}
