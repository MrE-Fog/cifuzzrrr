package build

import (
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/pkg/errors"

	"code-intelligence.com/cifuzz/pkg/log"
	"code-intelligence.com/cifuzz/util/envutil"
)

var JazzerFuzzTestRegex = regexp.MustCompile(`@FuzzTest|\sfuzzerTestOneInput\(`)

type Result struct {
	// A name which uniquely identifies the fuzz test and is a valid path
	Name string
	// Canonical path of the fuzz test executable
	Executable string
	// Canonical path of the fuzz test's generated corpus directory
	GeneratedCorpus string
	// Canonical path of the fuzz test's default seed corpus directory
	SeedCorpus string
	// Canonical path of the build directory
	BuildDir string
	// The engine for which the fuzz test was built
	Engine string
	// The sanitizers with which the fuzz test was built
	Sanitizers []string
	// The canonical paths of the fuzz test's runtime dependencies
	RuntimeDeps []string
	// Canonical path of the directory to which source file paths should
	// be made relative
	ProjectDir string
}

func CommonBuildEnv() ([]string, error) {
	var err error
	env := os.Environ()

	// Set CIFUZZ=1 to allow the build system to figure out that it was
	// started by cifuzz.
	env, err = envutil.Setenv(env, "CIFUZZ", "1")
	if err != nil {
		return nil, err
	}

	// On Windows, our preferred compiler is MSVC, which can't easily be run
	// from an arbitrary terminal as it requires about a dozen environment
	// variables to be set correctly. Thus, we assume users to run cifuzz from
	// a developer command prompt anyway and thus don't need to set the
	// compiler explicitly.
	if runtime.GOOS != "windows" {
		// Set the C/C++ compiler to clang/clang++ (if not already set),
		// which is needed to build a  binary with fuzzing instrumentation
		// gcc doesn't have -fsanitize=fuzzer.
		if val, ok := envutil.LookupEnv(env, "CC"); !ok || !strings.HasPrefix(path.Base(val), "clang") {
			env, err = envutil.Setenv(env, "CC", "clang")
			if err != nil {
				return nil, err
			}
		}
		if val, ok := envutil.LookupEnv(env, "CXX"); !ok || !strings.HasPrefix(path.Base(val), "clang++") {
			env, err = envutil.Setenv(env, "CXX", "clang++")
			if err != nil {
				return nil, err
			}
		}
	}

	// We don't want to fail if ASan is set up incorrectly for tools
	// built and executed during the build or they contain leaks.
	env, err = envutil.Setenv(env, "ASAN_OPTIONS", "detect_leaks=0:verify_asan_link_order=0")
	if err != nil {
		return nil, err
	}

	return env, nil
}

var commonCFlags = []string{
	// Keep debug symbols
	"-g",
	// Do optimizations which don't harm debugging
	"-Og",
	// To get good stack frames for better debugging
	"-fno-omit-frame-pointer",
	// Conventional macro to conditionally compile out fuzzer road blocks
	// See https://llvm.org/docs/LibFuzzer.html#fuzzer-friendly-build-mode
	"-DFUZZING_BUILD_MODE_UNSAFE_FOR_PRODUCTION",
	// Ensure that asserts are enabled regardless of compilation mode (e.g. explicit -DNDEBUG).
	"-UNDEBUG",
}

func LibFuzzerCFlags() []string {
	// These flags must not contain spaces, because the environment
	// variables that are set to these flags are space separated.
	// Note: Keep in sync with tools/cmake/CIFuzz/share/CIFuzz/CIFuzzFunctions.cmake
	return append(commonCFlags, []string{
		// ----- Flags used to build with libFuzzer -----
		// Compile with edge coverage and compare instrumentation. We
		// use fuzzer-no-link here instead of -fsanitize=fuzzer because
		// CFLAGS are often also passed to the linker, which would cause
		// errors if the build includes tools which have a main function.
		"-fsanitize=fuzzer-no-link",

		// ----- Flags used to build with ASan -----
		// Build with instrumentation for ASan and UBSan and link in
		// their runtime
		"-fsanitize=address,undefined",
		// To support recovering from ASan findings
		"-fsanitize-recover=address",
		// Use additional error detectors for use-after-scope bugs
		// TODO: Evaluate the slow down caused by this flag
		// TODO: Check if there are other additional error detectors
		//       which we want to use
		"-fsanitize-address-use-after-scope",
		// Disable source fortification, which is currently not supported
		// in combination with ASan, see https://github.com/google/sanitizers/issues/247
		"-U_FORTIFY_SOURCE",
	}...)
}

func CoverageCFlags(clangVersion *semver.Version) []string {
	cflags := append(commonCFlags, []string{
		// ----- Flags used to build with code coverage -----
		"-fprofile-instr-generate",
		"-fcoverage-mapping",
		// Disable source fortification to ensure that coverage builds
		// reach all code reached by ASan builds.
		"-U_FORTIFY_SOURCE",
	}...)

	if runtime.GOOS != "darwin" && clangVersion != nil {
		// LLVM's continuous mode requires compile-time support on non-macOS
		// platforms. This support is unstable in Clang 13 and lower, so we
		// only enable it on 14+.
		if clangVersion.Compare(semver.MustParse("14.0.0")) >= 0 {
			cflags = append(cflags, "-mllvm", "-runtime-counter-relocation")
		}
	}
	return cflags
}

func JazzerSeedCorpus(targetClass string, projectDir string) string {
	seedCorpus := targetClass + "Inputs"
	path := strings.Split(seedCorpus, ".")
	path = append([]string{"src", "test", "resources"}, path...)

	return filepath.Join(projectDir, filepath.Join(path...))
}

func JazzerGeneratedCorpus(targetClass string, projectDir string) string {
	return filepath.Join(projectDir, ".cifuzz-corpus", targetClass)
}

func ListJazzerFuzzTests(projectDir string) ([]string, error) {
	testDir := filepath.Join(projectDir, "src", "test", "java")

	var fuzzTests []string
	err := filepath.WalkDir(testDir, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			if filepath.Ext(path) != ".java" {
				return nil
			}

			bytes, err := os.ReadFile(path)
			if err != nil {
				return errors.WithStack(err)
			}

			match := JazzerFuzzTestRegex.MatchString(string(bytes))
			if match == true {
				classPath, err := filepath.Rel(testDir, path)
				if err != nil {
					return errors.WithStack(err)
				}

				classPath = filepath.Join(
					filepath.Dir(classPath),
					strings.TrimSuffix(filepath.Base(path), ".java"),
				)
				classPath = strings.ReplaceAll(classPath, string(os.PathSeparator), ".")

				fuzzTests = append(fuzzTests, classPath)
			}
		}

		return nil
	})
	if err != nil {
		log.Error(err, err.Error())
		return nil, err
	}

	return fuzzTests, nil
}
