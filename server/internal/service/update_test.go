package service

import (
	"bytes"
	"context"
	"ginnexttemplate/internal/pkg/common"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

type serverUpdateRoundTripFunc func(req *http.Request) (*http.Response, error)

func (f serverUpdateRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func resetServerUpgradeTestState(t *testing.T) {
	t.Helper()
	serverUpgradeState.Lock()
	serverUpgradeState.inProgress = false
	serverUpgradeState.status = ""
	serverUpgradeState.logs = nil
	serverUpgradeState.Unlock()
	manualServerBinaryState.Lock()
	cleanupManualServerBinaryCandidateLocked()
	manualServerBinaryState.Unlock()
}

func fakeServerBinaryFixture(version string) (string, []byte) {
	if runtime.GOOS == "windows" {
		return "ginnexttemplate-server-test.cmd", []byte("@echo off\r\necho " + version + "\r\n")
	}
	return "ginnexttemplate-server-test.sh", []byte("#!/bin/sh\necho " + version + "\n")
}

func TestIsVersionNewer(t *testing.T) {
	testCases := []struct {
		name     string
		current  string
		latest   string
		expected bool
	}{
		{name: "newer patch", current: "v1.2.3", latest: "v1.2.4", expected: true},
		{name: "same version", current: "v1.2.3", latest: "v1.2.3", expected: false},
		{name: "older remote", current: "v1.3.0", latest: "v1.2.9", expected: false},
		{name: "double digit segment", current: "v1.9.9", latest: "v1.10.0", expected: true},
		{name: "stable newer than prerelease", current: "v1.2.3-rc.1", latest: "v1.2.3", expected: true},
		{name: "prerelease not newer than same stable", current: "v1.2.3", latest: "v1.2.3-rc.1", expected: false},
		{name: "newer prerelease sequence", current: "v1.2.3-rc.1", latest: "v1.2.3-rc.2", expected: true},
		{name: "git describe newer than same tag", current: "v0.6.3", latest: "v0.6.3-2-gf4d36be", expected: true},
		{name: "git describe distance compares numerically", current: "v0.6.3-2-gf4d36be", latest: "v0.6.3-5-gabc1234", expected: true},
		{name: "dev build", current: "dev", latest: "v0.4.0", expected: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actual := isVersionNewer(testCase.current, testCase.latest)
			if actual != testCase.expected {
				t.Fatalf("unexpected compare result: current=%s latest=%s actual=%v expected=%v", testCase.current, testCase.latest, actual, testCase.expected)
			}
		})
	}
}

func TestBuildLatestServerReleaseView(t *testing.T) {
	originalVersion := common.Version
	common.Version = "v0.4.0"
	t.Cleanup(func() {
		common.Version = originalVersion
		serverUpgradeState.Lock()
		serverUpgradeState.inProgress = false
		serverUpgradeState.Unlock()
	})

	serverUpgradeState.Lock()
	serverUpgradeState.inProgress = true
	serverUpgradeState.Unlock()

	view := buildLatestServerReleaseView(&githubReleaseResponse{
		TagName:     "v0.5.0",
		Body:        "release notes",
		HTMLURL:     "https://github.com/Rain-kl/GinNextTemplate/releases/tag/v0.5.0",
		PublishedAt: "2026-03-11T00:00:00Z",
	}, ReleaseChannelStable)

	if view.CurrentVersion != "v0.4.0" {
		t.Fatalf("unexpected current version: %s", view.CurrentVersion)
	}
	if !view.HasUpdate {
		t.Fatal("expected has_update to be true")
	}
	if !view.InProgress {
		t.Fatal("expected in_progress to reflect upgrade state")
	}
	if view.TagName != "v0.5.0" {
		t.Fatalf("unexpected tag name: %s", view.TagName)
	}
	if view.Channel != ReleaseChannelStable.String() {
		t.Fatalf("unexpected channel: %s", view.Channel)
	}
}

func TestBuildLatestServerReleaseViewDevBuild(t *testing.T) {
	originalVersion := common.Version
	common.Version = "dev"
	t.Cleanup(func() {
		common.Version = originalVersion
		serverUpgradeState.Lock()
		serverUpgradeState.inProgress = false
		serverUpgradeState.Unlock()
	})

	view := buildLatestServerReleaseView(&githubReleaseResponse{
		TagName: "v0.5.0",
	}, ReleaseChannelStable)

	if view.HasUpdate {
		t.Fatal("expected dev build not to report update availability")
	}
	if view.UpgradeSupported {
		t.Fatal("expected dev build not to support self-upgrade")
	}
}

func TestBuildLatestServerReleaseViewPreview(t *testing.T) {
	originalVersion := common.Version
	common.Version = "v0.5.0-rc.1"
	t.Cleanup(func() {
		common.Version = originalVersion
		resetServerUpgradeTestState(t)
	})

	view := buildLatestServerReleaseView(&githubReleaseResponse{
		TagName:     "v0.5.0-rc.2",
		Prerelease:  true,
		PublishedAt: "2026-03-12T00:00:00Z",
	}, ReleaseChannelPreview)

	if !view.HasUpdate {
		t.Fatal("expected preview release to be newer")
	}
	if !view.Prerelease {
		t.Fatal("expected preview flag to be true")
	}
	if view.Channel != ReleaseChannelPreview.String() {
		t.Fatalf("unexpected channel: %s", view.Channel)
	}
}

// TestBuildLatestServerReleaseViewPreviewBypassVersionCheck verifies that switching to
// the preview channel always reports has_update=true, even when the preview tag uses a
// "major.minor.patch-git-<commit>" scheme that would otherwise compare as equal-or-older
// than the currently running stable version.
func TestBuildLatestServerReleaseViewPreviewBypassVersionCheck(t *testing.T) {
	originalVersion := common.Version
	common.Version = "v1.0.0"
	t.Cleanup(func() {
		common.Version = originalVersion
		resetServerUpgradeTestState(t)
	})

	// A typical preview tag: same base version as stable but with a git-commit suffix.
	// Without the bypass, isVersionNewer("v1.0.0", "v1.0.0-git-abc1234") returns false
	// because a version without a prerelease identifier is considered higher than one
	// with a prerelease identifier under semver rules.
	view := buildLatestServerReleaseView(&githubReleaseResponse{
		TagName:     "v1.0.0-git-abc1234",
		Prerelease:  true,
		PublishedAt: "2026-03-12T00:00:00Z",
	}, ReleaseChannelPreview)

	if !view.HasUpdate {
		t.Fatal("expected preview channel to bypass version comparison and report has_update=true")
	}
	if view.Channel != ReleaseChannelPreview.String() {
		t.Fatalf("unexpected channel: %s", view.Channel)
	}
}

func TestUploadManualServerBinary(t *testing.T) {
	originalVersion := common.Version
	common.Version = "v0.4.0"
	t.Cleanup(func() {
		common.Version = originalVersion
		resetServerUpgradeTestState(t)
	})

	fileName, content := fakeServerBinaryFixture("v0.5.0")
	info, err := UploadManualServerBinary(context.Background(), fileName, bytes.NewReader(content))
	if err != nil {
		t.Fatalf("expected upload to succeed: %v", err)
	}
	if !info.ReadyToUpgrade {
		t.Fatal("expected uploaded binary to be ready for upgrade")
	}
	if info.UploadToken == "" {
		t.Fatal("expected upload token to be returned")
	}
	if info.DetectedVersion != "v0.5.0" {
		t.Fatalf("unexpected detected version: %s", info.DetectedVersion)
	}

	manualServerBinaryState.Lock()
	candidate := manualServerBinaryState.candidate
	manualServerBinaryState.Unlock()
	if candidate == nil {
		t.Fatal("expected manual upgrade candidate to be stored")
	}
	if _, err := os.Stat(candidate.TempPath); err != nil {
		t.Fatalf("expected temporary binary to exist: %v", err)
	}
	if candidate.UploadToken != info.UploadToken {
		t.Fatalf("unexpected stored upload token: %s", candidate.UploadToken)
	}
	execPath, err := os.Executable()
	if err != nil {
		t.Fatalf("failed to get executable path: %v", err)
	}
	if filepath.Dir(candidate.TempPath) != filepath.Dir(execPath) {
		t.Fatalf("expected temporary binary in executable dir, got %s want %s", filepath.Dir(candidate.TempPath), filepath.Dir(execPath))
	}
}

func TestBuildUploadedServerBinaryViewAcceptsGitDescribeNewerThanTag(t *testing.T) {
	info := buildUploadedServerBinaryView("ginnexttemplate-server-test", "v0.6.3", "v0.6.3-2-gf4d36be", time.Now())
	if !info.HasUpdate || !info.ReadyToUpgrade {
		t.Fatalf("expected git describe binary to be upgradeable: %+v", info)
	}
}

func TestUploadManualServerBinaryRejectsSameVersion(t *testing.T) {
	originalVersion := common.Version
	common.Version = "v0.5.0"
	t.Cleanup(func() {
		common.Version = originalVersion
		resetServerUpgradeTestState(t)
	})

	fileName, content := fakeServerBinaryFixture("v0.5.0")
	info, err := UploadManualServerBinary(context.Background(), fileName, bytes.NewReader(content))
	if err != nil {
		t.Fatalf("expected upload to succeed: %v", err)
	}
	if info.ReadyToUpgrade {
		t.Fatal("expected same-version upload not to be upgradeable")
	}
	if info.UploadToken != "" {
		t.Fatal("expected same-version upload not to issue a token")
	}

	manualServerBinaryState.Lock()
	defer manualServerBinaryState.Unlock()
	if manualServerBinaryState.candidate != nil {
		t.Fatal("expected no pending manual upgrade candidate")
	}
}

func TestConfirmManualServerUpgrade(t *testing.T) {
	originalVersion := common.Version
	originalExecutor := ServerBinaryUpgradeExecutorForTest()
	originalDelay := ServerUpgradeDispatchDelayForTest()
	common.Version = "v0.4.0"
	called := make(chan string, 1)
	SetServerBinaryUpgradeExecutorForTest(func(execPath string, tempPath string) error {
		called <- tempPath
		return nil
	})
	SetServerUpgradeDispatchDelayForTest(0)
	t.Cleanup(func() {
		common.Version = originalVersion
		SetServerBinaryUpgradeExecutorForTest(originalExecutor)
		SetServerUpgradeDispatchDelayForTest(originalDelay)
		resetServerUpgradeTestState(t)
	})

	fileName, content := fakeServerBinaryFixture("v0.5.0")
	info, err := UploadManualServerBinary(context.Background(), fileName, bytes.NewReader(content))
	if err != nil {
		t.Fatalf("expected upload to succeed: %v", err)
	}

	confirmed, err := ConfirmManualServerUpgrade(info.UploadToken)
	if err != nil {
		t.Fatalf("expected confirm to succeed: %v", err)
	}
	if confirmed.UploadToken != info.UploadToken {
		t.Fatalf("unexpected confirmed upload token: %s", confirmed.UploadToken)
	}

	select {
	case tempPath := <-called:
		if tempPath == "" {
			t.Fatal("expected upgrade executor to receive temp path")
		}
	case <-time.After(time.Second):
		t.Fatal("expected manual upgrade executor to be called")
	}
}

func TestBuildLatestServerReleaseViewIncludesUpgradeLogs(t *testing.T) {
	originalVersion := common.Version
	common.Version = "v0.4.0"
	t.Cleanup(func() {
		common.Version = originalVersion
		resetServerUpgradeTestState(t)
	})

	serverUpgradeState.Lock()
	serverUpgradeState.inProgress = true
	serverUpgradeState.status = "running"
	serverUpgradeState.logs = []ServerUpgradeLogRecord{
		{
			Level:     "info",
			Message:   "download started",
			CreatedAt: time.Now(),
		},
	}
	serverUpgradeState.Unlock()

	view := buildLatestServerReleaseView(&githubReleaseResponse{
		TagName: "v0.5.0",
	}, ReleaseChannelStable)

	if view.UpgradeStatus != "running" {
		t.Fatalf("expected upgrade status to be running, got %s", view.UpgradeStatus)
	}
	if len(view.UpgradeLogs) != 1 {
		t.Fatalf("expected one upgrade log, got %d", len(view.UpgradeLogs))
	}
	if view.UpgradeLogs[0].Message != "download started" {
		t.Fatalf("unexpected upgrade log message: %s", view.UpgradeLogs[0].Message)
	}
}

func TestScheduleServerUpgradeUsesDownloadedBinaryValidation(t *testing.T) {
	originalVersion := common.Version
	originalClient := UpdateHTTPClientForTest()
	originalExecutor := ServerBinaryUpgradeExecutorForTest()
	originalDelay := ServerUpgradeDispatchDelayForTest()
	common.Version = "v0.4.0"
	called := make(chan string, 1)

	SetUpdateHTTPClientForTest(&http.Client{
		Transport: serverUpdateRoundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.String() {
			case "https://api.github.com/repos/Rain-kl/GinNextTemplate/releases/latest":
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body: io.NopCloser(strings.NewReader(`{
						"tag_name":"v0.5.0",
						"body":"release notes",
						"html_url":"https://github.com/Rain-kl/GinNextTemplate/releases/tag/v0.5.0",
						"published_at":"2026-03-11T00:00:00Z",
						"assets":[{"name":"ginnexttemplate-server-` + runtime.GOOS + `-` + runtime.GOARCH + `","browser_download_url":"https://downloads.example.com/ginnexttemplate-server"}]
					}`)),
				}, nil
			case "https://downloads.example.com/ginnexttemplate-server":
				_, content := fakeServerBinaryFixture("v0.5.0")
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       io.NopCloser(bytes.NewReader(content)),
				}, nil
			default:
				t.Fatalf("unexpected request url: %s", req.URL.String())
				return nil, nil
			}
		}),
	})
	SetServerBinaryUpgradeExecutorForTest(func(execPath string, tempPath string) error {
		called <- tempPath
		return nil
	})
	SetServerUpgradeDispatchDelayForTest(0)
	t.Cleanup(func() {
		common.Version = originalVersion
		SetUpdateHTTPClientForTest(originalClient)
		SetServerBinaryUpgradeExecutorForTest(originalExecutor)
		SetServerUpgradeDispatchDelayForTest(originalDelay)
		resetServerUpgradeTestState(t)
	})

	release, err := ScheduleServerUpgrade("stable")
	if err != nil {
		t.Fatalf("expected schedule to succeed: %v", err)
	}
	if !release.InProgress {
		t.Fatal("expected release to report in-progress upgrade")
	}

	select {
	case tempPath := <-called:
		if tempPath == "" {
			t.Fatal("expected upgrade executor to receive temp path")
		}
	case <-time.After(time.Second):
		t.Fatal("expected automatic upgrade executor to be called")
	}

	_, status, logs := snapshotServerUpgradeState()
	if status != "succeeded" {
		t.Fatalf("expected succeeded status after executor call, got %s", status)
	}
	if len(logs) == 0 {
		t.Fatal("expected upgrade logs to be recorded")
	}
}
