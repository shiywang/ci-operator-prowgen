package rehearse

import (
	"reflect"
	"sort"
	"testing"

	templateapi "github.com/openshift/api/template/v1"
	"github.com/openshift/ci-operator/pkg/api"
	"k8s.io/apimachinery/pkg/util/diff"
	prowconfig "k8s.io/test-infra/prow/config"

	"github.com/openshift/ci-operator-prowgen/pkg/config"
)

func TestRecordChangedCiopConfigs(t *testing.T) {
	testFilename := ""

	testCases := []struct {
		description string
		configs     []string
		expected    []string
	}{{
		description: "no changed configs",
		expected:    []string{},
	}, {
		description: "changed configs",
		configs:     []string{"org-repo-branch.yaml", "another-org-repo-branch.yaml"},
		expected:    []string{"another-org-repo-branch.yaml", "org-repo-branch.yaml"},
	}}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			metrics := NewMetrics(testFilename)
			testCiopConfig := config.CompoundCiopConfig{}
			for _, ciopConfig := range tc.configs {
				testCiopConfig[ciopConfig] = &api.ReleaseBuildConfiguration{}
			}
			metrics.RecordChangedCiopConfigs(testCiopConfig)
			sort.Strings(metrics.ChangedCiopConfigs)
			if !reflect.DeepEqual(tc.expected, metrics.ChangedCiopConfigs) {
				t.Errorf("Recorded changed ci-operator configs differ from expected:\n%s", diff.ObjectReflectDiff(tc.expected, metrics.ChangedCiopConfigs))
			}
		})
	}
}

func TestRecordChangedTemplates(t *testing.T) {
	testFilename := ""

	testCases := []struct {
		description string
		templates   []string
		expected    []string
	}{{
		description: "no changed templates",
		expected:    []string{},
	}, {
		description: "changed templates",
		templates:   []string{"awesome-openshift-installer.yaml", "old-ugly-ansible-installer.yaml"},
		expected:    []string{"awesome-openshift-installer.yaml", "old-ugly-ansible-installer.yaml"},
	}}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			metrics := NewMetrics(testFilename)
			testTemplates := config.CiTemplates{}
			for _, ciopConfig := range tc.templates {
				testTemplates[ciopConfig] = &templateapi.Template{}
			}
			metrics.RecordChangedTemplates(testTemplates)
			sort.Strings(metrics.ChangedTemplates)
			if !reflect.DeepEqual(tc.expected, metrics.ChangedTemplates) {
				t.Errorf("Recorded changed templates differ from expected:\n%s", diff.ObjectReflectDiff(tc.expected, metrics.ChangedTemplates))
			}
		})
	}
}

func TestRecordChangedPresubmits(t *testing.T) {
	testFilename := ""

	var testCases = []struct {
		description string
		presubmits  map[string][]string
		expected    []string
	}{{
		description: "no changed presubmits",
		expected:    []string{},
	}, {
		description: "changed in a single repo",
		presubmits:  map[string][]string{"org/repo": {"org-repo-job", "org-repo-another-job"}},
		expected:    []string{"org-repo-another-job", "org-repo-job"},
	}, {
		description: "changed in multiple repos",
		presubmits: map[string][]string{
			"org/repo":         {"org-repo-job", "org-repo-another-job"},
			"org/another-repo": {"org-another-repo-job"},
		},
		expected: []string{"org-another-repo-job", "org-repo-another-job", "org-repo-job"},
	},
	}
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			metrics := NewMetrics(testFilename)
			testPresubmits := config.Presubmits{}
			for repo, repoPresubmits := range tc.presubmits {
				testPresubmits[repo] = []prowconfig.Presubmit{}
				for _, presubmit := range repoPresubmits {
					testPresubmits[repo] = append(testPresubmits[repo], prowconfig.Presubmit{JobBase: prowconfig.JobBase{Name: presubmit}})
				}

			}
			metrics.RecordChangedPresubmits(testPresubmits)
			sort.Strings(metrics.ChangedPresubmits)
			if !reflect.DeepEqual(tc.expected, metrics.ChangedPresubmits) {
				t.Errorf("Recorded changed presubmits differ from expected:\n%s", diff.ObjectReflectDiff(tc.expected, metrics.ChangedPresubmits))
			}
		})
	}
}

func TestRecordOpportunity(t *testing.T) {
	testFilename := ""

	var testCases = []struct {
		description string
		existing    map[string][]string
		presubmits  map[string][]string
		reason      string
		expected    map[string][]string
	}{{
		description: "no opportunities",
		existing:    map[string][]string{},
		reason:      "no reason",
		expected:    map[string][]string{},
	}, {
		description: "opportunity in a single repo",
		existing:    map[string][]string{},
		presubmits:  map[string][]string{"org/repo": {"org-repo-job", "org-repo-another-job"}},
		reason:      "something changed",
		expected: map[string][]string{
			"org-repo-another-job": {"something changed"},
			"org-repo-job":         {"something changed"},
		},
	}, {
		description: "opportunities in multiple repos",
		existing:    map[string][]string{},
		presubmits: map[string][]string{
			"org/repo":         {"org-repo-job", "org-repo-another-job"},
			"org/another-repo": {"org-another-repo-job"},
		},
		reason: "something changed",
		expected: map[string][]string{
			"org-another-repo-job": {"something changed"},
			"org-repo-another-job": {"something changed"},
			"org-repo-job":         {"something changed"},
		},
	}, {
		description: "opportunities for multiple reasons",
		existing:    map[string][]string{"org-repo-job": {"something changed"}},
		presubmits:  map[string][]string{"org/repo": {"org-repo-job"}},
		reason:      "something else changed",
		expected: map[string][]string{
			"org-repo-job": {"something changed", "something else changed"},
		},
	}}
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			metrics := NewMetrics(testFilename)
			testPresubmits := config.Presubmits{}
			for repo, repoPresubmits := range tc.presubmits {
				testPresubmits[repo] = []prowconfig.Presubmit{}
				for _, presubmit := range repoPresubmits {
					testPresubmits[repo] = append(testPresubmits[repo], prowconfig.Presubmit{JobBase: prowconfig.JobBase{Name: presubmit}})
				}

			}
			metrics.Opportunities = tc.existing
			metrics.RecordOpportunity(testPresubmits, tc.reason)
			if !reflect.DeepEqual(tc.expected, metrics.Opportunities) {
				t.Errorf("Recorded rehearsal opportunities differ from expected:\n%s", diff.ObjectReflectDiff(tc.expected, metrics.Opportunities))
			}
		})
	}
}

func TestRecordActual(t *testing.T) {
	testFilename := ""
	testCases := []struct {
		description string
		jobs        []string
		expected    []string
	}{{
		description: "no actual rehearsals",
		expected:    []string{},
	}, {
		description: "actual rehearsals are recorded",
		jobs:        []string{"rehearse-org-repo-job", "rehearse-org-repo-another-job"},
		expected:    []string{"rehearse-org-repo-another-job", "rehearse-org-repo-job"},
	}}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			metrics := NewMetrics(testFilename)
			presubmits := []*prowconfig.Presubmit{}
			for _, name := range tc.jobs {
				presubmits = append(presubmits, &prowconfig.Presubmit{JobBase: prowconfig.JobBase{Name: name}})
			}
			metrics.RecordActual(presubmits)
			sort.Strings(metrics.Actual)
			if !reflect.DeepEqual(tc.expected, metrics.Actual) {
				t.Errorf("Recorded rehearsals differ from expected:\n%s", diff.ObjectReflectDiff(tc.expected, metrics.Actual))
			}
		})
	}
}
