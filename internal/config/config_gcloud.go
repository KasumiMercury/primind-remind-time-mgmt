//go:build gcloud

package config

import "errors"

func (c *PubSubConfig) Validate() error {
	if c.GCloudProjectID == "" {
		return errors.New("GCLOUD_PROJECT_ID is required for event publishing")
	}
	return nil
}
