// Copyright 2020-2021 Buf Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package buf

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/bufbuild/buf/internal/buf/bufcli"
	"github.com/bufbuild/buf/internal/pkg/osextended"
	"github.com/bufbuild/buf/internal/pkg/storage/storagearchive"
	"github.com/bufbuild/buf/internal/pkg/storage/storageos"
	"github.com/stretchr/testify/require"
)

func TestWorkspaceDir(t *testing.T) {
	// Directory paths contained within a workspace.
	t.Parallel()
	wd, err := osextended.Getwd()
	require.NoError(t, err)
	testRunStdout(
		t,
		nil,
		0,
		``,
		"build",
		filepath.Join("testdata", "workspace", "success", "dir", "proto"),
	)
	testRunStdout(
		t,
		nil,
		0,
		`testdata/workspace/success/dir/proto/rpc.proto`,
		"ls-files",
		filepath.Join("testdata", "workspace", "success", "dir", "proto"),
	)
	testRunStdout(
		t,
		nil,
		bufcli.ExitCodeFileAnnotation,
		`testdata/workspace/success/dir/proto/rpc.proto:3:1:Files with package "example" must be within a directory "example" relative to root but were in directory ".".
        testdata/workspace/success/dir/proto/rpc.proto:3:1:Package name "example" should be suffixed with a correctly formed version, such as "example.v1".`,
		"lint",
		filepath.Join("testdata", "workspace", "success", "dir", "proto"),
	)
	testRunStdout(
		t,
		nil,
		0,
		``,
		"build",
		filepath.Join("testdata", "workspace", "success", "dir", "other", "proto"),
	)
	testRunStdout(
		t,
		nil,
		0,
		`testdata/workspace/success/dir/other/proto/request.proto`,
		"ls-files",
		filepath.Join("testdata", "workspace", "success", "dir", "other", "proto"),
	)
	testRunStdout(
		t,
		nil,
		bufcli.ExitCodeFileAnnotation,
		`testdata/workspace/success/dir/other/proto/request.proto:3:1:Files with package "request" must be within a directory "request" relative to root but were in directory ".".
        testdata/workspace/success/dir/other/proto/request.proto:3:1:Package name "request" should be suffixed with a correctly formed version, such as "request.v1".`,
		"lint",
		filepath.Join("testdata", "workspace", "success", "dir", "other", "proto"),
	)
	testRunStdout(
		t,
		nil,
		bufcli.ExitCodeFileAnnotation,
		fmt.Sprintf(`%s/testdata/workspace/success/dir/other/proto/request.proto:3:1:Files with package "request" must be within a directory "request" relative to root but were in directory ".".
        %s/testdata/workspace/success/dir/other/proto/request.proto:3:1:Package name "request" should be suffixed with a correctly formed version, such as "request.v1".`, wd, wd),
		"lint",
		filepath.Join(wd, "testdata", "workspace", "success", "dir", "other", "proto"),
	)
	testRunStdout(
		t,
		nil,
		0,
		``,
		"build",
		filepath.Join("testdata", "workspace", "success", "dir"),
	)
	testRunStdout(
		t,
		nil,
		0,
		`testdata/workspace/success/dir/other/proto/request.proto
        testdata/workspace/success/dir/proto/rpc.proto`,
		"ls-files",
		filepath.Join("testdata", "workspace", "success", "dir"),
	)
	testRunStdout(
		t,
		nil,
		bufcli.ExitCodeFileAnnotation,
		`testdata/workspace/success/dir/other/proto/request.proto:3:1:Files with package "request" must be within a directory "request" relative to root but were in directory ".".
        testdata/workspace/success/dir/other/proto/request.proto:3:1:Package name "request" should be suffixed with a correctly formed version, such as "request.v1".
        testdata/workspace/success/dir/proto/rpc.proto:3:1:Files with package "example" must be within a directory "example" relative to root but were in directory ".".
        testdata/workspace/success/dir/proto/rpc.proto:3:1:Package name "example" should be suffixed with a correctly formed version, such as "example.v1".`,
		"lint",
		filepath.Join("testdata", "workspace", "success", "dir"),
	)
	testRunStdout(
		t,
		nil,
		0,
		``,
		"build",
		filepath.Join("testdata", "workspace", "success", "breaking"),
	)
	testRunStdout(
		t,
		nil,
		0,
		`testdata/workspace/success/breaking/other/proto/request.proto
        testdata/workspace/success/breaking/proto/rpc.proto`,
		"ls-files",
		filepath.Join("testdata", "workspace", "success", "breaking"),
	)
	testRunStdout(
		t,
		nil,
		bufcli.ExitCodeFileAnnotation,
		`testdata/workspace/success/breaking/other/proto/request.proto:5:1:Previously present field "1" with name "name" on message "Request" was deleted.
        testdata/workspace/success/breaking/proto/rpc.proto:8:5:Field "1" with name "request" on message "RPC" changed option "json_name" from "req" to "request".
        testdata/workspace/success/breaking/proto/rpc.proto:8:21:Field "1" on message "RPC" changed name from "req" to "request".`,
		"breaking",
		filepath.Join("testdata", "workspace", "success", "breaking"),
		"--against",
		filepath.Join("testdata", "workspace", "success", "dir"),
	)
}

