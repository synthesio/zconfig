# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## 3.0.0 - 2022-09-12
### Changed
- Handle injection via a dedicated `Inject` hook executed just after `Repository.Hook`
  to have proper injection of non-pointer fields.
  This is considered a breaking change because injection was previously 
  handled by the call to `Processor.Process` independently of any registered Hooks.
  **Important:** Users who do not call the `Configure` and do not use the `DefaultProcessor` must
  add the `Inject` hook to their `Processor`, otherwise injection will stop working.

## 2.0.1 - 2022-05-12
### Fixed
- Make sure to wrap errors when using fmt.Errorf
- Error returned by hooks such as Init can be unwrapped

## 2.0.0 - 2022-04-29
### Added
- New Init interface which takes a context as parameter

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
