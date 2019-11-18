# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