func TestWorkspaceArchiveDir(t *testing.T) {
	// Archive that defines a workspace at the root of the archive.
	t.Parallel()
	zipDir := createZipFromDir(
		t,
		filepath.Join("testdata", "workspace", "success", "dir"),
		"archive.zip",
	)
	testRunStdout(
		t,
		nil,
		0,
		``,
		"build",
		filepath.Join(zipDir, "archive.zip#subdir=proto"),
	)
	testRunStdout(
		t,
		nil,
		0,
		`proto/rpc.proto`,
		"ls-files",
		filepath.Join(zipDir, "archive.zip#subdir=proto"),
	)
	testRunStdout(
		t,
		nil,
		bufcli.ExitCodeFileAnnotation,
		`proto/rpc.proto:3:1:Files with package "example" must be within a directory "example" relative to root but were in directory ".".
        proto/rpc.proto:3:1:Package name "example" should be suffixed with a correctly formed version, such as "example.v1".`,
		"lint",
		filepath.Join(zipDir, "archive.zip#subdir=proto"),
	)
}

func TestWorkspaceNestedArchive(t *testing.T) {
	// Archive that defines a workspace in a sub-directory to the root.
	t.Parallel()
	zipDir := createZipFromDir(
		t,
		filepath.Join("testdata", "workspace", "success", "nested"),
		"archive.zip",
	)
	testRunStdout(
		t,
		nil,
		0,
		``,
		"build",
		filepath.Join(zipDir, "archive.zip#subdir=proto/internal"),
	)
	testRunStdout(
		t,
		nil,
		0,
		`proto/internal/internal.proto`,
		"ls-files",
		filepath.Join(zipDir, "archive.zip#subdir=proto/internal"),
	)
	testRunStdout(
		t,
		nil,
		bufcli.ExitCodeFileAnnotation,
		`proto/internal/internal.proto:3:1:Files with package "internal" must be within a directory "internal" relative to root but were in directory ".".
        proto/internal/internal.proto:3:1:Package name "internal" should be suffixed with a correctly formed version, such as "internal.v1".`,
		"lint",
		filepath.Join(zipDir, "archive.zip#subdir=proto/internal"),
	)
}

func TestWorkspaceGit(t *testing.T) {
	// Directory paths specified as a git reference within a workspace.
	t.Parallel()
	testRunStdout(
		t,
		nil,
		0,
		``,
		"build",
		filepath.Join("testdata", "workspace", "success", "../../../../../../../.git#ref=HEAD,subdir=internal/buf/cmd/buf/testdata/workspace/success/dir/proto"),
	)
	testRunStdout(
		t,
		nil,
		0,
		`internal/buf/cmd/buf/testdata/workspace/success/dir/proto/rpc.proto`,
		"ls-files",
		filepath.Join("testdata", "workspace", "success", "../../../../../../../.git#ref=HEAD,subdir=internal/buf/cmd/buf/testdata/workspace/success/dir/proto"),
	)
	testRunStdout(
		t,
		nil,
		bufcli.ExitCodeFileAnnotation,
		`internal/buf/cmd/buf/testdata/workspace/success/dir/proto/rpc.proto:3:1:Files with package "example" must be within a directory "example" relative to root but were in directory ".".
        internal/buf/cmd/buf/testdata/workspace/success/dir/proto/rpc.proto:3:1:Package name "example" should be suffixed with a correctly formed version, such as "example.v1".`,
		"lint",
		filepath.Join("testdata", "workspace", "success", "../../../../../../../.git#ref=HEAD,subdir=internal/buf/cmd/buf/testdata/workspace/success/dir/proto"),
	)
}

