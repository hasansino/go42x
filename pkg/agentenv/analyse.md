# Project Architecture Analysis

Analyze this project and extract STABLE, HIGH-LEVEL information that rarely changes. Focus on architectural decisions, design philosophy, and conceptual understanding rather than implementation details.

## Format:

- Focus on WHY decisions were made, not WHAT currently exists
- Describe concepts and philosophy, not current implementation
- Explain the mental model, not the code structure
- Wrap your analysis in tags: `### BEGIN ANALYSIS ###` and `### END ANALYSIS ###` (without the backticks)

## Example output:

```Markdown

## Project Architecture Analysis

This section contains insights and recommendations generated from an automated analysis of the project.

### Project Mission

go42x serves as a comprehensive developer productivity toolkit designed to enhance the software development experience through intelligent automation and AI-assisted workflows. The project addresses the friction points in modern development workflows by providing a suite of tools that integrate AI capabilities directly into the development process. Its core value proposition is reducing cognitive load and repetitive tasks while maintaining developer control and code quality standards.

### Architectural Philosophy

The architecture follows a modular, service-oriented design philosophy where distinct functional domains are encapsulated as independent services. This approach prioritizes composability over monolithic integration, allowing developers to use only the features they need. The design emphasizes clear separation between core business logic (services), user interaction layers (UI/CLI), and infrastructure concerns (providers/adapters). The architecture deliberately avoids tight coupling between components, enabling parallel development and independent testing of features.

... and so on for each section ...

```

## Sections to Include

1. **Project Mission**
    - Core purpose and vision
    - Problem domain and business value
    - Target audience and use cases
    - NOT: Current features or specific implementations

2. **Architectural Philosophy**
    - Overall design approach (monolithic, microservices, modular, etc.)
    - Key architectural patterns and why they were chosen
    - Separation of concerns strategy
    - NOT: Specific file structures or current module names

3. **Domain Model**
    - Core domain concepts and their relationships
    - Bounded contexts and domain boundaries
    - Business rules and invariants
    - NOT: Current database schemas or API endpoints

4. **Design Principles**
    - Code organization philosophy
    - Abstraction strategies
    - Extension and customization approach
    - NOT: Current function names or class hierarchies

5. **Technical Decisions**
    - Why specific technology choices were made
    - Trade-offs that were considered
    - Constraints and limitations accepted
    - NOT: Version numbers or specific dependencies

6. **Quality Philosophy**
    - Testing strategy and philosophy
    - Error handling approach
    - Performance considerations
    - NOT: Current test coverage percentages or metrics

7. **Evolution Strategy**
    - How the project is designed to grow
    - Extension points and plugin architecture
    - Backward compatibility approach
    - NOT: Current roadmap or pending features

8. **Folder Structure Philosophy**
    - Rationale behind the organization of code into folders
    - Document every case with examples
    - For every example draw tree diagrams

## Exclude From Analysis

- Current file paths and function names
- Specific dependencies and versions
- Implementation details and code snippets
- Current metrics and statistics
- Build commands and configurations
- Recent commits and changes
- Specific bug fixes and features
