package processor

import (
	"fmt"
	"hitman/api/v1alpha1"
	"hitman/internal/globals"
	"hitman/internal/kubernetes"
	"hitman/internal/template"
	"reflect"
	"regexp"
	"slices"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type Processor struct {
	Client *dynamic.DynamicClient
}

func NewProcessor() (processor *Processor, err error) {

	client, err := kubernetes.NewClient()
	if err != nil {
		return processor, err
	}

	return &Processor{
		Client: client,
	}, err
}

// TODO
func (p *Processor) SyncResources() (err error) {

	for _, configResource := range globals.ExecContext.Config.Spec.Resources {

		// Matching a name is required
		if reflect.ValueOf(configResource.Target.Name).IsZero() {
			globals.ExecContext.Logger.Infof("target name or namespace selector is missing for group '%s' version '%s' resource '%s'. Skipping",
				configResource.Target.Group, configResource.Target.Version, configResource.Target.Resource)
			continue
		}

		// Matching a name is required
		if configResource.Target.Name.MatchExact != "" && configResource.Target.Name.MatchRegex != "" {
			globals.ExecContext.Logger.Infof("target name can only have one selector: matchExact or matchRegex. Skipping")
			continue
		}

		if configResource.Target.Namespace.MatchExact != "" && configResource.Target.Namespace.MatchRegex != "" {
			globals.ExecContext.Logger.Infof("targets namespace can only have one selector: matchExact or matchRegex. Skipping")
			continue
		}

		// Get the resources of the target type
		gvr := schema.GroupVersionResource{
			Group:    configResource.Target.Group,
			Version:  configResource.Target.Version,
			Resource: configResource.Target.Resource,
		}

		resourceRaw := p.Client.Resource(gvr)

		if configResource.Target.Namespace.MatchExact != "" {
			resourceRaw.Namespace(configResource.Target.Namespace.MatchExact)
		}

		resourceList, err := resourceRaw.List(globals.ExecContext.Context, v1.ListOptions{})
		if err != nil {
			globals.ExecContext.Logger.Infof("error listing resources of type '%s' in namespace '%s': %s",
				gvr.String(), configResource.Target.Namespace, err)
			continue
		}

		//
		compiledRegex, err := regexp.Compile(configResource.Target.Name.MatchRegex)
		if err != nil {
			globals.ExecContext.Logger.Infof("error compiling regular expression '%s' for resource name: %s", configResource.Target.Name.MatchRegex, err)
			continue
		}

		compiledRegexNamespace, err := regexp.Compile(configResource.Target.Namespace.MatchRegex)
		if err != nil {
			globals.ExecContext.Logger.Infof("error compiling regular expression '%s' for resource namespace: %s", configResource.Target.Name.MatchRegex, err)
			continue
		}

		// Perform the actions over the resources
		for _, resource := range resourceList.Items {

			// Matching namespace by regex and resource does NOT meet? Skip
			if configResource.Target.Namespace.MatchRegex != "" &&
				!compiledRegexNamespace.MatchString(resource.GetNamespace()) {
				continue
			}

			// Matching name by exact string and resource does NOT meet? Skip
			if configResource.Target.Name.MatchExact != "" &&
				resource.GetName() != configResource.Target.Name.MatchExact {
				continue
			}

			// Matching name by regex and resource does NOT meet? Skip
			if configResource.Target.Name.MatchRegex != "" &&
				!compiledRegex.MatchString(resource.GetName()) {
				continue
			}

			// Process this object. Delete in case of success
			objectDeleted, err := p.processObject(gvr, resource, configResource.Conditions)
			if err != nil {
				globals.ExecContext.Logger.Infof("error processing object: %s", err)
				continue
			}

			if !objectDeleted {
				globals.ExecContext.Logger.Infof("resource '%s' in namespace '%s' did NOT meet the conditions",
					resource.GetName(), resource.GetNamespace())
				continue
			}

			globals.ExecContext.Logger.Infof("resource '%s' in namespace '%s' was deleted successfully", resource.GetName(), resource.GetNamespace())
		}
	}

	return err
}

// processObject process an object coming from arguments.
// It computes templating, evaluates conditions and decides whether to delete it or not.
func (p *Processor) processObject(gvr schema.GroupVersionResource, object unstructured.Unstructured, conditionList []v1alpha1.ConditionT) (result bool, err error) {

	globals.ExecContext.Logger.Debugf("processing object: group: '%s', version: '%s', resource: '%s', name: '%s', namespace: '%s'",
		gvr.Group, gvr.Version, gvr.Resource, object.GetName(), object.GetNamespace())

	// Create the object that will be injected on templating system
	templateInjectedObject := map[string]interface{}{}
	templateInjectedObject["object"] = object.Object

	// Evaluate the conditions for targeted object
	var conditionFlags []bool

	for _, condition := range conditionList {

		parsedKey, err := template.EvaluateTemplate(condition.Key, templateInjectedObject)
		if err != nil {
			return false, fmt.Errorf("error evaluating condition template: %s", err)
		}

		conditionFlags = append(conditionFlags, parsedKey == condition.Value)

		globals.ExecContext.Logger.Debugf("condition: key: '%s', value: '%s', equals: '%t'",
			parsedKey, condition.Value, parsedKey == condition.Value)
	}

	// Conditions not met. Skip
	if slices.Contains(conditionFlags, false) {
		return false, nil
	}

	if globals.ExecContext.DryRun {
		globals.ExecContext.Logger.Infof("dry-run enabled. Skipping deletion of object: '%s'/'%s'",
			object.GetNamespace(), object.GetName())
		return false, nil
	}

	// Finally, delete the object
	err = p.Client.Resource(gvr).Namespace(object.GetNamespace()).
		Delete(globals.ExecContext.Context, object.GetName(), v1.DeleteOptions{})
	if err != nil {
		return false, fmt.Errorf("error deleting object: %s", err)
	}

	return true, nil
}