func TestWorkspaceDetached(t *testing.T) {
	// The workspace doesn't include the 'proto' directory, so
	// its contents aren't included in the workspace.
	t.Parallel()
	testRunStdout(
		t,
		nil,
		0,
		``,
		"build",
		filepath.Join("testdata", "workspace", "success", "detached", "proto"),
	)
	testRunStdout(
		t,
		nil,
		0,
		`testdata/workspace/success/detached/proto/rpc.proto`,
		"ls-files",
		filepath.Join("testdata", "workspace", "success", "detached", "proto"),
	)
	testRunStdout(
		t,
		nil,
		bufcli.ExitCodeFileAnnotation,
		`testdata/workspace/success/detached/proto/rpc.proto:3:1:Files with package "example" must be within a directory "example" relative to root but were in directory ".".
        testdata/workspace/success/detached/proto/rpc.proto:3:1:Package name "example" should be suffixed with a correctly formed version, such as "example.v1".`,
		"lint",
		filepath.Join("testdata", "workspace", "success", "detached", "proto"),
	)
	testRunStdout(
		t,
		nil,
		0,
		``,
		"build",
		filepath.Join("testdata", "workspace", "success", "detached", "other", "proto"),
	)
	testRunStdout(
		t,
		nil,
		0,
		`testdata/workspace/success/detached/other/proto/request.proto`,
		"ls-files",
		filepath.Join("testdata", "workspace", "success", "detached", "other", "proto"),
	)
	testRunStdout(
		t,
		nil,
		bufcli.ExitCodeFileAnnotation,
		`testdata/workspace/success/detached/other/proto/request.proto:3:1:Files with package "request" must be within a directory "request" relative to root but were in directory ".".
        testdata/workspace/success/detached/other/proto/request.proto:3:1:Package name "request" should be suffixed with a correctly formed version, such as "request.v1".`,
		"lint",
		filepath.Join("testdata", "workspace", "success", "detached", "other", "proto"),
	)
	testRunStdout(
		t,
		nil,
		0,
		``,
		"build",
		filepath.Join("testdata", "workspace", "success", "detached"),
	)
	testRunStdout(
		t,
		nil,
		0,
		`testdata/workspace/success/detached/other/proto/request.proto`,
		"ls-files",
		filepath.Join("testdata", "workspace", "success", "detached"),
	)
	testRunStdout(
		t,
		nil,
		bufcli.ExitCodeFileAnnotation,
		`testdata/workspace/success/detached/other/proto/request.proto:3:1:Files with package "request" must be within a directory "request" relative to root but were in directory ".".
        testdata/workspace/success/detached/other/proto/request.proto:3:1:Package name "request" should be suffixed with a correctly formed version, such as "request.v1".`,
		"lint",
		filepath.Join("testdata", "workspace", "success", "detached"),
	)
}

