# Coding Tools

A collection of coding utilities, scripts, and developer tools I've built to make development faster and more convenient.

## Repository Structure

The `main` branch intentionally contains only this README.

Each tool is maintained in its own dedicated branch, keeping projects isolated while sharing a single repository.

```
main
├── README.md
├── tool-branch-1
├── tool-branch-2
├── tool-branch-3
└── ...
```

## Available Tools

Each branch contains a standalone project with its own documentation and usage instructions.

| Branch | Description |
|---------|-------------|
| *(coming soon)* | |

## How to Use

List all branches:

```bash
git branch -a
```

Clone the repository:

```bash
git clone https://github.com/<your-username>/coding-tools.git
cd coding-tools
```

Switch to the tool you want:

```bash
git switch <branch-name>
```

Or, if the branch doesn't exist locally yet:

```bash
git switch -c <branch-name> origin/<branch-name>
```

## Why Branches?

Instead of creating a separate repository for every small utility, each tool lives in its own branch. This keeps related projects together while allowing each one to evolve independently.

## Contributing

Suggestions, bug reports, and improvements are always welcome. Feel free to open an issue or submit a pull request.

## License

See the repository license for details.
