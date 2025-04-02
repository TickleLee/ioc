# Changelog

## v0.0.5

### 🚀 Enhancements

- Implemented Advanced Logging System
  - Integrated Zap-based logging interface
  - Multiple output modes support (console, file)
  - Comprehensive log levels (debug, info, warn, error, fatal)
  - Structured logging fields capability
  - Colored log output in development mode

- Enhanced Container Implementation
  - Added logger component integration
  - Implemented detailed logging in key methods
  - Improved error handling with context information

- Extended Global Functions
  - Added logging configuration API
  - Provided quick debug logging enablement
  - Implemented container logger accessor methods

- Introduced Container Initialization Phases
  - NotInitialized: Container in uninitialized state
  - InjectionPhase: Dependency injection phase
  - PostConstructPhase: PostConstruct method execution phase
  - Initialized: Fully initialized state

- Implemented Two-Phase Initialization Mechanism
  - Phase 1: Create all instances and inject dependencies
  - Phase 2: Execute all PostConstruct methods

- Added Bean Definition Status Flags
  - injected: Marks dependency injection completion status
  - initialized: Marks PostConstruct initialization completion status

- Optimized Bean Initialization Process
  - Split initBean method into three independent methods:
    - createBeanInstance: Responsible for creating bean instances
    - injectBeanDependencies: Handles dependency injection
    - initializeBean: Executes PostConstruct initialization

- Enhanced `GetSafe()` method for thread-safe bean retrieval
  - Supports safe bean retrieval during different initialization phases
  - Enables safe instance return during PostConstruct phase
  - Adds circular dependency detection to prevent initialization issues
  - Supports thread-safe retrieval of beans


## v0.0.4

### 🐛 fix
- Added RegisterTypeWithName() method for global

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