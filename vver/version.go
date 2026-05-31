package vver

import versionimpl "github.com/imajinyun/go-knifer/internal/version"

const DefaultVersionsDelimiter = versionimpl.DefaultVersionsDelimiter

// CompareVersion compares two version strings.
//
// It returns a negative value when version1 is smaller, a positive value when
// version1 is greater, and zero when both versions are equivalent.
func CompareVersion(version1, version2 string) int {
	return versionimpl.CompareVersion(version1, version2)
}

// AnyMatch reports whether currentVersion matches any expression in compareVersions.
func AnyMatch(currentVersion string, compareVersions ...string) bool {
	return versionimpl.AnyMatch(currentVersion, compareVersions...)
}

// AnyMatchSlice reports whether currentVersion matches any expression in compareVersions.
func AnyMatchSlice(currentVersion string, compareVersions []string) bool {
	return versionimpl.AnyMatchSlice(currentVersion, compareVersions)
}

// IsGreaterThan reports whether currentVersion is greater than compareVersion.
func IsGreaterThan(currentVersion, compareVersion string) bool {
	return versionimpl.IsGreaterThan(currentVersion, compareVersion)
}

// IsGreaterThanOrEqual reports whether currentVersion is greater than or equal to compareVersion.
func IsGreaterThanOrEqual(currentVersion, compareVersion string) bool {
	return versionimpl.IsGreaterThanOrEqual(currentVersion, compareVersion)
}

// IsLessThan reports whether currentVersion is less than compareVersion.
func IsLessThan(currentVersion, compareVersion string) bool {
	return versionimpl.IsLessThan(currentVersion, compareVersion)
}

// IsLessThanOrEqual reports whether currentVersion is less than or equal to compareVersion.
func IsLessThanOrEqual(currentVersion, compareVersion string) bool {
	return versionimpl.IsLessThanOrEqual(currentVersion, compareVersion)
}

// MatchEl reports whether currentVersion satisfies a semicolon-separated version expression.
func MatchEl(currentVersion, versionEl string) bool {
	return versionimpl.MatchEl(currentVersion, versionEl)
}

// MatchElWithDelimiter reports whether currentVersion satisfies versionEl using versionsDelimiter.
func MatchElWithDelimiter(currentVersion, versionEl, versionsDelimiter string) bool {
	return versionimpl.MatchElWithDelimiter(currentVersion, versionEl, versionsDelimiter)
}

// MatchElWithDelimiterErr validates the delimiter and reports expression matching errors.
func MatchElWithDelimiterErr(currentVersion, versionEl, versionsDelimiter string) error {
	return versionimpl.MatchElWithDelimiterErr(currentVersion, versionEl, versionsDelimiter)
}

// MatchElByDelimiter is a bool-returning convenience wrapper around MatchElWithDelimiter.
func MatchElByDelimiter(currentVersion, versionEl, versionsDelimiter string) bool {
	return versionimpl.MatchElByDelimiter(currentVersion, versionEl, versionsDelimiter)
}
