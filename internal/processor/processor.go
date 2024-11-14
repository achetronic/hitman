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

		// Preprocess the targets list to clean the items not matching the user-desired criteria
		for rawResourceIndex, rawResourceObject := range resourceList.Items {

			// Matching namespace by regex and resource does NOT meet? Skip
			if configResource.Target.Namespace.MatchRegex != "" &&
				!compiledRegexNamespace.MatchString(rawResourceObject.GetNamespace()) {

				deleteItemByIndex(&resourceList.Items, rawResourceIndex)
				continue
			}

			// Matching name by exact string and resource does NOT meet? Skip
			if configResource.Target.Name.MatchExact != "" &&
				rawResourceObject.GetName() != configResource.Target.Name.MatchExact {

				deleteItemByIndex(&resourceList.Items, rawResourceIndex)
				continue
			}

			// Matching name by regex and resource does NOT meet? Skip
			if configResource.Target.Name.MatchRegex != "" &&
				!compiledRegex.MatchString(rawResourceObject.GetName()) {

				deleteItemByIndex(&resourceList.Items, rawResourceIndex)
				continue
			}
		}

		templateInjectedObject := &map[string]interface{}{} // TODO, review potential nil pointer dereference

		// Perform global user-defined actions when 'preStep' is set in the config
		// This is useful to group resources, pre-filter some of them, etc, before evaluating one by one
		if configResource.PreStep != "" {
			err = p.processPrestep(configResource.PreStep, templateInjectedObject, resourceList.Items)
			if err != nil {
				globals.ExecContext.Logger.Infof("error processing prestep: %s", err)
				continue
			}
		}

		// Perform the actions over the resources
		for _, resource := range resourceList.Items {

			// Process this object. Delete in case of success
			objectDeleted, err := p.processObject(gvr, resource, templateInjectedObject, configResource.Conditions)
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

// processPrestep process a list with all the user-desired targets
// It receive the .targets and is able to store variables inside .vars that are available into conditions' later evaluation
func (p *Processor) processPrestep(userTemplate string, templateInjectedData *map[string]interface{}, targetList []unstructured.Unstructured) (err error) {

	// Convert injected data into allowed type
	injectedTargetList := []map[string]interface{}{}
	for _, target := range targetList {
		injectedTargetList = append(injectedTargetList, target.Object)
	}

	//
	(*templateInjectedData)["targets"] = injectedTargetList

	_, err = template.EvaluateTemplate(userTemplate, templateInjectedData)
	if err != nil {
		return fmt.Errorf("error evaluating prestep template: %s", err.Error())
	}

	delete((*templateInjectedData), "targets")

	return nil
}

// processObject process an object coming from arguments.
// It computes templating, evaluates conditions and decides whether to delete it or not.
func (p *Processor) processObject(gvr schema.GroupVersionResource, object unstructured.Unstructured, templateInjectedData *map[string]interface{}, conditionList []v1alpha1.ConditionT) (result bool, err error) {

	globals.ExecContext.Logger.Debugf("processing object: group: '%s', version: '%s', resource: '%s', name: '%s', namespace: '%s'",
		gvr.Group, gvr.Version, gvr.Resource, object.GetName(), object.GetNamespace())

	// Create the object that will be injected on templating system
	(*templateInjectedData)["object"] = object.Object

	// Evaluate the conditions for targeted object
	var conditionFlags []bool

	for _, condition := range conditionList {

		parsedKey, err := template.EvaluateTemplate(condition.Key, templateInjectedData)
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

	// Define a grace period (in seconds) for the pod deletion
	// Ref: https://github.com/kubernetes/apimachinery/blob/master/pkg/apis/meta/v1/types.go#L507
	gracePeriodSeconds := int64(0) // 0 for immediate deletion

	// Finally, delete the object
	err = p.Client.Resource(gvr).Namespace(object.GetNamespace()).
		Delete(globals.ExecContext.Context, object.GetName(), v1.DeleteOptions{
			GracePeriodSeconds: &gracePeriodSeconds,
		})
	if err != nil {
		return false, fmt.Errorf("error deleting object: %s", err)
	}

	return true, nil
}