func TestWorkspaceNoModuleConfig(t *testing.T) {
	// The workspace points to modules that don't contain a buf.yaml.
	t.Parallel()
	testRunStdout(
		t,
		nil,
		0,
		``,
		"build",
		filepath.Join("testdata", "workspace", "success", "noconfig", "proto"),
	)
	testRunStdout(
		t,
		nil,
		0,
		`testdata/workspace/success/noconfig/proto/rpc.proto`,
		"ls-files",
		filepath.Join("testdata", "workspace", "success", "noconfig", "proto"),
	)
	testRunStdout(
		t,
		nil,
		bufcli.ExitCodeFileAnnotation,
		`testdata/workspace/success/noconfig/proto/rpc.proto:3:1:Files with package "example" must be within a directory "example" relative to root but were in directory ".".
        testdata/workspace/success/noconfig/proto/rpc.proto:3:1:Package name "example" should be suffixed with a correctly formed version, such as "example.v1".`,
		"lint",
		filepath.Join("testdata", "workspace", "success", "noconfig", "proto"),
	)
	testRunStdout(
		t,
		nil,
		0,
		``,
		"build",
		filepath.Join("testdata", "workspace", "success", "noconfig", "other", "proto"),
	)
	testRunStdout(
		t,
		nil,
		0,
		`testdata/workspace/success/noconfig/other/proto/request.proto`,
		"ls-files",
		filepath.Join("testdata", "workspace", "success", "noconfig", "other", "proto"),
	)
	testRunStdout(
		t,
		nil,
		bufcli.ExitCodeFileAnnotation,
		`testdata/workspace/success/noconfig/other/proto/request.proto:3:1:Files with package "request" must be within a directory "request" relative to root but were in directory ".".
        testdata/workspace/success/noconfig/other/proto/request.proto:3:1:Package name "request" should be suffixed with a correctly formed version, such as "request.v1".`,
		"lint",
		filepath.Join("testdata", "workspace", "success", "noconfig", "other", "proto"),
	)
}

func TestWorkspaceWithLock(t *testing.T) {
	// The workspace points to a module that includes a buf.lock, but
	// the listed dependency is defined in the workspace so the module
	// cache is unused.
	t.Parallel()
	testRunStdout(
		t,
		nil,
		0,
		``,
		"build",
		filepath.Join("testdata", "workspace", "success", "lock", "a"),
	)
	testRunStdout(
		t,
		nil,
		0,
		`testdata/workspace/success/lock/a/a.proto`,
		"ls-files",
		filepath.Join("testdata", "workspace", "success", "lock", "a"),
	)
	testRunStdout(
		t,
		nil,
		bufcli.ExitCodeFileAnnotation,
		`testdata/workspace/success/lock/a/a.proto:3:1:Files with package "a" must be within a directory "a" relative to root but were in directory ".".
        testdata/workspace/success/lock/a/a.proto:3:1:Package name "a" should be suffixed with a correctly formed version, such as "a.v1".`,
		"lint",
		filepath.Join("testdata", "workspace", "success", "lock", "a"),
	)
	testRunStdout(
		t,
		nil,
		0,
		``,
		"build",
		filepath.Join("testdata", "workspace", "success", "lock", "b"),
	)
	testRunStdout(
		t,
		nil,
		0,
		`testdata/workspace/success/lock/b/b.proto`,
		"ls-files",
		filepath.Join("testdata", "workspace", "success", "lock", "b"),
	)
	testRunStdout(
		t,
		nil,
		bufcli.ExitCodeFileAnnotation,
		`testdata/workspace/success/lock/b/b.proto:3:1:Files with package "b" must be within a directory "b" relative to root but were in directory ".".
        testdata/workspace/success/lock/b/b.proto:3:1:Package name "b" should be suffixed with a correctly formed version, such as "b.v1".`,
		"lint",
		filepath.Join("testdata", "workspace", "success", "lock", "b"),
	)
}

