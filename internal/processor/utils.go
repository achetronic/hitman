package processor

import "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

// deleteItemByIndex replaces the selected object with the last one from the list,
// then replace the whole list with it, minus the last.
func deleteItemByIndex(itemList *[]unstructured.Unstructured, index int) {
	*itemList = append((*itemList)[:index], (*itemList)[index+1:]...)
}
