# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## 2.3.0 - 2025-08-08
### Added
- Dotenv support

## 2.2.0 - 2025-01-30
### Added
- Add UsageVal to replace Usage, and that allows receiving the value from the `--help` flag

### Changed
- Format the output of the default usage handler in nice columns
- Allow the default usage handler to omit the CLI or ENV column based on the value of the `--help` flag

## 2.1.0 - 2022-09-12
### Changed
- `Processor.Process` will now return an error when a non pointer type is used as an injection source 

## 2.0.1 - 2022-05-12
### Fixed
- make sure to wrap errors when using fmt.Errorf
- error returned by hooks such as Init can be unwrapped

## 2.0.0 - 2022-04-29
### Added
- new Init interface which takes a context as parameter

## 1.4.1 - 2021-12-15
### Fixed
- Avoid ranging over recursive dependencies

## 1.4.0 - 2021-02-22
### Changed
- Separate required and optional params in default usage (`--help`)

## 1.3.0 - 2020-04-16
### Added
- Allow parsing slice of int and int64 values.

## 1.2.1 - 2019-11-18
### Fixed
- Export the DefaultProcessor and DefaultProvider for convenience.

## 1.2.0 - 2019-11-12
### Added
- Usage message is now overridable in the Processor struct.

## 1.1.2 - 2019-10-01
### Added
- The provider information is now used

## 1.1.1 - 2019-06-25
### Changed
- Remove the internal state of the `EnvProvider`

## 1.1.0 - 2019-06-25
### Added
- Full module compliance

### Changed
- Modify the `Parser` and `Provider` signatures to allow retrieving any value,
  not only strings

### Removed
- Dependency to the `github.com/pkg/errors` package
- Dependency to the `github.com/fatih/structtag` package

### Fixed
- README typos

## 1.0.0 - 2018-12-26
### Added
- Initial release of the project