func TestWorkspaceWithTransitiveDependencies(t *testing.T) {
	// The workspace points to a module that includes transitive
	// dependencies (i.e. a depends on b, and b depends on c).
	t.Parallel()
	testRunStdout(
		t,
		nil,
		0,
		``,
		"build",
		filepath.Join("testdata", "workspace", "success", "transitive", "proto"),
	)
	testRunStdout(
		t,
		nil,
		0,
		`testdata/workspace/success/transitive/proto/a.proto`,
		"ls-files",
		filepath.Join("testdata", "workspace", "success", "transitive", "proto"),
	)
	testRunStdout(
		t,
		nil,
		bufcli.ExitCodeFileAnnotation,
		`testdata/workspace/success/transitive/proto/a.proto:3:1:Files with package "a" must be within a directory "a" relative to root but were in directory ".".
        testdata/workspace/success/transitive/proto/a.proto:3:1:Package name "a" should be suffixed with a correctly formed version, such as "a.v1".`,
		"lint",
		filepath.Join("testdata", "workspace", "success", "transitive", "proto"),
	)
	testRunStdout(
		t,
		nil,
		0,
		``,
		"build",
		filepath.Join("testdata", "workspace", "success", "transitive", "private", "proto"),
	)
	testRunStdout(
		t,
		nil,
		0,
		`testdata/workspace/success/transitive/private/proto/b.proto`,
		"ls-files",
		filepath.Join("testdata", "workspace", "success", "transitive", "private", "proto"),
	)
	testRunStdout(
		t,
		nil,
		bufcli.ExitCodeFileAnnotation,
		`testdata/workspace/success/transitive/private/proto/b.proto:3:1:Files with package "b" must be within a directory "b" relative to root but were in directory ".".
        testdata/workspace/success/transitive/private/proto/b.proto:3:1:Package name "b" should be suffixed with a correctly formed version, such as "b.v1".`,
		"lint",
		filepath.Join("testdata", "workspace", "success", "transitive", "private", "proto"),
	)
	testRunStdout(
		t,
		nil,
		0,
		``,
		"build",
		filepath.Join("testdata", "workspace", "success", "transitive", "other", "proto"),
	)
	testRunStdout(
		t,
		nil,
		0,
		`testdata/workspace/success/transitive/other/proto/c.proto`,
		"ls-files",
		filepath.Join("testdata", "workspace", "success", "transitive", "other", "proto"),
	)
	testRunStdout(
		t,
		nil,
		bufcli.ExitCodeFileAnnotation,
		`testdata/workspace/success/transitive/other/proto/c.proto:3:1:Files with package "c" must be within a directory "c" relative to root but were in directory ".".
        testdata/workspace/success/transitive/other/proto/c.proto:3:1:Package name "c" should be suffixed with a correctly formed version, such as "c.v1".`,
		"lint",
		filepath.Join("testdata", "workspace", "success", "transitive", "other", "proto"),
	)
	testRunStdout(
		t,
		nil,
		0,
		``,
		"build",
		filepath.Join("testdata", "workspace", "success", "transitive"),
	)
	testRunStdout(
		t,
		nil,
		0,
		`testdata/workspace/success/transitive/other/proto/c.proto
        testdata/workspace/success/transitive/private/proto/b.proto
        testdata/workspace/success/transitive/proto/a.proto`,
		"ls-files",
		filepath.Join("testdata", "workspace", "success", "transitive"),
	)
	testRunStdout(
		t,
		nil,
		bufcli.ExitCodeFileAnnotation,
		`testdata/workspace/success/transitive/other/proto/c.proto:3:1:Files with package "c" must be within a directory "c" relative to root but were in directory ".".
        testdata/workspace/success/transitive/other/proto/c.proto:3:1:Package name "c" should be suffixed with a correctly formed version, such as "c.v1".
        testdata/workspace/success/transitive/private/proto/b.proto:3:1:Files with package "b" must be within a directory "b" relative to root but were in directory ".".
        testdata/workspace/success/transitive/private/proto/b.proto:3:1:Package name "b" should be suffixed with a correctly formed version, such as "b.v1".
        testdata/workspace/success/transitive/proto/a.proto:3:1:Files with package "a" must be within a directory "a" relative to root but were in directory ".".
        testdata/workspace/success/transitive/proto/a.proto:3:1:Package name "a" should be suffixed with a correctly formed version, such as "a.v1".`,
		"lint",
		filepath.Join("testdata", "workspace", "success", "transitive"),
	)
}

