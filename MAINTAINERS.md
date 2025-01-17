# How to Release
## Comparing branches or tags

Execute branch.py to compare branches or tags.
branch.py has 2 positional arguments:
```
  from_ref           the ref to show the path from
  to_ref             the ref to show the path to
```
You can find more info by executing `branch.py -h`

Example:
`python branch.py v1.2 master` - this line compares the master branch with the release branch v1.2
The output of this command will be as follows:
```
[- 5e000f4b] release v1.2
[+ c000872d] satellite/payments: coupon value feature
...
[+ ec008dcf] build: Go 1.14.4
```
`-` sign means that commit not in the master branch(`from_ref`)
`+` sign means that commit not in v1.2 branch(`to_ref`)

From the above it follows that commits with a plus sign will be included in the next release v1.3

## How to write github changelog

the next step is to create the page on Confluence with our changelog for the release v1.3.
Example: [Release v1.31](https://storjlabs.atlassian.net/wiki/spaces/ENG/pages/1812791357/Release%2Bv1.31)
Here we need to highlight important changes for each topic(storj-sim, Uplink, Sattelite, Storage Node, General etc.)

## Cutting release branch

Then its time to cut the release branch:
`git checkout -b v1.3` - will create and checkout branch v1.3
`git push origin v1.3`- will push release branch to the repo

The next step is to create tag using `tag-release.sh` which is in storj/scripts folder and push it.
Example:
`./scripts/tag-release.sh v1.3.0-rc`
`git push origin v1.3.0-rc`
Then verify that the Jenkins job of the build Storj V3 for such tag and branch has finished successfully.
Together with that you need to create the same release branch for [tardigrade branding repo](https://github.com/storj/tardigrade-satellite-theme).

## How to cherry pick

If you need to cherry-pick something after the release branch has been created then you need to create point release.
Make sure that you have the latest changes, checkout the release branch and execute cherry-pick:
`git cherry-pick <your commit hash>`
You need to create pull request to the release branch with that commit. After the pull request will be approved and merged you should create new release tag:
`./scripts/tag-release.sh v1.3.1`
and push the tag to the repo:
`git push origin v1.3.1`
Verify that the Jenkins job of the build Storj V3 for such tag has finished successfully.
Double check that there was no change on the tardigrade branding in the meantime. Otherwise, the point release might get a broken tardigrade branding. If there are additional commits on the tardigrade branding it is better to create a release branch and revert them. You also have to update the Jenkins job to build from the release branch. Do not forget to change it back to master for the following regular release.
```
git clone git@github.com:storj/tardigrade-satellite-theme
git reset --hard <your commit hash>
git checkout -b release-v1.3
git push origin release-v1.3
```
Update Jenkins job.

## Where to find the release binaries and upload them to github

The release binaries you can find [here](https://storj-v3-alpha-builds.storage.googleapis.com/index.html).
Use search to find current release and open it.
Here you will need to download these binaries
```
"identity_darwin_amd64.zip", "identity_freebsd_amd64.zip", "identity_linux_amd64.zip", "identity_linux_arm.zip", "identity_linux_arm64.zip",
"identity_windows_amd64.zip", "storagenode-updater_linux_amd64.zip", "storagenode-updater_linux_arm.zip", "storagenode-updater_linux_arm64.zip",
"storagenode-updater_windows_amd64.zip", "storagenode_freebsd_amd64.zip", "storagenode_linux_amd64.zip", "storagenode_linux_arm.zip", "storagenode_linux_arm64.zip",
"storagenode_windows_amd64.msi.zip", "storagenode_windows_amd64.zip", "uplink_darwin_amd64.zip", "uplink_freebsd_amd64.zip", "uplink_linux_amd64.zip",
"uplink_linux_arm.zip", "uplink_linux_arm64.zip", "uplink_windows_amd64.zip", "uplinkng_darwin_amd64.zip", "uplinkng_freebsd_amd64.zip", "uplinkng_linux_amd64.zip",
"uplinkng_linux_arm.zip", "uplinkng_linux_arm64.zip", "uplinkng_windows_amd64.zip"
```
Then you will need to go to the storj repo on github and open  [releases](https://github.com/storj/storj/releases).
Press button [draft new release](https://github.com/storj/storj/releases/new), select the tag and upload binaries.
Also you will need to post here a changelog from the Confluence page, and press the button `Save draft`. You can look at examples from previous releases on GitHub.

## Which tests do we want to execute
Everything that could break production.
From the perspective of a storage node operator the storage node needs to run stable, should not get disqualified or suspended, payout and usage data should be available on the dashboard, graceful exit, garbage collection. Everything that touches on of these topics is most likely worth a test.
From the perspective of a satellite operator the durability, availability and accounting is important. This includes the audit and repair system. Also customer and storage node signups should work. Anything in these area is worth a test.
For an uplink the network just needs to work. Node selection is critical. If we keep selecting full or bad nodes the uplink will have a hard time to upload enough pieces. It also affects performance.

## Forum post changelog
On the forum you need to highlight the most important changes for the storage node operators and describe it a little.

## Useful links
[Deployment documentation](https://storjlabs.atlassian.net/wiki/spaces/OPS/pages/153190401/Satellite+-+post+phoenix#satellite.qa.storj.io)
[Storage node rollout process](https://storjlabs.atlassian.net/wiki/spaces/OPS/pages/138084357/Storagenode)
