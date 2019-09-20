package masonry

import (
	"os"
	"path/filepath"

	"github.com/opencontrol/compliance-masonry/pkg/cli/get/resources"
	"github.com/opencontrol/compliance-masonry/pkg/lib/common"
	"github.com/opencontrol/compliance-masonry/pkg/lib/opencontrol"
	"github.com/opencontrol/compliance-masonry/pkg/lib/opencontrol/versions/1.0.0"
)

// BuildCache of opencontrols as defined in opencontrol.yaml
func BuildCache() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	destination := filepath.Join(wd, ".masonry_cache")

	repo := make([]common.RemoteSource, 1)
	repo[0] = schema.VCSEntry{
		URL:      "https://github.com/isimluk/ComplianceAsCode_redhat",
		Revision: "opencontrol-components",
		Path:     ""}
	getter := resources.NewVCSAndLocalGetter(opencontrol.YAMLParser{})
	return getter.GetRemoteResources(destination, "opencontrols", repo)
}