func TestWorkspaceWithDiamondDependency(t *testing.T) {
	// The workspace points to a module that includes a diamond
	// dependency (i.e. a depends on b and c, and b depends on c).
	t.Parallel()
	testRunStdout(
		t,
		nil,
		0,
		``,
		"build",
		filepath.Join("testdata", "workspace", "success", "diamond", "proto"),
	)
	testRunStdout(
		t,
		nil,
		0,
		`testdata/workspace/success/diamond/proto/a.proto`,
		"ls-files",
		filepath.Join("testdata", "workspace", "success", "diamond", "proto"),
	)
	testRunStdout(
		t,
		nil,
		bufcli.ExitCodeFileAnnotation,
		`testdata/workspace/success/diamond/proto/a.proto:3:1:Files with package "a" must be within a directory "a" relative to root but were in directory ".".
        testdata/workspace/success/diamond/proto/a.proto:3:1:Package name "a" should be suffixed with a correctly formed version, such as "a.v1".`,
		"lint",
		filepath.Join("testdata", "workspace", "success", "diamond", "proto"),
	)
	testRunStdout(
		t,
		nil,
		0,
		``,
		"build",
		filepath.Join("testdata", "workspace", "success", "diamond", "private", "proto"),
	)
	testRunStdout(
		t,
		nil,
		0,
		`testdata/workspace/success/diamond/private/proto/b.proto`,
		"ls-files",
		filepath.Join("testdata", "workspace", "success", "diamond", "private", "proto"),
	)
	testRunStdout(
		t,
		nil,
		bufcli.ExitCodeFileAnnotation,
		`testdata/workspace/success/diamond/private/proto/b.proto:3:1:Files with package "b" must be within a directory "b" relative to root but were in directory ".".
        testdata/workspace/success/diamond/private/proto/b.proto:3:1:Package name "b" should be suffixed with a correctly formed version, such as "b.v1".`,
		"lint",
		filepath.Join("testdata", "workspace", "success", "diamond", "private", "proto"),
	)
	testRunStdout(
		t,
		nil,
		0,
		``,
		"build",
		filepath.Join("testdata", "workspace", "success", "diamond", "other", "proto"),
	)
	testRunStdout(
		t,
		nil,
		0,
		`testdata/workspace/success/diamond/other/proto/c.proto`,
		"ls-files",
		filepath.Join("testdata", "workspace", "success", "diamond", "other", "proto"),
	)
	testRunStdout(
		t,
		nil,
		bufcli.ExitCodeFileAnnotation,
		`testdata/workspace/success/diamond/other/proto/c.proto:3:1:Files with package "c" must be within a directory "c" relative to root but were in directory ".".
        testdata/workspace/success/diamond/other/proto/c.proto:3:1:Package name "c" should be suffixed with a correctly formed version, such as "c.v1".`,
		"lint",
		filepath.Join("testdata", "workspace", "success", "diamond", "other", "proto"),
	)
}

func TestWorkspaceSymlink(t *testing.T) {
	// The workspace includes valid symlinks.
	t.Parallel()
	testRunStdout(
		t,
		nil,
		0,
		``,
		"build",
		filepath.Join("testdata", "workspace", "success", "symlink"),
	)
	testRunStdout(
		t,
		nil,
		0,
		`testdata/workspace/success/symlink/a/a.proto
        testdata/workspace/success/symlink/b/b.proto
        testdata/workspace/success/symlink/c/c.proto`,
		"ls-files",
		filepath.Join("testdata", "workspace", "success", "symlink"),
	)
	testRunStdout(
		t,
		nil,
		bufcli.ExitCodeFileAnnotation,
		`testdata/workspace/success/symlink/a/a.proto:3:1:Files with package "a" must be within a directory "a" relative to root but were in directory ".".
        testdata/workspace/success/symlink/a/a.proto:3:1:Package name "a" should be suffixed with a correctly formed version, such as "a.v1".
        testdata/workspace/success/symlink/b/b.proto:3:1:Files with package "b" must be within a directory "b" relative to root but were in directory ".".
        testdata/workspace/success/symlink/b/b.proto:3:1:Package name "b" should be suffixed with a correctly formed version, such as "b.v1".
        testdata/workspace/success/symlink/c/c.proto:3:1:Files with package "c" must be within a directory "c" relative to root but were in directory ".".
        testdata/workspace/success/symlink/c/c.proto:3:1:Package name "c" should be suffixed with a correctly formed version, such as "c.v1".`,
		"lint",
		filepath.Join("testdata", "workspace", "success", "symlink"),
	)
}

