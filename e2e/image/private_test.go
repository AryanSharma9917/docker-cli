package image

import (
	"strings"
	"testing"

	"github.com/docker/cli/e2e/internal/fixtures"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/icmd"
)

const privateRegistryPrefix = "private-registry:5001"

// Regression test for https://github.com/docker/cli/issues/5963
func TestPullPushPrivateRepository(t *testing.T) {
	t.Parallel()

	dir := fixtures.SetupConfigFile(t)
	t.Cleanup(dir.Remove)
	emptyConfigDir := t.TempDir()

	sourceImage := fixtures.AlpineImage
	privateImage := privateRegistryPrefix + "/private/alpine:test-private-pull-push"

	icmd.RunCommand("docker", "pull", sourceImage).Assert(t, icmd.Success)
	t.Cleanup(func() {
		icmd.RunCommand("docker", "image", "rm", "-f", privateImage).Assert(t, icmd.Success)
	})

	icmd.RunCommand("docker", "tag", sourceImage, privateImage).Assert(t, icmd.Success)

	pushNoAuth := icmd.RunCmd(
		icmd.Command("docker", "push", privateImage),
		fixtures.WithConfig(emptyConfigDir),
	)
	pushNoAuth.Assert(t, icmd.Expected{ExitCode: 1})
	assertAuthDenied(t, pushNoAuth)

	pushWithAuth := icmd.RunCmd(
		icmd.Command("docker", "push", privateImage),
		fixtures.WithConfig(dir.Path()),
	)
	pushWithAuth.Assert(t, icmd.Success)
	assert.Check(t, strings.Contains(pushWithAuth.Combined(), "The push refers to repository ["+privateImage+"]"), pushWithAuth.Combined())

	icmd.RunCommand("docker", "image", "rm", "-f", privateImage).Assert(t, icmd.Success)

	pullNoAuth := icmd.RunCmd(
		icmd.Command("docker", "pull", privateImage),
		fixtures.WithConfig(emptyConfigDir),
	)
	pullNoAuth.Assert(t, icmd.Expected{ExitCode: 1})
	assertAuthDenied(t, pullNoAuth)

	pullWithAuth := icmd.RunCmd(
		icmd.Command("docker", "pull", privateImage),
		fixtures.WithConfig(dir.Path()),
	)
	pullWithAuth.Assert(t, icmd.Success)
	assert.Check(t, strings.Contains(pullWithAuth.Combined(), privateImage), pullWithAuth.Combined())
}

func assertAuthDenied(t *testing.T, result *icmd.Result) {
	t.Helper()
	output := result.Combined()

	assert.Check(t,
		strings.Contains(output, "requested access to the resource is denied") ||
			strings.Contains(output, "no basic auth credentials") ||
			strings.Contains(output, "unauthorized") ||
			strings.Contains(output, "authentication required"),
		output,
	)
}
