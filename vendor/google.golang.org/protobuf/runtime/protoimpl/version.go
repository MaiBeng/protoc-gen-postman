// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package protoimpl

import (
	"fmt"
	"strings"
)

// These constants determine the current version of this module.
//
//
// For our release process, we enforce the following rules:
//	* Tagged releases use a tag that is identical to VersionString.
//	* Tagged releases never reference a commit where the VersionString
//	contains "devel".
//	* The set of all commits in this repository where VersionString
//	does not contain "devel" must have a unique VersionString.
//
//
// Steps for tagging a new release:
//	1. Create a new CL.
//
//	2. Update versionMinor, versionPatch, and/or versionPreRelease as necessary.
//	versionPreRelease must not contain the string "devel".
//
//	3. Since the last released minor version, have there been any changes to
//	generator that relies on new functionality in the runtime?
//	If yes, then increment GenVersion.
//
//	4. Since the last released minor version, have there been any changes to
//	the runtime that removes support for old .pb.go source code?
//	If yes, then increment MinVersion.
//
//	5. Send out the CL for review and submit it.
//	Note that the next CL in step 8 must be submitted after this CL
//	without any other CLs in-between.
//
//	6. Tag a new version, where the tag is is the current VersionString.
//
//	7. Write release notes for all notable changes
//	between this release and the last release.
//
//	8. Create a new CL.
//
//	9. Update versionPreRelease to include the string "devel".
//	For example: "" -> "devel" or "rc.1" -> "rc.1.devel"
//
//	10. Send out the CL for review and submit it.
const (
	versionMajor      = 1
	versionMinor      = 20
	versionPatch      = 0
	versionPreRelease = ""
)

// VersionString formats the version string for this module in semver format.
//
// Examples:
//	v1.20.1
//	v1.21.0-rc.1
func VersionString() string {
	v := fmt.Sprintf("v%d.%d.%d", versionMajor, versionMinor, versionPatch)
	if versionPreRelease != "" {
		v += "-" + versionPreRelease

		// TODO: Add metadata about the commit or build hash.
		// See https://golang.org/issue/29814
		// See https://golang.org/issue/33533
		var versionMetadata string
		if strings.Contains(versionPreRelease, "devel") && versionMetadata != "" {
			v += "+" + versionMetadata
		}
	}
	return v
}

const (
	// MaxVersion is the maximum supported version for generated .pb.go files.
	// It is always the current version of the module.
	MaxVersion = versionMinor

	// GenVersion is the runtime version required by generated .pb.go files.
	// This is incremented when generated code relies on new functionality
	// in the runtime.
	GenVersion = 20

	// MinVersion is the minimum supported version for generated .pb.go files.
	// This is incremented when the runtime drops support for old code.
	MinVersion = 0
)

// EnforceVersion is used by code generated by protoc-gen-go
// to statically enforce minimum and maximum versions of this package.
// A compilation failure implies either that:
//	* the runtime package is too old and needs to be updated OR
//	* the generated code is too old and needs to be regenerated.
//
// The runtime package can be upgraded by running:
//	go get google.golang.org/protobuf
//
// The generated code can be regenerated by running:
//	protoc --go_out=${PROTOC_GEN_GO_ARGS} ${PROTO_FILES}
//
// Example usage by generated code:
//	const (
//		// Verify that this generated code is sufficiently up-to-date.
//		_ = protoimpl.EnforceVersion(genVersion - protoimpl.MinVersion)
//		// Verify that runtime/protoimpl is sufficiently up-to-date.
//		_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - genVersion)
//	)
//
// The genVersion is the current minor version used to generated the code.
// This compile-time check relies on negative integer overflow of a uint
// being a compilation failure (guaranteed by the Go specification).
type EnforceVersion uint

// This enforces the following invariant:
//	MinVersion ≤ GenVersion ≤ MaxVersion
const (
	_ = EnforceVersion(GenVersion - MinVersion)
	_ = EnforceVersion(MaxVersion - GenVersion)
)