func TestWorkspaceWKT(t *testing.T) {
	// The workspace includes multiple images that import the same
	// well-known type (empty.proto).
	t.Parallel()
	testRunStdout(
		t,
		nil,
		0,
		``,
		"build",
		filepath.Join("testdata", "workspace", "success", "wkt", "other", "proto"),
	)
	testRunStdout(
		t,
		nil,
		0,
		`testdata/workspace/success/wkt/other/proto/c/c.proto`,
		"ls-files",
		filepath.Join("testdata", "workspace", "success", "wkt", "other", "proto"),
	)
	testRunStdout(
		t,
		nil,
		bufcli.ExitCodeFileAnnotation,
		`testdata/workspace/success/wkt/other/proto/c/c.proto:6:1:Package name "c" should be suffixed with a correctly formed version, such as "c.v1".`,
		"lint",
		filepath.Join("testdata", "workspace", "success", "wkt", "other", "proto"),
	)
	testRunStdout(
		t,
		nil,
		0,
		``,
		"build",
		filepath.Join("testdata", "workspace", "success", "wkt"),
	)
	testRunStdout(
		t,
		nil,
		0,
		`testdata/workspace/success/wkt/other/proto/c/c.proto
        testdata/workspace/success/wkt/proto/a/a.proto
        testdata/workspace/success/wkt/proto/b/b.proto`,
		"ls-files",
		filepath.Join("testdata", "workspace", "success", "wkt"),
	)
	testRunStdout(
		t,
		nil,
		bufcli.ExitCodeFileAnnotation,
		`testdata/workspace/success/wkt/other/proto/c/c.proto:6:1:Package name "c" should be suffixed with a correctly formed version, such as "c.v1".
        testdata/workspace/success/wkt/proto/a/a.proto:3:1:Package name "a" should be suffixed with a correctly formed version, such as "a.v1".
        testdata/workspace/success/wkt/proto/b/b.proto:3:1:Package name "b" should be suffixed with a correctly formed version, such as "b.v1".`,
		"lint",
		filepath.Join("testdata", "workspace", "success", "wkt"),
	)
	testRunStdout(
		t,
		nil,
		0,
		``,
		"breaking",
		filepath.Join("testdata", "workspace", "success", "wkt"),
		"--against",
		filepath.Join("testdata", "workspace", "success", "wkt"),
	)
}

func TestWorkspaceBreakingFail(t *testing.T) {
	// The two workspaces define a different number of
	// images, so it's impossible to verify compatibility.
	testRunStdout(
		t,
		nil,
		0,
		``,
		"build",
		filepath.Join("testdata", "workspace", "fail", "breaking"),
	)
	testRunStdoutStderr(
		t,
		nil,
		1,
		``,
		`Failed to "breaking": input contained 1 images, whereas against contained 2 images.`,
		"breaking",
		filepath.Join("testdata", "workspace", "fail", "breaking"),
		"--against",
		filepath.Join("testdata", "workspace", "success", "breaking"),
	)
}

func TestWorkspaceDuplicateFail(t *testing.T) {
	// The workspace includes multiple images that define the same file.
	testRunStdoutStderr(
		t,
		nil,
		1,
		``,
		`Failed to "build": foo.proto exists in multiple locations: testdata/workspace/fail/duplicate/other/proto/foo.proto testdata/workspace/fail/duplicate/proto/foo.proto.`,
		"build",
		filepath.Join("testdata", "workspace", "fail", "duplicate"),
	)
}

func TestWorkspaceNotExistFail(t *testing.T) {
	// The directory defined in the workspace does not exist.
	testRunStdoutStderr(
		t,
		nil,
		1,
		``,
		`Failed to "build": module "notexist" listed in testdata/workspace/fail/notexist/buf.work contains no .proto files.`,
		"build",
		filepath.Join("testdata", "workspace", "fail", "notexist"),
	)
}

func TestWorkspaceJumpContextFail(t *testing.T) {
	// The workspace directories cannot jump context.
	testRunStdoutStderr(
		t,
		nil,
		1,
		``,
		`Failed to "build": ../breaking/other/proto: is outside the context directory.`,
		"build",
		filepath.Join("testdata", "workspace", "fail", "jumpcontext"),
	)
}

