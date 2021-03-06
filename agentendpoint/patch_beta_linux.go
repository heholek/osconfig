//  Copyright 2018 Google Inc. All Rights Reserved.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package agentendpoint

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/GoogleCloudPlatform/osconfig/ospatch"
	"github.com/GoogleCloudPlatform/osconfig/packages"

	agentendpointpb "google.golang.org/genproto/googleapis/cloud/osconfig/agentendpoint/v1beta"
)

func (r *patchTaskBeta) runUpdates(ctx context.Context) error {
	var errs []string
	const retryPeriod = 3 * time.Minute
	// Check for both apt-get and dpkg-query to give us a clean signal.
	if packages.AptExists && packages.DpkgQueryExists {
		opts := []ospatch.AptGetUpgradeOption{
			ospatch.AptGetDryRun(r.Task.GetDryRun()),
		}
		switch r.Task.GetPatchConfig().GetApt().GetType() {
		case agentendpointpb.AptSettings_DIST:
			opts = append(opts, ospatch.AptGetUpgradeType(packages.AptGetDistUpgrade))
		}
		r.debugf("Installing APT package updates.")
		if err := retryFunc(retryPeriod, "installing APT package updates", func() error { return ospatch.RunAptGetUpgrade(opts...) }); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if packages.YumExists && packages.RPMQueryExists {
		opts := []ospatch.YumUpdateOption{
			ospatch.YumUpdateSecurity(r.Task.GetPatchConfig().GetYum().GetSecurity()),
			ospatch.YumUpdateMinimal(r.Task.GetPatchConfig().GetYum().GetMinimal()),
			ospatch.YumUpdateExcludes(r.Task.GetPatchConfig().GetYum().GetExcludes()),
			ospatch.YumDryRun(r.Task.GetDryRun()),
		}
		r.debugf("Installing YUM package updates.")
		if err := retryFunc(retryPeriod, "installing YUM package updates", func() error { return ospatch.RunYumUpdate(opts...) }); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if packages.ZypperExists && packages.RPMQueryExists {
		opts := []ospatch.ZypperPatchOption{
			ospatch.ZypperPatchCategories(r.Task.GetPatchConfig().GetZypper().GetCategories()),
			ospatch.ZypperPatchSeverities(r.Task.GetPatchConfig().GetZypper().GetSeverities()),
			ospatch.ZypperUpdateWithUpdate(r.Task.GetPatchConfig().GetZypper().GetWithUpdate()),
			ospatch.ZypperUpdateWithOptional(r.Task.GetPatchConfig().GetZypper().GetWithOptional()),
			ospatch.ZypperUpdateDryrun(r.Task.GetDryRun()),
		}
		r.debugf("Installing Zypper updates.")
		if err := retryFunc(retryPeriod, "installing Zypper updates", func() error { return ospatch.RunZypperPatch(opts...) }); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if errs == nil {
		return nil
	}
	return errors.New(strings.Join(errs, ",\n"))
}
