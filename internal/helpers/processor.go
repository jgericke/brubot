/*
   Any functions that are used repeatedly - i.e. to provide processing/cleanups for
   team names should end up here...
*/

package helpers

import "strings"

// CleanName simply lowers case and removes articles at the moment.
func CleanName(name string) string {
	return strings.Replace(strings.ToLower(name), "the ", "", -1)
}
