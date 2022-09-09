<a href="https://www.code-intelligence.com/">
<img src="https://www.code-intelligence.com/hubfs/Logos/CI%20Logos/Logo_quer_white.png" alt="Code Intelligence logo" width="450px">
</a>

# cifuzz

> **_IMPORTANT:_** This project is under active development.
Be aware that the behavior of the commands or the configuration
can change.

[![Tests](https://github.com/CodeIntelligenceTesting/cifuzz/actions/workflows/pipeline_pr.yml/badge.svg?branch=main)](https://github.com/CodeIntelligenceTesting/cifuzz/actions/workflows/pipeline_pr.yml)

**cifuzz** is a CLI tool that helps you to integrate and run fuzzing
based tests into your project.

## Getting started
If you are new to the world of fuzzing, we recommend you to take a
look at our [Glossary](docs/Glossary.md).

### Installation

**Prerequisites**
* [CMake >= 3.16](https://cmake.org/)
* [LLVM >= 11](https://clang.llvm.org/get_started.html)

#### Installing required dependencies
**Ubuntu / Debian**
<!-- when changing this, please make sure it is in sync with the E2E pipeline -->
```bash
sudo apt install cmake clang llvm
```

**Arch**
<!-- when changing this, please make sure it is in sync with the E2E pipeline -->
```bash
sudo pacman -S cmake clang llvm
```

**MacOS**
<!-- when changing this, please make sure it is in sync with the E2E pipeline -->
```bash
brew install cmake llvm
```

**Windows**
<!-- when changing this, please make sure it is in sync with the E2E pipeline -->
<!-- clang is included in the llvm package --->
At least Visual Studio 2022 version 17 is required.
```bash
choco install cmake llvm
```

#### Installing cifuzz
You can get the latest release [here](https://github.com/CodeIntelligenceTesting/cifuzz/releases/latest)
or by running our install script:

```bash
sh -c "$(curl -fsSL https://raw.githubusercontent.com/CodeIntelligenceTesting/cifuzz/main/install.sh)"
```

If you are using Windows you can download the [latest release](https://github.com/CodeIntelligenceTesting/cifuzz/releases/latest/download/cifuzz_installer_windows.exe) 
and execute it.

By default, cifuzz gets installed in your home directory under `cifuzz`.
You can customize the installation directory with `./cifuzz_installer -i /target/dir`.

Do not forget to add the installation directory to your `PATH`.

### Setup / Create your first fuzz test

**cifuzz** commands will interactively guide you through the needed
options and show next steps. You can find a complete
list of the available commands with all supported options and
parameters by calling `cifuzz command --help` or
[here](https://github.com/CodeIntelligenceTesting/cifuzz/wiki/cifuzz).

1. To initialize your project with cifuzz just execute `cifuzz init`
in the root directory of your project. This will create a file named
`cifuzz.yaml` containing the needed configuration and print out any
necessary steps to setup your project.

2. The next step is to create a fuzz test. Execute `cifuzz create`
and follow the instructions given by the command. This will create a
stub for your fuzz test, lets say it is called `my_fuzz_test_1.cpp` and
tell you how to integrate it into your project. Usually you also have to
add instructions in your CMakeLists.txt file to link the fuzz test with 
the software under test (e.g. use the `target_link_libraries directive`). 
The `add_fuzz_test` directive can be treated like `add_executable`.

3. Edit `my_fuzz_test_1.cpp` so it actually calls the function you want
to test with the input generated by the fuzzer. To learn more about
writing fuzz tests you can take a look at our
[Tutorial](docs/How-To-Write-A-Fuzz-Test.md) or one of the
[example projects](examples).

4. Start the fuzzing by executing `cifuzz run my_fuzz_test_1`.
**cifuzz** now tries to build the fuzz test and starts a fuzzing run.

### Generate coverage report

Once you executed a fuzz test, you can generate a coverage report which
shows the line by line coverage of the fuzzed code:

    cifuzz coverage my_fuzz_test_1

See [here](docs/Coverage-ide-integrations.md) for instructions on how to
generate and visualize coverage reports right from your IDE.

### Regression testing

**Important:** In general there are two ways to run your fuzz test:

1. An actual fuzzing run by calling: `cifuzz run my_fuzz_test_1`.
The fuzzer will rapidly generate new inputs and feed them into your
fuzz test. Any input that covers new parts of the fuzzed project will
be added to the generated corpus. cifuzz will run until a crash occurs
and report detailed information about the finding.

2. As a regression test, by invoking it through your IDE/editor or by
directly executing the replayer binary
(see [here](docs/How-To-Write-A-Fuzz-Test.md#regression-test--replayer)
on how to build that binary).
This will use the replayer to apply existing input data from the
seed corpus, which has to be stored in the directory
`<fuzz-test-name>_inputs` beside your fuzz test. Note that this
directory has to be created manually. In case a crash was found, the
directory will be created and the crashing input
is added to this directory automatically.
The fuzz test will stop immediately after
applying all inputs or earlier if a regression occurs.


### Sandboxing

On Linux, **cifuzz** runs the fuzz tests in a sandbox by default, to
avoid the fuzz test accidentally harming the system, for example by
deleting files or killing processes. It uses [Minijail](https://google.github.io/minijail/minijail0.1.html) for
that.

If you experience problems when running fuzz tests via **cifuzz** and
you don't expect your fuzz tests to do any harm to the system (or you're
already running **cifuzz** in a container), you might want to disable
the sandbox via the `--use-sandbox=false` flag or the
[`use-sandbox: false` config file setting](docs/Configuration.md#use-sandbox).

## Contributing
Want to help improve cifuzz? Check out our [contributing documentation](CONTRIBUTING.md).
There you will find instructions for building the tool locally.

If you find an issue, please report it on the issue tracker.