func TestWorkspaceSymlinkFail(t *testing.T) {
	// The workspace includes a symlink that isn't buildable.
	testRunStdoutStderr(
		t,
		nil,
		bufcli.ExitCodeFileAnnotation,
		``,
		`testdata/workspace/fail/symlink/b/b.proto:5:8:c.proto: does not exist`,
		"build",
		filepath.Join("testdata", "workspace", "fail", "symlink"),
	)
}

func TestWorkspaceAbsoluteFail(t *testing.T) {
	// The buf.work file cannot specify absolute paths.
	testRunStdoutStderr(
		t,
		nil,
		1,
		``,
		`Failed to "build": module "/home/buf" listed in testdata/workspace/fail/absolute/buf.work must be a relative path.`,
		"build",
		filepath.Join("testdata", "workspace", "fail", "absolute"),
	)
}

func TestWorkspaceDirOverlapFail(t *testing.T) {
	// The buf.work file cannot specify overlapping diretories.
	testRunStdoutStderr(
		t,
		nil,
		1,
		``,
		`Failed to "build": module "foo" contains module "foo/bar" in testdata/workspace/fail/diroverlap/buf.work; see https://docs.buf.build/faq for more details.`,
		"build",
		filepath.Join("testdata", "workspace", "fail", "diroverlap"),
	)
}

func TestWorkspaceInputOverlapFail(t *testing.T) {
	// The target input cannot overlap with any of the directories defined
	// in the workspace.
	testRunStdoutStderr(
		t,
		nil,
		1,
		``,
		`Failed to "build": failed to build input "proto/buf" because it is contained by module "proto" listed in testdata/workspace/fail/overlap/buf.work; see https://docs.buf.build/faq for more details.`,
		"build",
		filepath.Join("testdata", "workspace", "fail", "overlap", "proto", "buf"),
	)
	testRunStdoutStderr(
		t,
		nil,
		1,
		``,
		`Failed to "build": failed to build input "other" because it contains module "other/proto" listed in testdata/workspace/success/dir/buf.work; see https://docs.buf.build/faq for more details.`,
		"build",
		filepath.Join("testdata", "workspace", "success", "dir", "other"),
	)
}

func TestWorkspaceRegularFileFail(t *testing.T) {
	// Build directory inputs must be directories.
	testRunStdoutStderr(
		t,
		nil,
		1,
		``,
		`Failed to "build": testdata/workspace/success/dir/proto/rpc.proto: not a directory.`,
		"build",
		filepath.Join("testdata", "workspace", "success", "dir", "proto", "rpc.proto"),
	)
}

func createZipFromDir(t *testing.T, rootPath string, archiveName string) string {
	t.Helper()
	zipDir := filepath.Join(os.TempDir(), rootPath)
	t.Cleanup(
		func() {
			require.NoError(t, os.RemoveAll(zipDir))
		},
	)
	require.NoError(t, os.MkdirAll(zipDir, 0755))

	storageosProvider := storageos.NewProvider(storageos.ProviderWithSymlinks())
	testdataBucket, err := storageosProvider.NewReadWriteBucket(
		rootPath,
		storageos.ReadWriteBucketWithSymlinksIfSupported(),
	)
	require.NoError(t, err)

	buffer := bytes.NewBuffer(nil)
	require.NoError(t, storagearchive.Zip(
		context.Background(),
		testdataBucket,
		buffer,
		true,
	))

	zipBucket, err := storageosProvider.NewReadWriteBucket(
		zipDir,
		storageos.ReadWriteBucketWithSymlinksIfSupported(),
	)
	require.NoError(t, err)

	zipCloser, err := zipBucket.Put(
		context.Background(),
		archiveName,
	)
	require.NoError(t, err)
	t.Cleanup(
		func() {
			require.NoError(t, zipCloser.Close())
		},
	)
	_, err = zipCloser.Write(buffer.Bytes())
	require.NoError(t, err)
	return zipDir
}
