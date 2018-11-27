package framework

import "time"

// A VMManifest contains a name and a manifest of
// a virtual machine or virtual machine instance.
type VMManifest struct {
	Name     string
	Manifest string
}

const (
	UsernameTestUser     = "kubevirt-test-user"
	NamespaceTestDefault = "kubevirt-test-default"
	UsernameAdminUser    = "test_admin"
	PasswordAdminUser    = "123456"

	CDI_LABEL_KEY      = "app"
	CDI_LABEL_VALUE    = "containerized-data-importer"
	CDI_LABEL_SELECTOR = CDI_LABEL_KEY + "=" + CDI_LABEL_VALUE
	paramFlag          = "-p"

	ShortTimeout = time.Duration(2) * time.Minute
	LongTimeout  = time.Duration(4) * time.Minute
)
