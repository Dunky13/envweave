package releaseidentity

import (
	"strings"

	"github.com/Masterminds/semver/v3"
)

// CompareUpdateVersions orders nightlies after their stable base and its other
// prereleases: they may
// contain migrations added after that stable release. All other comparisons
// retain SemVer ordering. This is eligibility, not migration authorization.
func CompareUpdateVersions(a, b *semver.Version) int {
	if a.Major() == b.Major() && a.Minor() == b.Minor() && a.Patch() == b.Patch() {
		aNightly := strings.HasPrefix(a.Prerelease(), "nightly.")
		bNightly := strings.HasPrefix(b.Prerelease(), "nightly.")
		if aNightly && !bNightly {
			return 1
		}
		if !aNightly && bNightly {
			return -1
		}
	}
	return a.Compare(b)
}
