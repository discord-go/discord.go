# Release verification

Before publishing a GitHub release, push any required nested DAVE, MLS and
GoDave module tags. Next push the parent `v...` tag. Run the **Release consumer
check** workflow for that tag and wait for it to pass before publishing the
GitHub release. The tag must be publicly fetchable for the check to work.

The workflow uses Go 1.22 and runs `scripts/verify-release.sh VERSION` in a
fresh temporary module with `GOWORK=off`. It imports the bot, voice and
voiceplayer packages without local replacements. It downloads from the Git
repository directly to avoid cached negative proxy results after a new tag.
The version must be a valid published tag, not a branch or local checkout.

A successful build proves that the chosen package imports resolve and compile;
it does not replace full tests, a voice interoperability test, or checking that
public module proxies have refreshed. See `dave-go/PORTING.md` for the nested
module release procedure.
