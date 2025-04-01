# Changelog

## v0.0.3

### 🚀 Enhancements

- Added `RegisterTypeWithName()` method to register beans with custom type name and bean name
  - Supports developers to specify both type name and bean name simultaneously, providing a more flexible registration method
  - Facilitates precise retrieval through the combination of type and name
  - Addresses the management of multiple beans of the same type

## v0.0.2

### 🚀 Enhancements

- Added `GetAll()` method to retrieve all registered beans
- Added `GetAllNames()` method to retrieve all registered bean names
- Added comprehensive test cases in `container_test.go`
- Improved naming conventions in test cases
- Updated documentation with new method descriptions
- Added link to detailed injection documentation (Inject.md)

## v0.0.1 (2025-03-30)

### 🎉 First Release

- Initial release of the IoC library
- Core features implemented:
  - Singleton and Prototype scope support
  - Dependency injection via struct tags
  - Type-based dependency injection
  - Factory method for object creation
  - Automatic type matching for dependencies
  - Initialization method support (PostConstruct)
  - Manual injection support
- Basic documentation in English and Chinese
- Example code for demonstration