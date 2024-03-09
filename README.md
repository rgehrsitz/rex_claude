cmd/: Contains the main entry points for the preprocessor and runtime executables.
preprocessor/main.go: The main file for the preprocessor executable.
runtime/main.go: The main file for the runtime executable.
pkg/: Contains the packages that implement the core functionality of REX.
preprocessor/: Package for the preprocessor-related functionality.
parser.go: Implements the parsing of JSON rule definitions.
optimizer.go: Implements the optimization of rules.
generator.go: Generates the intermediate representation of rules.
runtime/: Package for the runtime-related functionality.
engine.go: Implements the main rule evaluation engine.
evaluator.go: Implements the rule evaluation logic.
redis.go: Implements the interaction with Redis for pub/sub and caching.
rules/: Package for defining the rule-related structs and interfaces.
rule.go: Defines the structs for representing rules.
condition.go: Defines the structs for representing rule conditions.
internal/: Contains the internal packages used by REX.
representation/: Package for the intermediate representation of rules.
intermediate.go: Defines the structs and methods for the intermediate representation.
test/: Contains the test files for the preprocessor and runtime packages.
preprocessor/preprocessor_test.go: Contains the unit tests for the preprocessor package.
runtime/runtime_test.go: Contains the unit tests for the runtime package.
config/: Contains the configuration-related files.
config.go: Defines the configuration structs and methods for REX.
go.mod: The Go module file that defines the module path and dependencies.
go.sum: The checksum file for the module dependencies.
README.md: The README file for the project.
This project structure organizes the code into logical packages and separates the concerns of the preprocessor and runtime. The cmd/ directory contains the main entry points for the executables, while the pkg/ directory contains the core functionality packages.

The internal/ directory is used for internal packages that are not meant to be imported by external packages. The test/ directory contains the test files for the packages.

The config/ directory is used for configuration-related files, and the go.mod and go.sum files are standard Go module files.
