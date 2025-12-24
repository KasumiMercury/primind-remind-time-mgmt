//go:build !gcloud

package config

func (c *PubSubConfig) Validate() error {
	return nil
}
