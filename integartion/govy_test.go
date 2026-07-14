package integartion

import (
	"context"
	"regexp"
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"

	dc "github.com/pancsta/defer_contracts"
)

type Service struct {
	*dc.Contracts[*Service]

	APIVersion string
	Kind       string
	Metadata   ServiceMetadata
	Spec       ServiceSpec
}

type ServiceMetadata struct {
	Name        string
	DisplayName string
	Project     string
	Labels      Labels
}

type ServiceSpec struct {
	Description string
}

type Labels map[string][]string

const (
	minLabelKeyLength   = 1
	maxLabelKeyLength   = 63
	maxLabelValueLength = 200
)

var labelKeyRegexp = regexp.MustCompile(`^\p{Ll}([_\-0-9\p{Ll}]*[0-9\p{Ll}])?$`)

// fixtures
var serviceDefault = Service{
	APIVersion: "n9/v1alpha",
	Kind:       "Service",
	Metadata: ServiceMetadata{
		Name:        "slo-statusapi",
		DisplayName: "SLO Status API",
		Project:     "default-project",
		Labels: Labels{
			"key": {"value1", "value 2"},
		},
	},
	Spec: ServiceSpec{
		Description: "Status API allows users to retrieve the latest metrics for defined SLOs",
	},
}

func TestGovy(t *testing.T) {
	ctx := context.TODO()
	dc.ContractsEnable(true)

	// init
	serviceInstance := serviceDefault
	serviceInstance.Contracts = dc.NewContracts(&serviceInstance, true)
	defValidatorLabels(serviceInstance)
	processLabels(ctx, &serviceInstance)

	defValidatorApiVersion(serviceInstance)
	processLabels(ctx, &serviceInstance)
	processApiVersion(ctx, &serviceInstance)

	assert.Panic(t, func() {
		// fails on custom post-condition
		_, _ = processPost(ctx, &serviceInstance)
	})
}

func TestGovyPostCond(t *testing.T) {
	ctx := context.TODO()
	dc.ContractsEnable(true)

	// init
	serviceInstance := serviceDefault
	serviceInstance.Contracts = dc.NewContracts(&serviceInstance, true)
	defValidatorLabels(serviceInstance)
	processLabels(ctx, &serviceInstance)

	defValidatorApiVersion(serviceInstance)
	processLabels(ctx, &serviceInstance)
	processApiVersion(ctx, &serviceInstance)
	// data corruption happens here
	serviceInstance.APIVersion = "bad version"

	assert.Panic(t, func() {
		// previously ok, but now fails from an inherited contract
		processLabels(ctx, &serviceInstance)
	})
}

func defValidatorApiVersion(serviceInstance Service) {
	apiVersionValidator := govy.For(func(s Service) string { return s.APIVersion }).
		WithName("apiVersion").
		Required().
		Rules(rules.EQ("n9/v1alpha"))
	serviceInstance.Contracts.Add(func(ctx context.Context, service *Service) {
		if err := apiVersionValidator.Validate(*service); err != nil {
			panic(err)
		}
	})
}

func defValidatorLabels(serviceInstance Service) {
	labelsValidator := govy.New[Labels](
		govy.ForMap(govy.GetSelf[Labels]()).
			RulesForKeys(
				rules.StringLength(minLabelKeyLength, maxLabelKeyLength),
				rules.StringMatchRegexp(labelKeyRegexp),
			).
			IncludeForValues(govy.New[[]string](
				govy.ForSlice(govy.GetSelf[[]string]()).
					Rules(rules.SliceUnique(rules.HashFuncSelf[string]())).
					RulesForEach(
						rules.StringMaxLength(maxLabelValueLength),
					),
			)),
	)
	serviceInstance.Contracts.Add(func(ctx context.Context, service *Service) {
		if err := labelsValidator.Validate(service.Metadata.Labels); err != nil {
			panic(err)
		}
	})
}

func processLabels(ctx context.Context, instance *Service) {
	instance.Contracts.Check(ctx)
	defer instance.Contracts.Check(ctx, func(ctx context.Context, scope *Service) {})
}

func processApiVersion(ctx context.Context, instance *Service) {
	instance.Contracts.Check(ctx)
	defer instance.Contracts.Check(ctx, func(ctx context.Context, scope *Service) {})
}

func processPost(ctx context.Context, instance *Service) (ret string, err error) {
	instance.Contracts.Check(ctx)
	defer instance.Contracts.Check(ctx, func(ctx context.Context, scope *Service) {
		if len(scope.Metadata.Labels) < 2 {
			panic("labels must have at least 2 entries")
		}
	})

	return ret, err
}
