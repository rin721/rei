package app

import (
	"github.com/rin721/rei/internal/repository"
	"github.com/rin721/rei/internal/service"
	sampleservice "github.com/rin721/rei/internal/service/sample"
)

type sampleModuleProvider struct{}

func (sampleModuleProvider) Provide(_ businessProvisioning, repos *repository.Set) (service.SampleService, error) {
	return sampleservice.New(sampleservice.Dependencies{
		Samples: repos.Samples,
		Demos:   sampleservice.DefaultToolkitDemos(),
	})
}
