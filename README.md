# Not Used Key Checker

This is a simple script that checks for unused keys in a JSON file. It is useful when you have a large JSON file and you want to know which keys are not being used.

## Usage

```bash
go run main.go -json=<json_file_path> -dir=<project_directory> -ext=<extensions>
```

- `json_file_path`: The path to the JSON file.
- `project_directory`: The path to the project directory.
- `ext`: The file extensions to search for. Default is `.js,.ts,.tsx`.

## Example

```bash
go run main.go -json=example.json -dir=../example -ext=.go
```


